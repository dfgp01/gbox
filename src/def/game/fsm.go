package game

// ============================================================
// 有限状态机（FSM）模块 —— 公共定义层（def/game）
//
// 本文件只包含 FSM 的“公共类 / 枚举 / 接口”：
//   - 状态枚举：StateID（业务状态由使用方定义，从 1 开始，0 为哨兵）
//   - 组件：State（状态数据）、StateContext（状态上下文，可选）
//   - 节点接口：StateNode（OnEnter / OnUpdate / OnExit）
//   - 拓扑：TransitionRule（转换规则）、StateMachine（节点注册表 + 规则，全局一份）
//
// 驱动逻辑（StateMachineSystem）是默认实现，位于 internal/game/ecs，
// 通过 pkg/game.NewStateMachineSystem 对外提供。
// ============================================================

// StateID 状态 ID。
type StateID int32

const (
	// StateInvalid 无效状态（哨兵，业务状态从非 0 开始）。
	StateInvalid StateID = 0
)

// State 状态组件：纯数据。
type State struct {
	Current   StateID // 当前状态
	Previous  StateID // 上一个状态（转换时记录）
	EnterTime float32 // 进入当前状态的时间戳（由 StateMachineSystem 内部时钟累计）
	Duration  float32 // 当前状态已持续时长（每帧累加，可做定时/冷却判断）

	entered bool // 内部：该实体是否已进入过初始状态（由 StateMachineSystem 维护）
}

// Entered 该实体是否已触发过初始 OnEnter（由 StateMachineSystem 驱动使用）。
func (s *State) Entered() bool { return s.entered }

// MarkEntered 标记该实体已触发过初始 OnEnter（由 StateMachineSystem 驱动使用）。
func (s *State) MarkEntered() { s.entered = true }

// StateContext 状态上下文（可选）：存储状态机所需的额外判断/行为数据。
type StateContext struct {
	TargetEntity EntityID // 攻击目标
	WaypointIdx  int      // 巡逻路点
}

// StateNode 状态节点接口：框架使用者以 OOP 方式实现每个状态的行为。
// 节点与状态拓扑一样是全局一份（每个 StateID 对应一个节点实例）。
type StateNode interface {
	// OnEnter 进入该状态时执行一次。
	OnEnter(w World, e EntityID, s *State)
	// OnUpdate 处于该状态的每帧执行（不发送 Event）。
	OnUpdate(w World, e EntityID, s *State, dt float32)
	// OnExit 退出该状态时执行一次。
	OnExit(w World, e EntityID, s *State)
}

// TransitionRule 转换规则：当 Condition 满足（或为空=恒真）时，从 From 转到 To。
// 规则按添加顺序求值，先满足者先转换；因此“被动/高优先级”的规则应放在前面
// （例如受击应优先于主动移动/跳跃）。
type TransitionRule struct {
	From StateID
	To   StateID
	// Condition 转换条件。为 nil 表示无条件转换。
	Condition func(w World, e EntityID, s *State) bool
	// Apply 可选：转换提交后立即执行，通常用于“消费”一次性请求组件
	// （例如受击请求 HitRequest、跳跃请求 JumpRequest），
	// 避免条件在后续帧再次满足导致重复转换。
	Apply func(w World, e EntityID, s *State)
}

// StateMachine 状态机定义：节点注册表 + 转换规则拓扑（全局一份）。
type StateMachine struct {
	nodes   []StateNode
	rules   []TransitionRule
	elapsed float32 // 内部时钟（秒），用于 EnterTime
}

// NewStateMachine 创建一个空状态机。
func NewStateMachine() *StateMachine {
	return &StateMachine{}
}

// Tick 累计内部时钟（每帧由驱动系统 StateMachineSystem 调用）。
func (m *StateMachine) Tick(dt float32) { m.elapsed += dt }

// Elapsed 返回内部时钟累计值（秒）。
func (m *StateMachine) Elapsed() float32 { return m.elapsed }

// RegisterNode 注册 StateID 对应的状态节点。空缺档位用 noopNode 兜底。
func (m *StateMachine) RegisterNode(id StateID, n StateNode) {
	idx := int(id)
	for len(m.nodes) < idx {
		m.nodes = append(m.nodes, noopNode{})
	}
	m.nodes = append(m.nodes, n)
}

// Node 返回 StateID 对应的节点；未注册时返回 noopNode（空实现，安全调用）。
func (m *StateMachine) Node(id StateID) StateNode {
	idx := int(id)
	if idx < 0 || idx >= len(m.nodes) {
		return noopNode{}
	}
	return m.nodes[idx]
}

// AddRule 追加一条转换规则。
func (m *StateMachine) AddRule(r TransitionRule) {
	m.rules = append(m.rules, r)
}

// AddRules 批量追加转换规则。
func (m *StateMachine) AddRules(rs ...TransitionRule) {
	m.rules = append(m.rules, rs...)
}

// Resolve 在 m.rules 中查找第一条满足条件（From == s.Current 且 Condition 成立）的规则。
func (m *StateMachine) Resolve(w World, e EntityID, s *State) (*TransitionRule, bool) {
	for i := range m.rules {
		r := &m.rules[i]
		if r.From == s.Current && (r.Condition == nil || r.Condition(w, e, s)) {
			return r, true
		}
	}
	return nil, false
}

// noopNode 未注册状态的兜底节点，避免 nil 调用（内部实现细节，非公共 API）。
type noopNode struct{}

func (noopNode) OnEnter(w World, e EntityID, s *State)              {}
func (noopNode) OnUpdate(w World, e EntityID, s *State, dt float32) {}
func (noopNode) OnExit(w World, e EntityID, s *State)               {}
