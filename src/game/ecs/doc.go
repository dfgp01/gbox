// Package ecs 提供一套轻量级、类型安全的 ECS（Entity-Component-System）框架，
// 以及构建在其之上的有限状态机（FSM）模块。
//
// 设计理念
//
//  1. Entity 只有 ID
//     EntityID 只是唯一标识，本身不携带任何数据；实体与组件的关系由 World 维护。
//
//  2. Component 只是纯数据
//     组件是普通 struct 值类型，不强制实现接口、也不继承基类；身份由 Go 类型系统
//     （reflect.Type）天然提供。组件类型按世界注册：RegisterComponent[T](w)，
//     注册表在每个 World 内独立，组件 ID 只在所属 World 内有效；
//     每个 World 最多支持 64 种组件类型（ComponentMask 为 uint64 位图，见 component.go）。
//
//  3. Archetype：掩码驱动的列式存储（见 archetype.go）
//     相同组件组合的实体共享一个 Archetype，用 ComponentMask（uint64 位图）精确标识。
//     World.archetypes 以“精确掩码”为 key、按需创建——无需预定义任何组合结构体；
//     实体在 Archetype 内列式存储、swap-remove 删除（O(1)）；
//     增删组件 = 实体从一个 Archetype 迁移到另一个 Archetype。
//
//  4. World 是接口（见 world.go）
//     World 接口暴露 Spawn/Despawn/Alive/Count/Get/Add/Remove/RegisterComponent/Query；
//     反射只出现在 archetype.go；泛型便捷函数 RegisterComponent[T]/Add[T] 以包级函数提供
//     （Go 不允许接口的泛型方法）。
//
//  5. Query：掩码匹配 + 回调遍历（见 query.go）
//     w.Query(mask, fn) 超集匹配“至少包含 mask 中全部组件”的实体，
//     一个查询掩码可命中多个 Archetype（如查 A+B 命中 AB、ABC、ABCD...）；
//     fn(e EntityID) bool 对每个匹配实体回调，返回 false 中断遍历；
//     回调内用 w.Get(e, id) 读取/修改组件。
//
//  6. System 通过 Query 批量处理（见 system.go）
//     System 统一接口 Update(w World, dt)；内嵌 BaseSystem 获得优先级与软开关；
//     SystemManager 按优先级稳定排序执行，运行期只读遍历 + Enable/Disable 软开关。
//
//  7. 无生命周期限制
//     注册组件、创建/销毁实体、增删组件、增删 System 都可在任意时刻进行；
//     框架不做线程安全控制，并发访问需由外层调用方自行保证。
//
// 有限状态机（见 fsm.go）
//
//   - 状态数据（State）作为组件挂在实体上；业务状态（StateID 常量）由使用方定义；
//   - 状态节点（StateNode）与状态拓扑（转换规则）全局只有一份，
//     使用者以 OOP 方式定义每个状态的行为，由 StateMachineSystem 通过 ECS 机制驱动；
//   - 每帧执行策略：OnEnter / OnUpdate / OnExit 按进入/运行/退出标记执行
//     （Event 发送为本模块占位，属事件模块职责）。
package ecs
