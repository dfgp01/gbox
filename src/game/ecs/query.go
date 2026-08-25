package ecs

// Query 批量查询：匹配“拥有 mask 中全部组件（允许更多）”的实体。
//
// 说明：World.archetypes 以“精确掩码”为 key，一个精确掩码只对应一个 Archetype；
// 但 Query 是“超集匹配”（m&mask == mask，即至少包含这些组件），
// 因此一个查询掩码可能命中多个 Archetype。
// 例如查 A+B 会命中 {AB}、{ABC}、{ABD}、{ABCD}... 多个 Archetype，
// 所以 arches 是数组，而不是单个 Archetype。
//
// Query 是轻量对象：每次 w.Query(mask) 都会扫描一次 Archetype 表，
// 在热路径可缓存 Query 对象复用。
//
// 迭代过程中不要对“当前实体”增删组件（会触发 Archetype 迁移、使迭代器失效）。
// 若确有需要，可先收集实体快照再处理（StateMachineSystem 就是这么做的，见 fsm.go）。
type Query struct {
	mask    ComponentMask // 查询掩码
	arches  []*Archetype  // 匹配的 Archetype 列表（超集匹配）
	archIdx int           // 当前 Archetype 下标（超集匹配）
	rowIdx  int           // 当前 Archetype 内的行号（下一行）
}

// Next 前进到下一个匹配实体；无更多匹配时返回 false。
// 语义：rowIdx 表示“下一个待消费的行号”，当前行 = rowIdx-1。
func (q *Query) Next() bool {
	// 超集匹配：遍历 arches 列表，跨 Archetype 前进
	for {
		if q.archIdx >= len(q.arches) {
			return false
		}
		arch := q.arches[q.archIdx]
		if q.rowIdx < arch.n {
			q.rowIdx++
			return true
		}
		// 当前 Archetype 遍历完，切换到下一个
		q.archIdx++
		q.rowIdx = 0
	}
}

// Entity 返回当前实体 ID。
func (q *Query) Entity() EntityID {
	return q.arches[q.archIdx].entities[q.rowIdx-1]
}

// Get 返回当前实体上组件 id 的指针（any 中持有 *T），可原位修改。
// 例如：pos := q.Get(idPos).(*Pos)；pos.X += 1 会真实写回存储。
func (q *Query) Get(id ComponentID) any {
	return q.arches[q.archIdx].getPtr(id, q.rowIdx-1)
}
