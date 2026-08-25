package ecs

// EntityID 实体 ID，仅是一个唯一标识，本身不携带任何数据。
// 组件与实体的关系由 World / Archetype 维护，实体只是“组件集合”的身份标识。
type EntityID uint32

// InvalidEntity 无效实体（哨兵值），用于表示“无实体”，Spawn 不会分配该值。
const InvalidEntity EntityID = 0

// World 世界接口：负责管理所有实体与组件数据（内部为 Archetype 列式存储）。
// 按功能分组接口：
// 0) 初始化操作：RegisterComponent
// 1) 实体操作：Spawn / Despawn / Alive / Count
// 2) 实体单个组件操作：Get / Add / Remove
// 3) 批量查询操作：Query
// 4) 检查和debug：Alive / Count / Entities / BlockCount
type World interface {
	// RegisterComponent 以实例 v 注册组件类型，返回其组件 ID（幂等）。具体注册策略（如数量上限）由实现决定。
	RegisterComponent(v any) ComponentID
	// Spawn 创建实体并附加组件值（任意组合、任意顺序）。
	Spawn(comps ...any) EntityID
	// Despawn 销毁实体。
	Despawn(e EntityID)
	// Get 读取实体组件 id 的指针（any 中持有 *T，可原位修改）；不存在时返回 nil。
	Get(e EntityID, id ComponentID) any
	// Add 为实体添加组件 v，返回其组件 ID；组件类型未注册则自动注册。
	// 实体没有该组件则创建并写入 v，已有则不覆盖；实体不存在则忽略。不 panic。
	Add(e EntityID, v any) ComponentID
	// Remove 移除实体的组件 id；实体不存在或没有该组件时为空操作。
	Remove(e EntityID, id ComponentID)
	// Query 超集匹配查询：用回调遍历“至少包含 mask 中全部组件”的实体；
	// fn 收到当前实体 e 与查询迭代器 q（回调内可用 q.Get(id) 取组件），
	// 返回 false 则中断遍历。
	Query(mask ComponentMask, fn func(e EntityID, q *Query) bool)
	// Alive 实体是否存活。
	Alive(e EntityID) bool
	// Count 存活实体数量。
	Count() int
	// Entities 返回当前存活实体列表（用于观察/调试）
	Entities() []EntityID
	// BlockCount 返回当前存储块数量（如Archetype,用于观察/调试）
	BlockCount() int
}

// world 是 World 接口的默认实现。内部采用Archetype存储
type world struct {
	nextEntityID EntityID
	reg          *componentRegistry           // 组件类型注册表（反射集中在 archetype.go）
	archetypes   map[ComponentMask]*Archetype // 精确掩码 -> Archetype（按需创建）
	entities     map[EntityID]*Archetype      // 实体 -> 所在 Archetype（兼作存活表）
	// 超集匹配缓存（先不做）
	// queryCache map[ComponentMask]*Query
}

// NewWorld 创建一个空 World。
func NewWorld() World {
	return &world{
		nextEntityID: InvalidEntity,
		reg:          newComponentRegistry(),
		archetypes:   make(map[ComponentMask]*Archetype),
		entities:     make(map[EntityID]*Archetype),
	}
}

// ---- 0) 初始化：组件类型注册 ----

// RegisterComponent 以实例 v 注册组件类型，返回其组件 ID（幂等）。
// 数量上限由调用方负责：超过 64 后不再强制 panic，掩码位运算由调用方自行保证。
func (w *world) RegisterComponent(v any) ComponentID {
	return w.reg.register(v)
}

// ---- 1) 实体操作 ----

// Spawn 创建实体并附加给定组件值（任意组合、任意顺序）。
// 组件必须是已注册的值类型（不允许指针/nil），且每种类型最多出现一次。
// 内部会按掩码按需创建/复用 Archetype，无需预定义任何组合结构体。
func (w *world) Spawn(comps ...any) EntityID {
	// 解析组件 ID 并校验
	ids := make([]ComponentID, 0, len(comps))
	seen := ComponentMask(0)
	for _, c := range comps {
		id, ok := w.reg.idOf(c)
		if !ok {
			panic("ecs: Spawn 组件类型未注册，请先 RegisterComponent")
		}
		if seen&maskOf(id) != 0 {
			panic("ecs: Spawn 不能包含重复的组件类型")
		}
		seen |= maskOf(id)
		ids = append(ids, id)
	}

	// 生成掩码并获取/创建 Archetype
	arch := w.getOrCreateArchetype(Mask(ids...))
	e := w.allocEntity()
	row := arch.addEntity(e)
	w.entities[e] = arch

	// 写入组件值：按组件 ID 找到对应列，保证“任意顺序”也正确
	for i, c := range comps {
		ci, _ := arch.columnOf(ids[i])
		arch.columns[ci].set(row, c)
	}
	return e
}

// Despawn 销毁实体，移除其全部组件并回收身份。
func (w *world) Despawn(e EntityID) {
	arch := w.entities[e]
	if arch == nil {
		return
	}
	arch.removeEntity(e)
	delete(w.entities, e)
}

// ---- 2) 实体单个组件操作 ----

// Get 读取实体组件 id 的指针（any 中持有 *T，可原位修改）；
// 实体不存在或没有该组件时返回 nil。
func (w *world) Get(e EntityID, id ComponentID) any {
	arch := w.entities[e]
	if arch == nil {
		return nil
	}
	ci, ok := arch.columnOf(id)
	if !ok {
		return nil
	}
	row, ok := arch.row[e]
	if !ok {
		return nil
	}
	return arch.columns[ci].ptr(row)
}

