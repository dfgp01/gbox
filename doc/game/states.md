# 层级状态机

## 包职责

- `gbox/def/game/states`：状态机公共接口、状态 ID、上下文、Guard、Action、Transition 等定义。
- `gbox/pkg/game/states`：对外 Graph/FSM API，以及 `All`、`Any`、`Not` 等 Guard 组合实现。
- `gbox/internal/game/states`：节点基础实现和层级节点调度实现。

本模块独立于 ECS，也不依赖旧版 `gbox/def/game` FSM。

## 状态图与实例

状态图由 Node 及其子节点、出边组成。图定义可由多个 FSM 实例共享；每个单位的当前状态、上一个状态、进入时间、持续时间、上下文和层级节点运行数据则保存在各自的 FSM 中。建图后应将节点图视为只读，不要在回调或 Guard 中捕获并修改某个单位的实例数据；实例数据通过 `FSMContext` 传递。

`Node` 提供 ID、名称、类型、生命周期回调、节点级 Guards、子节点、默认出口和出边。节点只描述自己的出边，不描述由哪个节点进入。ID 在整张图中必须非零且唯一；同一父节点下的子节点名称必须唯一。

internal 中的私有 `nodeBase` 只保存所有节点真正共有的 ID、名称和节点级 Guards，不作为对外类型。`LeafNode` 增加默认出口和出边，并通过叶子行为承载 Actions/Addons；业务节点可嵌入 `LeafNode` 并覆写 `OnEnter`、`OnTick`、`OnExit`。`TerminalNode` 是独立的可执行叶子，不提供默认出口或出边，但同样可以配置 Actions/Addons。`SerialNode`、`ParallelNode`、`SelectorNode` 各自只保存自己的子节点，并在自己的 `OnEnter`、`OnTick`、`OnExit` 中初始化、驱动和清理子节点。每个实例的层级运行数据经 `FSMContext.Execution` 存放在对应 FSM 中，不保存在共享节点上。

```go
idle := states.NewLeafNode(1, "idle")
moving := states.NewLeafNode(2, "moving")
done := states.NewTerminalNode(3, "done")

idle.Edges = []def.Transition{{
	Weight: 10,
	Guards: states.FunctionalGuards{Check: func(ctx def.FSMContext) bool {
		return ctx.Data.(PlayerContext).WantsToMove
	}},
	To: moving,
}}

graph, err := states.NewGraph(idle)
if err != nil {
	return err
}
playerFSM, err := states.NewFSM(graph, idle.ID(), PlayerContext{})
```

## Run-to-completion

FSM 每帧调用一次 `Tick(dt)`。Tick 驱动当前状态及活动层级子节点，计算迁移意图，再由 FSM 主调度器统一仲裁并提交迁移。迁移提交后，本 tick 不会继续评估目标节点的出边：例如 A 切换到 B 后，B 的 `OnEnter` 即使产生了新的迁移相关上下文标记，也要等后续 tick 再评估，不能在同一调用链里继续从 B 切到 C。这样避免进入回调递归触发迁移形成死循环。

每个节点按权重从高到低、同权重按声明顺序检查出边，并只取第一条满足条件的边。层级节点可以各自产生迁移意图；如果同一 FSM 的活动节点同时提出多个意图，主调度器按权重选择一条，平权时按父节点优先、再按子节点声明顺序决定。每个 FSM 每帧最多提交一次状态切换。

迁移提交的生命周期顺序为：当前节点 `OnExit`、获选 Transition 的 Action、更新 Current/Previous/Duration/EnterTime、目标节点 `OnEnter`。目标节点本 tick 不再执行 `OnTick` 或迁移评估，下一 tick 才开始正常驱动。

## Guards 与默认出口

Guard 必须保持纯粹，只判断“是否允许迁移”，不得扣血、扣弹药、消费请求或修改外部状态。

- **节点级 Guards**：Tick 当前节点前评估。检查失败时不运行该节点的 tick 行为，而是请求跳转到 `DefaultTo`（实现字段为 `FailTo`）。未配置默认出口的终态可以没有节点级 Guards。
- **迁移边 Guards**：按出边顺序和权重逐条检查。某条边的 Guard 不通过，只表示该边不匹配，继续检查下一条；所有迁移边都不匹配时保持当前状态，不会因此触发节点默认出口。
- **副作用**：放在获选迁移的 Action 或目标状态的 `OnEnter` 中执行，并遵守 run-to-completion；Guard 只负责判定。`Action` 是定义 `OnEnter`、`OnTick`、`OnExit` 的接口。迁移 Action 在提交阶段执行 `OnEnter`；叶子节点和终态节点可配置 Action 序列，在对应生命周期执行。复合节点不承载 Actions/Addons。

## 上下文与事件标记

`FSMContext` 保存每个 FSM 实例的数据和本帧信息：`Data` 是业务上下文，`Time` 是状态机累计时间，`Delta` 是本帧间隔，`Duration` 是当前状态持续时间。回调接收上下文并返回更新后的上下文。

事件只用于向上冒泡，不提供事件监听器、订阅或分发系统。需要上报时，在上下文的事件标记中记录，由调用方在 tick 返回后处理；标记本身不应绕过主调度器直接触发第二次迁移。

## Guard 组合

`All` 所有子 Guard 成立才通过，空集合按逻辑恒真处理，并在遇到首个失败时短路。`Any` 任一子 Guard 成立即通过，空集合为假，并在首个成功分支处短路。`Not` 要求且只接受一个子 Guard；参数数量不正确时返回失败。组合 Guard 返回被选择或失败的分支索引路径，供上层记录。

## 层级节点

- `SerialNode`：按声明顺序运行子节点；子节点成功后推进，遇到失败则整体失败。
- `SelectorNode`：按声明顺序尝试子节点；首个成功即整体成功，全部失败则整体失败。
- `ParallelNode`：按声明顺序驱动所有尚未完成的子节点；全部成功时整体成功，任一失败则整体失败。

活动子节点和完成状态属于各 FSM 实例，不存储在共享节点定义上。活动子节点提出的迁移同样交由 FSM 主调度器统一仲裁。当前实现中，若子节点的迁移意图获选，目标会成为 FSM 的当前节点，并退出当前复合节点树；尚未实现保留复合父节点的局部子节点间迁移。
