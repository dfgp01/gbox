package ecs

import (
	"reflect"
	"strconv"
)

// componentRegistry 组件类型注册表：维护 reflect.Type <-> ComponentID 的映射。
// 反射逻辑集中在本文件（archetype.go），其他文件不直接使用 reflect。
type componentRegistry struct {
	byType map[reflect.Type]ComponentID // 反射类型 -> 组件 ID
	byID   map[ComponentID]reflect.Type // 组件 ID -> 反射类型
}

// newComponentRegistry 创建一个空注册表。
func newComponentRegistry() *componentRegistry {
	return &componentRegistry{
		byType: make(map[reflect.Type]ComponentID),
		byID:   make(map[ComponentID]reflect.Type),
	}
}

// register 以实例 v 识别组件类型并注册，返回其组件 ID（幂等）。
// v 必须是值类型（不允许指针/nil）；数量上限由调用方负责。
func (r *componentRegistry) register(v any) ComponentID {
	t := reflect.TypeOf(v)
	if t == nil || t.Kind() == reflect.Pointer {
		panic("ecs: 组件必须是值类型，不允许传指针或 nil")
	}
	if id, ok := r.byType[t]; ok {
		return id
	}
	id := ComponentID(len(r.byID)) + 1 // 从 1 开始分配，0 保留为无效哨兵
	r.byType[t] = id
	r.byID[id] = t
	return id
}

// idOf 返回实例 v 对应类型的组件 ID；未注册返回 (0, false)。
func (r *componentRegistry) idOf(v any) (ComponentID, bool) {
	t := reflect.TypeOf(v)
	if t == nil {
		return 0, false
	}
	id, ok := r.byType[t]
	return id, ok
}

// typOf 返回组件 ID 对应的类型（供 Archetype 建列使用）。
func (r *componentRegistry) typOf(id ComponentID) reflect.Type {
	return r.byID[id]
}

// Column 列式存储中的一列：存储“同一组件类型的所有实例”。
//
// 与“参考设计”中 data []any 不同，这里用类型化切片 []T（reflect.Value 持有），
// 好处是：
//   - 每列元素零装箱/拆箱、按值紧凑存储、缓存友好；
//   - 元素可寻址（addressable），对外可以安全返回 *T 指针供直接修改，
//     例如 pos.X += 1 会真实写回存储；
//   - 列类型在创建时确定，写入/读取由列自身保证类型一致。
type Column struct {
	typ  reflect.Type  // 元素类型 T
	data reflect.Value // 类型化切片 []T
}

// ptr 返回行 row 的组件指针（any 中持有 *T，可原位修改）。
func (c *Column) ptr(row int) any {
	return c.data.Index(row).Addr().Interface()
}

// set 将值 v 写入行 row。
func (c *Column) set(row int, v any) {
	c.data.Index(row).Set(reflect.ValueOf(v))
}

// swapRows 交换两行的数据（swap-remove 删除时用）。
func (c *Column) swapRows(a, b int) {
	av := c.data.Index(a)
	bv := c.data.Index(b)
	tmp := reflect.New(c.typ).Elem()
	tmp.Set(av)
	av.Set(bv)
	bv.Set(tmp)
}

// clearRow 将行 row 置零（删除后清空，避免残留引用）。
func (c *Column) clearRow(row int) {
	c.data.Index(row).Set(reflect.Zero(c.typ))
}

// Archetype 是“组件组合完全相同的一批实体”的存储单元（列式存储）。
//
//   - mask 精确描述该原型的组件组合；同一个精确掩码只会对应一个 Archetype
//     （由 World.archetypes 保证，按需创建、无需预定义任何组合结构体）；
//   - entities 与各列按“行号”对齐：第 i 行就是第 i 个实体的组件集合；
//   - 实体在 Archetype 内使用 swap-remove 方式删除（O(1)）；
//   - 增删组件本质是“实体从一个 Archetype 迁移到另一个 Archetype”。
type Archetype struct {
	mask     ComponentMask    // 该原型包含的组件组合（身份）
	compIDs  []ComponentID    // 列对应的组件 ID（columns[i] 的元素类型即组件 i）
	columns  []Column         // 列式存储
	entities []EntityID       // 与各列行号对齐的实体 ID（dense）
	row      map[EntityID]int // 实体 ID -> 行号（swap-remove 时同步维护）
	n        int              // 有效行数（末尾为删除后残留的槽位）
}

// newArchetype 以掩码创建 Archetype（按位展开得到列定义）。
func newArchetype(mask ComponentMask, w *world) *Archetype {
	a := &Archetype{
		mask: mask,
		row:  make(map[EntityID]int),
	}
	// 逐位展开：第 bit 位为 1 ⇒ 组件 ID = bit+1
	for bit := ComponentID(0); bit < maxComponents; bit++ {
		if mask&(1<<bit) != 0 {
			id := bit + 1
			a.compIDs = append(a.compIDs, id)
			a.columns = append(a.columns, Column{typ: w.reg.typOf(id)})
		}
	}
	return a
}

// hasType 是否包含指定组件。
func (a *Archetype) hasType(id ComponentID) bool {
	_, ok := a.columnOf(id)
	return ok
}

// columnOf 返回组件 ID 对应的列索引。
func (a *Archetype) columnOf(id ComponentID) (int, bool) {
	for i, cid := range a.compIDs {
		if cid == id {
			return i, true
		}
	}
	return 0, false
}

// addEntity 追加实体，返回其行号。
func (a *Archetype) addEntity(e EntityID) int {
	row := a.n
	a.n++
	if len(a.entities) < a.n {
		// 扩容各列（首次 MakeSlice，之后 Append）
		a.entities = append(a.entities, e)
		for i := range a.columns {
			col := &a.columns[i]
			if !col.data.IsValid() {
				col.data = reflect.MakeSlice(reflect.SliceOf(col.typ), 1, 4)
			} else {
				col.data = reflect.Append(col.data, reflect.Zero(col.typ))
			}
		}
	} else {
		// 复用被删除后残留的槽位（removeEntity 已清空其内容）
		a.entities[row] = e
	}
	a.row[e] = row
	return row
}

// removeEntity 移除实体（swap-remove，O(1)）。
func (a *Archetype) removeEntity(e EntityID) {
	row, ok := a.row[e]
	if !ok {
		return
	}
	last := a.n - 1
	if row != last {
		// 把最后一行换到被删行
		le := a.entities[last]
		a.entities[row] = le
		a.row[le] = row
		for i := range a.columns {
			a.columns[i].swapRows(row, last)
		}
	}
	// 清空尾部槽位，避免引用悬挂
	for i := range a.columns {
		a.columns[i].clearRow(last)
	}
	delete(a.row, e)
	a.n--
}

// getPtr 返回实体行 row 上组件 id 的指针（any 中持有 *T）。
func (a *Archetype) getPtr(id ComponentID, row int) any {
	ci, ok := a.columnOf(id)
	if !ok {
		panic("ecs: 该 Archetype 不包含组件 id=" + strconv.Itoa(int(id)))
	}
	return a.columns[ci].ptr(row)
}
