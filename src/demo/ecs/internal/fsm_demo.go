package internal

import (
	"fmt"

	def "gbox/def/game"
	game "gbox/pkg/game"
)

// ---- 业务状态定义（业务层，不属于框架）----

// 状态 ID 从 1 开始（0 为框架保留的 StateInvalid）。
const (
	StateIdle def.StateID = iota + 1 // 站立
	StateRun                         // 走动
	StateJump                        // 跳跃
	StateHit1                        // 被打1：站立/走动时被打
	StateHit2                        // 被打2：跳跃时被打
)

// ---- 状态机所需的额外组件 ----

// MoveCommand 标记：有移动指令（存在即“要移动”）。
type MoveCommand struct{}

// JumpRequest 一次性请求：请求跳跃（转换时由 Apply 消费，避免重复触发）。
type JumpRequest struct{}

// HitRequest 一次性请求：受击（转换时由 Apply 消费，避免重复触发）。
type HitRequest struct{}

// Grounded 标记：着地（用于“跳跃→站立”的落地条件）。
type Grounded struct{}

// ---- 状态节点（OOP 方式定义每个状态的行为）----

type idleNode struct{ name string }

func (n *idleNode) OnEnter(w def.World, e def.EntityID, s *def.State) {
	fmt.Printf("[%s] OnEnter 站立\n", n.name)
}
func (n *idleNode) OnUpdate(w def.World, e def.EntityID, s *def.State, dt float32) {
	// 站立行为：待机/巡逻等（示例省略）
}
func (n *idleNode) OnExit(w def.World, e def.EntityID, s *def.State) {
	fmt.Printf("[%s] OnExit  站立\n", n.name)
}

type runNode struct{ name string }

func (n *runNode) OnEnter(w def.World, e def.EntityID, s *def.State) {
	fmt.Printf("[%s] OnEnter 走动\n", n.name)
}
func (n *runNode) OnUpdate(w def.World, e def.EntityID, s *def.State, dt float32) {
	// 走动行为：加速度/寻路等由各自行为 System 处理，节点这里不做转换判断
}
func (n *runNode) OnExit(w def.World, e def.EntityID, s *def.State) {
	fmt.Printf("[%s] OnExit  走动\n", n.name)
}

// jumpNode 需要移除 Grounded 组件，因此持有其组件 ID。
type jumpNode struct {
	name       string
	groundedID def.ComponentID
}

func (n *jumpNode) OnEnter(w def.World, e def.EntityID, s *def.State) {
	fmt.Printf("[%s] OnEnter 跳跃\n", n.name)
	// 起跳离地：移除着地标记
	w.Remove(e, n.groundedID)
}
func (n *jumpNode) OnUpdate(w def.World, e def.EntityID, s *def.State, dt float32) {
	// 跳跃行为：抛物线等（示例省略）
}
func (n *jumpNode) OnExit(w def.World, e def.EntityID, s *def.State) {
	fmt.Printf("[%s] OnExit  跳跃\n", n.name)
}

type hitNode struct {
	name string
	kind int // 1：站立/走动被打；2：跳跃被打
}

func (n *hitNode) OnEnter(w def.World, e def.EntityID, s *def.State) {
	fmt.Printf("[%s] OnEnter 被打%d（硬直开始）\n", n.name, n.kind)
}
func (n *hitNode) OnUpdate(w def.World, e def.EntityID, s *def.State, dt float32) {
	// 硬直表现（示例省略）
}
func (n *hitNode) OnExit(w def.World, e def.EntityID, s *def.State) {
	fmt.Printf("[%s] OnExit  被打%d（硬直结束）\n", n.name, n.kind)
}

// ---- 系统 ----

// runMoveSystem 是“行为 System”：不关心状态转换，只关心“当前是不是我的状态”。
// 只有 StateRun 的实体才会被移动。内嵌 BaseSystem 获得优先级与软开关。
// 掩码查询（State, Pos, Vel），运行期用 Next/Get 迭代，零查找。
type runMoveSystem struct {
	game.BaseSystem
	mask    def.ComponentMask
	stateID def.ComponentID
	posID   def.ComponentID
	velID   def.ComponentID
}

