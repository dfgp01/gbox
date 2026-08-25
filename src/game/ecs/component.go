package ecs

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
// 组件类型约束（按世界注册）：
//   - 组件类型通过 RegisterComponent[T](w) 在“所属 World”上注册，
//     注册表按世界独立，组件 ID 只在所属 World 内有效；
//   - 每个 World 最多支持 64 种组件类型（受 ComponentMask 位宽限制），
//     超出由调用方自行负责（框架不再强制 panic）；
//   - 注册/增删可在任意时刻进行，不做生命周期限制；
//     框架本身不保证线程安全，并发访问由外层调用方控制；
//   - 若使用未注册的组件类型，将直接 panic，以尽早暴露错误。
//
// 命名上：同一包内使用 Pos、Move 这类短名即可；若担心跨包混淆，
// 也可以命名为 PosComponent、MoveComponent，框架不做强制。

// ComponentID 组件类型 ID（世界内全局唯一，1~64）。
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

// maxComponents 支持的最大组件类型数（64：ID 1~64 对应 bit 0~63）。
const maxComponents = 64

// maskOf 返回组件 id 在掩码中占用的位值（ID=i 占第 i-1 位，即 1<<(i-1)）。
// ComponentIDInvalid（0）贡献 0，不参与任何掩码。
func maskOf(id ComponentID) ComponentMask {
	if id == ComponentIDInvalid {
		return 0
	}
	return 1 << (id - 1)
}

// RegisterComponent 在世界 w 中注册组件类型 T，返回其组件 ID（幂等）。
// 具体注册策略（如数量上限）由 World 实现自行决定；任意时刻都可调用，
// 线程安全由外层调用方保证。
func RegisterComponent[T any](w World) ComponentID {
	var zero T
	return w.RegisterComponent(zero)
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
