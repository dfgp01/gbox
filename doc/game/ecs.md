# gbox 游戏域（game）：ECS / FSM 设计文档

> 分层：公共定义 `def/game`、对外 API `pkg/game`、默认实现 `internal/game/ecs`。
> 本文件是使用方的 API 与设计说明。完整的三层规则见 [../架构.md](../架构.md)。

## 1. 依赖与导入约定

```go
import (
    def  "gbox/def/game" // 类型、接口、公共类
    game "gbox/pkg/game" // 工厂、流程
)
```

> 禁止 import `gbox/internal/game/ecs`（Go `internal` 规则下外部也无法导入）。

## 2. 核心概念

### 2.1 Entity 只有 ID

`def.EntityID`（`uint32`）只是唯一标识，不携带数据；`InvalidEntity = 0` 为哨兵。
实体与组件的关系由 `World` 维护。

### 2.2 Component 只是纯数据

- 普通 `struct` 值类型，无需实现接口、无需继承基类；类型身份由 Go 反射天然提供。
- 组件**不含** `EntityId` 字段；按值传给 `Spawn` / `RegisterComponent`，不允许传指针/nil。
- 组件类型按**组件组**注册（由 `ComponentMapper` 管理），一个组件组可被多个 `World` 副本共用；组件 ID 在所属组件组内唯一。
- 每个组件组最多 **64** 种组件类型（`ComponentMask` 为 `uint64` 位图）。

### 2.3 Archetype：掩码驱动的列式存储（内部）

相同组件组合的实体共享一个 `Archetype`，以精确 `ComponentMask` 标识，按需创建；
列式存储 + swap-remove 删除（O(1)）；增删组件 = 实体在 Archetype 间迁移。
这是 `internal` 的实现细节，使用方无需感知。

## 3. 使用流程

### 3.1 组件映射器（组件组）

`ComponentMapper` 专门管理一组组件的注册映射（组件类型 <-> 组件 ID）：
注册组件、批量注册、注册时类型校验、查找组件与生成组件 ID 都归它管。
按业务分类划分“组件组”，可被多个 `World` 副本共用（注册一次即可）。

```go
m := game.DefaultComponentMapper()         // 全局默认组件映射器（独一份）
idPos := m.RegisterComponent(Pos{})        // 注册一个组件类型，返回 def.ComponentID
idVel := m.RegisterComponent(Vel{})
ids := m.RegisterComponents(Pos{}, Vel{})  // 批量注册，返回 []def.ComponentID
```

一个游戏整体可以有多个组件组，用 int key 区分：

```go
charMgr := game.GetComponentMapper(1)      // 角色组件组
combatMgr := game.GetComponentMapper(2)    // 战斗组件组
charMgr.RegisterComponents(Pos{}, Vel{}, State{})
combatMgr.RegisterComponents(Hp{}, Atk{})
```

> 泛型快捷方式 `game.RegisterComponent[T]()` 在全局默认组件映射器中注册；
> 指定组件组用 `game.RegisterComponentIn[T](m)`。

### 3.2 创建世界（绑定组件组）

```go
w := game.NewWorld()                       // 使用全局默认组件组的注册表（副本共用）
w2 := game.NewWorldFrom(charMgr)           // 或绑定指定组件组
```

### 3.3 实体增删改查

```go
e := w.Spawn(Pos{X: 1, Y: 2}, Vel{Vx: 3})  // 任意组合、任意顺序
p := w.Get(e, idPos).(*Pos)                // 返回 *T 指针，可原位修改
p.X += 10

w.Add(e, Vel{Vx: 5})                       // 自动注册未注册类型；已存在则不覆盖
w.Remove(e, idVel)                         // 实体不存在/无该组件时为空操作
w.Despawn(e)

w.Alive(e)                                 // bool
w.Count()                                  // 存活实体数
```

> 增删组件会触发 Archetype 迁移（O(1)），任意时刻都可进行。

### 3.4 掩码与查询

```go
m := def.Mask(idPos, idVel)                // 组合掩码（超集匹配）
w.Query(m, func(e def.EntityID, q def.Query) bool {
    pos := q.Get(idPos).(*Pos)             // 可原位修改
    vel := q.Get(idVel).(*Vel)
    pos.X += vel.Vx
    return true                            // 返回 false 中断遍历
})
```

- `Query` 是**超集匹配**：命中“至少包含掩码中全部组件”的实体（如查 `A+B` 会命中 `AB`、`ABC`…）。
- 迭代过程中**不要对当前实体增删组件**（会迁移 Archetype、使迭代器失效）；确有需要可先收集快照。

### 3.5 系统与调度

系统实现 `def.System`：`Update(w def.World, dt float32)`。内嵌 `game.BaseSystem`（pkg）获得优先级与软开关：

```go
type MoveSystem struct {
    game.BaseSystem
    mask    def.ComponentMask
    stateID def.ComponentID
    posID   def.ComponentID
    velID   def.ComponentID
}

func (s *MoveSystem) Update(w def.World, dt float32) {
    w.Query(s.mask, func(e def.EntityID, q def.Query) bool {
        // ...
        return true
    })
}
```