func (s *runMoveSystem) Update(w def.World, dt float32) {
	w.Query(s.mask, func(e def.EntityID, q def.Query) bool {
		state := q.Get(s.stateID).(*def.State)
		pos := q.Get(s.posID).(*Pos)
		vel := q.Get(s.velID).(*Vel)
		if state.Current != StateRun {
			return true // 不是跑动状态，跳过
		}
		pos.X += vel.Vx * dt
		pos.Y += vel.Vy * dt
		fmt.Printf("        [move] pos=(%.2f, %.2f)\n", pos.X, pos.Y)
		return true
	})
}

// scriptStep 脚本化输入，用于演示各条转换路径。
type scriptStep struct {
	frame  int
	action string // move / jump / hit / land
	on     bool   // 对 move 有效：开 / 关
}

var scriptSteps = []scriptStep{
	{2, "move", true},   // 输入：移动 → 站立→走动
	{5, "jump", true},   // 输入：跳跃（同时停止移动）→ 走动→跳跃
	{7, "hit", true},    // 输入：受击（跳跃中）→ 跳跃→被打2
	{15, "land", true},  // 输入：落地 → 跳跃→站立
	{19, "hit", true},   // 输入：受击（站立）→ 站立→被打1
	{28, "move", true},  // 输入：移动 → 站立→走动
	{32, "move", false}, // 输入：停止移动 → 走动→站立
}

// scriptSystem 按帧注入输入组件。内嵌 BaseSystem 获得优先级与软开关。
// 移除组件用组件 ID（moveCmdID）；新增组件直接用 w.Add 传值。
type scriptSystem struct {
	game.BaseSystem
	player    def.EntityID
	moveCmdID def.ComponentID
	steps     []scriptStep
	frame     int
}

func (s *scriptSystem) Update(w def.World, dt float32) {
	s.frame++
	for _, st := range s.steps {
		if st.frame != s.frame {
			continue
		}
		switch st.action {
		case "move":
			if st.on {
				fmt.Println(">>> 输入：移动")
				w.Add(s.player, MoveCommand{})
			} else {
				fmt.Println(">>> 输入：停止移动")
				w.Remove(s.player, s.moveCmdID)
			}
		case "jump":
			fmt.Println(">>> 输入：跳跃")
			w.Remove(s.player, s.moveCmdID) // 起跳即停止移动
			w.Add(s.player, JumpRequest{})
		case "hit":
			fmt.Println(">>> 输入：受击")
			w.Add(s.player, HitRequest{})
		case "land":
			fmt.Println(">>> 输入：落地")
			w.Add(s.player, Grounded{})
		}
	}
}