// Add 为实体添加组件 v，返回其组件 ID；组件类型未注册则自动注册。
// 实体没有该组件则创建并写入 v，已有则不覆盖；实体不存在则忽略（不 panic）。
// 注：v 必须是值类型（不允许指针/nil）。
func (w *world) Add(e EntityID, v any) ComponentID {
	id := w.RegisterComponent(v) // 内部完成值类型校验 + 注册（幂等）

	arch := w.entities[e]
	if arch == nil {
		return id // 实体不存在：忽略
	}
	ci, ok := arch.columnOf(id)
	if !ok {
		// 实体没有该组件 → 迁移到新 Archetype 并写入 v
		newArch := w.getOrCreateArchetype(arch.mask | maskOf(id))
		w.moveEntity(e, arch, newArch)
		arch = newArch
		ci, _ = arch.columnOf(id)
		row := arch.row[e]
		arch.columns[ci].set(row, v)
	}
	return id
}

// Remove 移除实体的组件 id；实体不存在或没有该组件时为空操作。
// 注意：移除最后一个组件后实体将不再拥有任何组件，会被自动注销。
func (w *world) Remove(e EntityID, id ComponentID) {
	arch := w.entities[e]
	if arch == nil {
		return
	}
	if _, ok := arch.columnOf(id); !ok {
		return
	}

	// 最后一个组件也被移除 → 实体注销
	// 不得不说deepseek这个设计很妙
	if arch.mask == maskOf(id) {
		w.Despawn(e)
		return
	}

	newArch := w.getOrCreateArchetype(arch.mask &^ maskOf(id))
	w.moveEntity(e, arch, newArch)
}

// 超集匹配缓存（先不做）
/**
func (w *world) findQueryCached(mask ComponentMask) *Query {
	// todo 这里有个问题待解决：如果一个 Archetype 的组件组合发生了变化（增删组件），那么它可能会从一个 Query 的结果中消失，或者出现在另一个 Query 的结果中。当前的缓存机制没有处理这种情况，可能会导致 Query 返回过时的结果。需要在增删组件时更新缓存，或者在 Query 时重新计算匹配的 Archetype。
	if q, ok := w.queryCache[mask]; ok {
		q.archIdx = 0
		q.rowIdx = 0
		return q
	}
	q := &Query{mask: mask}
	q.arches = make([]*Archetype, 0)
	for m, arch := range w.archetypes {
		if m&mask == mask {
			q.arches = append(q.arches, arch)
		}
	}
	w.queryCache[mask] = q
	return q
}
*/

// Query 查询拥有“至少”这些组件的实体（匹配所有包含该掩码的 Archetype）。
func (w *world) Query(mask ComponentMask, fn func(e EntityID, q *Query) bool) {
	q := &Query{mask: mask}
	q.arches = make([]*Archetype, 0)
	for m, arch := range w.archetypes {
		if m&mask == mask {
			q.arches = append(q.arches, arch)
		}
	}
	for q.Next() {
		if !fn(q.Entity(), q) {
			break
		}
	}
}

// ---- 4) 检查与调试 ----

// Alive 实体是否存活。
func (w *world) Alive(e EntityID) bool {
	_, ok := w.entities[e]
	return ok
}

// Count 存活实体数量。
func (w *world) Count() int {
	return len(w.entities)
}

// Entities 返回当前存活实体列表（用于观察/调试）。
func (w *world) Entities() []EntityID {
	list := make([]EntityID, 0, len(w.entities))
	for e := range w.entities {
		list = append(list, e)
	}
	return list
}

// BlockCount 返回当前存储块数量（Archetype 数量，用于观察/调试）。
func (w *world) BlockCount() int {
	return len(w.archetypes)
}

// ---- 泛型便捷函数 ----

// Add 为实体添加组件 T（零值），返回指向该组件的指针（可直接赋值初始化）。
// 组件类型未注册会自动注册；实体已有该组件则不覆盖；实体不存在时返回 nil。
// 其余增删改查以组件 ID 为主：Get(e,id) / Remove(e,id) 直接使用接口方法；
// 泛型 Get[T]/Has[T]/Set[T]/Remove[T] 已搁置（需要类型→ID 解析，见需求说明）。
func Add[T any](w World, e EntityID) *T {
	var zero T
	id := w.Add(e, zero)
	p := w.Get(e, id)
	if p == nil {
		return nil
	}
	return p.(*T)
}

// ---- 内部：实体分配 ----

// allocEntity 分配一个新的实体 ID（从 1 开始，0 保留为 InvalidEntity）。
func (w *world) allocEntity() EntityID {
	w.nextEntityID++
	return w.nextEntityID
}

// ---- 内部：Archetype 获取与迁移 ----

// getOrCreateArchetype 按精确掩码获取或创建 Archetype（按需创建）。
// 这正是本设计的关键：任何组件组合首次出现时才创建，无需预定义。
func (w *world) getOrCreateArchetype(mask ComponentMask) *Archetype {
	if a, ok := w.archetypes[mask]; ok {
		return a
	}
	a := newArchetype(mask, w)
	w.archetypes[mask] = a
	return a
}

// moveEntity 将实体从 from Archetype 迁移到 to Archetype，
// 并把 from 中已有的组件值整体拷贝到 to。
func (w *world) moveEntity(e EntityID, from, to *Archetype) {
	fromRow := from.row[e]
	toRow := to.addEntity(e)
	for i, cid := range from.compIDs {
		ci, ok := to.columnOf(cid)
		if !ok {
			continue // 不会发生：to 是 from 的超集或子集
		}
		to.columns[ci].data.Index(toRow).Set(from.columns[i].data.Index(fromRow))
	}
	from.removeEntity(e)
	w.entities[e] = to
}
