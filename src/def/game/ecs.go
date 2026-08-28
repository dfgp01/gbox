package game

// 组件（Component）约定
//
// 组件在本框架中被约定为“纯数据”：
//   - 使用普通 struct 值类型，不需要实现任何接口，也不需要继承任何基类；
//   - 允许被拷贝、按值存储，不持有外部依赖；
//   - 不包含 EntityId 字段——实体与组件的关系由 World/Archetype 维护；
//   - 组件必须按“值”传给 RegisterComponent / Spawn，不允许传指针。
//
// 之所以不强制“规范接口 / 统一基类”：
//  1. 组件是数据，强制接口会引入耦合，破坏“纯数据”的初衷；
//  2. 类型身份（identity）由 Go 的类型系统（reflect.Type）天然提供，
//     框架据此区分不同组件；
//  3. 值语义的组件方便按值存储、复用与拷贝。
//
// 组件类型约束（按组件组注册）：
//   - 组件类型通过 pkg/game.ComponentMapper 注册，
//     一个组件映射器对应一个“组件组”，可被多个 World 副本共用；
//   - 每个组件组最多支持 64 种组件类型（受 ComponentMask 位宽限制），
//     超出由调用方自行负责（框架不再强制 panic）；
//   - 注册/增删可在任意时刻进行，不做生命周期限制；
//     框架本身不保证线程安全，并发访问由外层调用方控制；
//   - 若使用未注册的组件类型，将直接 panic，以尽早暴露错误。
//
// 命名上：同一包内使用 Pos、Move 这类短名即可；若担心跨包混淆，
// 也可以命名为 PosComponent、MoveComponent，框架不做强制。

// ComponentID 组件类型 ID（所属组件组内全局唯一，1~64）。
// 每种组件类型占 ComponentMask 中的一个 bit：组件 ID=i 占第 i-1 位（见 maskOf）。
// 0 保留为 ComponentIDInvalid（无效哨兵），其掩码贡献为 0（全 0，与任何掩码相与仍为 0）。
type ComponentID uint8

// ComponentIDInvalid 无效组件 ID（哨兵值），不会分配给任何组件类型。
const ComponentIDInvalid ComponentID = 0

// ComponentMask 组件组合掩码（位图）。uint64 共有 64 个 bit，编号 0~63；
// 组件 ID=i 使用第 i-1 位（ID 1→bit0、ID 2→bit1、…、ID 64→bit63），
// 因此一个 World 最多可注册 64 种组件类型（ID 1~64），64 个 bit 全部用上。
// 它是 Archetype 的身份：掩码相同 ⇒ 组件组合相同 ⇒ 属于同一个 Archetype。
type ComponentMask uint64

// MaxComponents 默认支持的最大组件类型数（64：ID 1~64 对应 bit 0~63）。
const MaxComponents = 64

// maskOf 返回组件 id 在掩码中占用的位值（ID=i 占第 i-1 位，即 1<<(i-1)）。
// ComponentIDInvalid（0）贡献 0，不参与任何掩码。
func maskOf(id ComponentID) ComponentMask {
	if id == ComponentIDInvalid {
		return 0
	}
	return 1 << (id - 1)
}

// Mask 将一组组件 ID 组合为掩码。
// 例如 Mask(idPos, idVel) 得到“同时拥有 Pos 与 Vel”的掩码。
func Mask(ids ...ComponentID) ComponentMask {
	var m ComponentMask
	for _, id := range ids {
		m |= maskOf(id)
	}
	return m
}

// EntityID 实体 ID，仅是一个唯一标识，本身不携带任何数据。
// 组件与实体的关系由 World / Archetype 维护，实体只是“组件集合”的身份标识。
type EntityID uint32

// InvalidEntity 无效实体（哨兵值），用于表示“无实体”，Spawn 不会分配该值。
const InvalidEntity EntityID = 0

// System 统一系统接口：方法为 Update。
type System interface {
	// Update 每帧执行一次；w 为所属世界，dt 为帧间隔（毫秒）。
	Update(w World, dt float32)
	Priority() int
	Enable()
	Disable()
	Enabled() bool
}

// Query 查询迭代器接口：World.Query 回调内通过它读取当前实体的组件。
type Query interface {
	// Entity 返回当前实体 ID。
	Entity() EntityID
	// Get 返回当前实体上组件 id 的指针（any 中持有 *T），可原位修改。
	Get(id ComponentID) any
}

// World 世界引擎接口：负责管理所有实体与组件数据（内部为 Archetype 列式存储）。
// 按功能分组接口：
// 1) 实体操作：Spawn / Despawn
// 2) 实体单个组件操作：Get / Add / Remove
// 3) 批量查询操作：Query
// 4) 检查和debug：Alive / Count / Entities
type World interface {
	// Spawn 创建实体并附加组件值（任意组合、任意顺序）comp 必须是 struct{}类型。
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
	Query(mask ComponentMask, fn func(e EntityID, q Query) bool)
	// Alive 实体是否存活。
	Alive(e EntityID) bool
	// Count 存活实体数量。
	Count() int
	// Entities 返回当前存活实体列表（用于观察/调试）
	Entities() []EntityID
}