// RunFSM 演示有限状态机（站立/走动/跳跃/被打1/被打2）。
func RunFSM() {
	fmt.Println("===== FSM 演示 =====")

	// 创建世界，并在世界中注册所有组件类型（未注册的类型会 panic）
	w := game.NewWorld()
	idState := game.RegisterComponent[def.State]()
	game.RegisterComponent[def.StateContext]()
	idMoveCmd := game.RegisterComponent[MoveCommand]()
	idJumpReq := game.RegisterComponent[JumpRequest]()
	idHitReq := game.RegisterComponent[HitRequest]()
	idGrounded := game.RegisterComponent[Grounded]()
	idPos := game.RegisterComponent[Pos]()
	idVel := game.RegisterComponent[Vel]()

	// 转换条件与副作用（闭包捕获组件 ID，基于 w.Get / w.Remove 判断）。
	// 注意：条件签名固定为 func(w, e, s)，因此以闭包形式构建。
	gotHit := func(w def.World, e def.EntityID, s *def.State) bool { return w.Get(e, idHitReq) != nil }
	consumeHit := func(w def.World, e def.EntityID, s *def.State) { w.Remove(e, idHitReq) }
	hasMove := func(w def.World, e def.EntityID, s *def.State) bool { return w.Get(e, idMoveCmd) != nil }
	noMove := func(w def.World, e def.EntityID, s *def.State) bool { return w.Get(e, idMoveCmd) == nil }
	wantJump := func(w def.World, e def.EntityID, s *def.State) bool { return w.Get(e, idJumpReq) != nil }
	consumeJump := func(w def.World, e def.EntityID, s *def.State) { w.Remove(e, idJumpReq) }
	landed := func(w def.World, e def.EntityID, s *def.State) bool { return w.Get(e, idGrounded) != nil }
	// hitRecovered 硬直 0.1 秒（60fps 约 6 帧）后恢复。
	hitRecovered := func(w def.World, e def.EntityID, s *def.State) bool { return s.Duration >= 0.1 }

	// 全局状态机：节点注册表 + 转换规则拓扑（全局一份）
	m := def.NewStateMachine()
	m.RegisterNode(StateIdle, &idleNode{name: "idle"})
	m.RegisterNode(StateRun, &runNode{name: "run"})
	m.RegisterNode(StateJump, &jumpNode{name: "jump", groundedID: idGrounded})
	m.RegisterNode(StateHit1, &hitNode{name: "hit1", kind: 1})
	m.RegisterNode(StateHit2, &hitNode{name: "hit2", kind: 2})

	// 受击是被动事件，优先级最高，放在最前。
	//
	// 状态拓扑（全局一份）：站立↔走动、站立/走动→跳跃、落地→站立；
	// 站立/走动被打→被打1，跳跃被打→被打2；被打1恢复→站立，被打2恢复→跳跃。
	rules := []def.TransitionRule{
		// 受击（被动）
		{From: StateIdle, To: StateHit1, Condition: gotHit, Apply: consumeHit},
		{From: StateRun, To: StateHit1, Condition: gotHit, Apply: consumeHit},
		{From: StateJump, To: StateHit2, Condition: gotHit, Apply: consumeHit},
		// 跳跃
		{From: StateIdle, To: StateJump, Condition: wantJump, Apply: consumeJump},
		{From: StateRun, To: StateJump, Condition: wantJump, Apply: consumeJump},
		// 移动
		{From: StateIdle, To: StateRun, Condition: hasMove},
		{From: StateRun, To: StateIdle, Condition: noMove},
		// 落地
		{From: StateJump, To: StateIdle, Condition: landed},
		// 受击恢复：被打1 恢复站立；被打2 恢复跳跃（在空中继续下落，落地后再回站立）
		{From: StateHit1, To: StateIdle, Condition: hitRecovered},
		{From: StateHit2, To: StateJump, Condition: hitRecovered},
	}
	m.AddRules(rules...)

	// 创建玩家实体：初始站立、着地
	player := w.Spawn(
		def.State{Current: StateIdle},
		def.StateContext{}, // 可选：存储目标/路点等上下文
		Grounded{},
		Pos{X: 0, Y: 0},
		Vel{Vx: 2, Vy: 0},
	)

	// 系统调度：输入 → 状态机 → 行为（移动）。
	// 打乱注册顺序，Add 后系统组立即按优先级稳定排序：
	//   脚本输入(100) > 状态机(50) > 移动行为(10)
	// 业务系统内嵌 BaseSystem，通过 SetPriority 设置优先级。
	moveSys := &runMoveSystem{
		mask:    def.Mask(idState, idPos, idVel),
		stateID: idState,
		posID:   idPos,
		velID:   idVel,
	}
	moveSys.SetPriority(10)
	fsmSys := game.NewStateMachineSystem(m)
	fsmSys.SetPriority(50)
	scriptSys := &scriptSystem{player: player, moveCmdID: idMoveCmd, steps: scriptSteps}
	scriptSys.SetPriority(100)

	sm := game.NewSystemGroup()
	sm.Add(moveSys)
	sm.Add(fsmSys)
	sm.Add(scriptSys)

	const dt = 1.0 / 60.0
	for frame := 0; frame < 40; frame++ {
		// 演示软开关：游戏进行中仅切换 Enable/Disable，不做数组增删
		if frame == 29 {
			fmt.Println(">>> 软开关：禁用移动系统")
			moveSys.Disable()
		} else if frame == 30 {
			fmt.Println(">>> 软开关：重新启用移动系统")
			moveSys.Enable()
		}
		sm.Update(w, dt)
	}

	pos := w.Get(player, idPos).(*Pos)
	fmt.Printf("end state=%d pos=(%.2f, %.2f)\n", w.Get(player, idState).(*def.State).Current, pos.X, pos.Y)
	fmt.Println()
}