```go
sm := game.NewSystemGroup()                // 系统组（仿照 ComponentMapper）
move := &MoveSystem{mask: def.Mask(...)}
move.SetPriority(10)                       // 越大越先执行（同优先级保持注册顺序）
sm.Add(move)                               // Add 后立即按优先级稳定排序

const dt = 1.0 / 60.0
sm.Update(w, dt)                           // 按优先级执行；可运行期 Enable/Disable 软开关
```

> 优先级/软开关由 `game.BaseSystem` 提供：`Priority() / SetPriority() / Enable() / Disable() / Enabled()`。
> 系统组也提供全局默认（`game.DefaultSystemGroup()`）与按 int key 的多系统组（`game.GetSystemGroup(key)`）。

## 4. FSM 有限状态机

### 4.1 公共定义（def/game/fsm.go）

- `StateID`：状态 ID，业务常量从 1 开始（0 = `StateInvalid` 哨兵）。
- `State`：状态组件（挂到实体上），记录 `Current/Previous/EnterTime/Duration`。
- `StateContext`：可选上下文组件（目标/路点等）。
- `StateNode`：节点接口 `OnEnter/OnUpdate/OnExit`，业务方以 OOP 方式实现。
- `TransitionRule`：`From → To` + `Condition` + `Apply`（消费一次性请求）。
- `StateMachine`：节点注册表 + 转换规则拓扑（全局一份），`NewStateMachine()` 创建。

### 4.2 业务状态定义

```go
const (
    StateIdle def.StateID = iota + 1
    StateRun
    StateJump
)
```

### 4.3 定义节点与规则

```go
type idleNode struct{}
func (n *idleNode) OnEnter(w def.World, e def.EntityID, s *def.State) {}
func (n *idleNode) OnUpdate(w def.World, e def.EntityID, s *def.State, dt float32) {}
func (n *idleNode) OnExit(w def.World, e def.EntityID, s *def.State)  {}

m := def.NewStateMachine()
m.RegisterNode(StateIdle, &idleNode{})
// ...
m.AddRules([]def.TransitionRule{
    {From: StateIdle, To: StateRun, Condition: hasMove},
    // ...（规则按添加顺序求值，被动/高优先级规则放前面）
}...)
```

### 4.4 接入世界与调度

```go
player := w.Spawn(def.State{Current: StateIdle}, Pos{...}, ...)

fsmSys := game.NewStateMachineSystem(m)   // game.StateMachineSystem，可 SetPriority/Enable/Disable
fsmSys.SetPriority(50)

sm := game.NewSystemGroup()
sm.Add(fsmSys)
sm.Update(w, dt)                          // 每帧驱动：OnEnter/OnUpdate/OnExit
```

### 4.5 每帧执行策略

对每个拥有 `State` 组件的实体：

1. 检查“退出此节点的标记”（发生转换）→ `OnExit`；
2. 检查“进入此节点的标记”（转换已提交 / 实体首次运行）→ `OnEnter`；
3. 检查“运行标记”→ `OnUpdate`（每帧，不发送 Event）。

> `State` 组件由 `StateMachineSystem` 首次 Update 时**惰性注册**，无需手动预注册。
> “发送 Event”目前为占位，属事件模块职责。

## 5. API 速查表

| 符号 | 位置 | 说明 |
|:---|:---|:---|
| `game.DefaultComponentMapper()` | pkg | 全局默认组件映射器（独一份） |
| `game.GetComponentMapper(key)` | pkg | 按 int key 获取/创建组件映射器（多组件组） |
| `m.RegisterComponent(v)` / `m.RegisterComponents(vs...)` | pkg | 在组件组中注册组件类型 |
| `m.IDOf(v)` / `m.TypeOf(id)` / `m.Len()` | pkg | 查找组件 / 生成对应类型 / 数量 |
| `game.NewWorld()` / `game.NewWorldFrom(m)` | pkg | 创建 `def.World`（默认 / 指定组件组） |
| `game.NewSystemGroup()` / `game.DefaultSystemGroup()` / `game.GetSystemGroup(key)` | pkg | 创建/获取系统组（按优先级调度） |
| `g.Add(s)` / `g.Update(w, dt)` | pkg | 系统组：添加并按优先级排序 / 按序执行 |
| `game.NewStateMachineSystem(m)` | pkg | 创建驱动 `m` 的状态机系统（`game.StateMachineSystem`） |
| `game.RegisterComponent[T]()` / `game.RegisterComponentIn[T](m)` | pkg | 在默认 / 指定组件组中注册组件类型 |
| `game.Add[T](w, e)` | pkg | 添加零值组件 T，返回 `*T` |
| `def.Mask(ids...)` | def | 组合 `def.ComponentMask` |
| `def.World` / `def.System` / `def.Query` | def | 核心接口 |
| `game.BaseSystem` | pkg | 系统基类（优先级 + 软开关） |
| `def.StateID / State / StateContext / StateNode / TransitionRule / StateMachine` | def | FSM 公共定义 |

## 6. 线程安全

框架本身不做线程安全控制：注册组件、增删实体、增删组件、增删 System 都可在任意时刻进行，
但并发访问需由外层调用方自行保证（主循环内部保持单线程，见 `pkg/game.GameStart` 说明）。

## 7. 示例

完整可运行示例见 `src/demo/ecs/`（`RunECSBasic` 演示 ECS 增删改查与查询，`RunFSM` 演示状态机）。
