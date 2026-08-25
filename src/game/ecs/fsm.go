package ecs

// ============================================================
// 有限状态机（FSM）模块 —— 分层设计
//
// 底层：System —— StateMachineSystem
//    驱动“状态节点组件(State)”与“状态上下文组件(StateContext)”
//    （二者均为 Component），对每个拥有 State 的实体按帧执行
//    节点的 OnEnter/OnUpdate/OnExit。
// 中层：Component
//    状态节点组件（State）：记录当前/上一个状态、进入时间、持续时长；
//    状态上下文组件（StateContext）：存放状态机所需的额外判断/行为数据。
// 顶层：状态节点拓扑
//    StateMachine：StateID -> StateNode 注册表 + 转换规则（全局一份）；
//    StateNode 预留 OnEnter/OnUpdate/OnExit 接口，供开发者二次开发。
//
// 说明：业务状态（如站立/走动/跳跃等）不在框架内定义，
// 由使用方自行定义 StateID 常量，并在拓扑中注册节点与转换规则。
//
// StateMachineSystem 首次 Update 时会惰性解析 State 组件 ID（RegisterComponent 幂等），
// 因此无需手动预注册 State。
//
// 每帧执行策略（对每个拥有 State 组件的实体）：
//   1) 检查“退出此节点的标记”（即发生了转换）→ 执行 OnExit，并发送相关 Event；
//   2) 检查“进入此节点的标记”（转换已提交 / 实体首次运行）→ 执行 OnEnter，
//      并发送相关 Event；
//   3) 检查“运行标记” → 执行 OnUpdate，不发送 Event。
//
// “发送相关 Event”属于事件模块（core 包）职责，本模块只留占位注释。
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

	entered bool // 内部：该实体是否已进入过初始状态（用于首次触发 OnEnter）
}

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

// noopNode 未注册状态的兜底节点，避免 nil 调用。
type noopNode struct{}

func (noopNode) OnEnter(w World, e EntityID, s *State)              {}
func (noopNode) OnUpdate(w World, e EntityID, s *State, dt float32) {}
func (noopNode) OnExit(w World, e EntityID, s *State)               {}

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

// RegisterNode 注册 StateID 对应的状态节点。空缺档位用 noopNode 兜底。
func (m *StateMachine) RegisterNode(id StateID, n StateNode) {
	idx := int(id)
	for len(m.nodes) < idx {
		m.nodes = append(m.nodes, noopNode{})
	}
	m.nodes = append(m.nodes, n)
}

// Node 返回 StateID 对应的节点；未注册时返回 noopNode。
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

// StateMachineSystem 驱动状态机运转的 System。
//
// 由于节点回调（OnEnter/OnUpdate/OnExit）是使用者自定义的 OOP 代码，
// 很可能在回调中对“当前实体”增删组件（触发 Archetype 迁移），
// 因此本系统先收集“拥有 State 的实体快照”，再逐个处理，
// 避免在查询迭代过程中被自身回调破坏。
type StateMachineSystem struct {
	BaseSystem
	machine  *StateMachine
	stateID  ComponentID // 惰性解析的 State 组件 ID（首次 Update 时确定）
	resolved bool
	buf      []EntityID // 每帧复用的实体快照缓冲
}

// NewStateMachineSystem 创建一个驱动指定状态机的系统。
// 内嵌 BaseSystem，可直接 Add 到 SystemManager，并通过
// SetPriority / Enable / Disable 进行调度，例如：
//
//	fsm := ecs.NewStateMachineSystem(m)
//	fsm.SetPriority(50)
//	sm.Add(fsm)
func NewStateMachineSystem(m *StateMachine) *StateMachineSystem {
	return &StateMachineSystem{machine: m}
}

// SetMachine 替换/绑定状态机。
func (s *StateMachineSystem) SetMachine(m *StateMachine) { s.machine = m }

// getState 读取实体上的 State 组件指针；实体不存在或没有该组件时返回 nil。
func getState(w World, e EntityID, id ComponentID) *State {
	p := w.Get(e, id)
	if p == nil {
		return nil
	}
	return p.(*State)
}

// Update 实现 System 接口。
func (s *StateMachineSystem) Update(w World, dt float32) {
	if s.machine == nil {
		return
	}
	s.machine.elapsed += dt

	// 惰性解析 State 组件 ID（RegisterComponent 幂等：已注册则直接返回，无副作用）
	if !s.resolved {
		s.stateID = RegisterComponent[State](w)
		s.resolved = true
	}

	// 1) 收集本帧需要处理的实体快照（避免回调增删组件破坏迭代）
	s.buf = s.buf[:0]
	w.Query(Mask(s.stateID), func(e EntityID, q *Query) bool {
		s.buf = append(s.buf, e)
		return true
	})

	// 2) 逐实体驱动状态机
	for _, e := range s.buf {
		st := getState(w, e, s.stateID)
		if st == nil {
			continue // 实体可能已被其他系统销毁
		}

		// 首次进入：实体创建时已带 State，首帧触发一次 OnEnter
		if !st.entered {
			st.entered = true
			st.EnterTime = s.machine.elapsed
			s.machine.Node(st.Current).OnEnter(w, e, st)
			// TODO: 发送「进入状态」Event（占位，属事件模块职责）
			// 节点回调可能增删组件（触发 Archetype 迁移），重新获取 State 指针
			st = getState(w, e, s.stateID)
			if st == nil {
				continue // 实体可能已被回调销毁
			}
		}

		st.Duration += dt

		// 3) 转换检测：检查“退出当前节点的标记”（找到满足条件的转换）
		if r, ok := s.machine.Resolve(w, e, st); ok && r.To != st.Current {
			// 退出当前状态节点
			s.machine.Node(st.Current).OnExit(w, e, st)
			// TODO: 发送「退出状态」Event（占位，属事件模块职责）
			// OnExit 可能迁移/销毁实体，重新获取 State 指针
			st = getState(w, e, s.stateID)
			if st == nil {
				continue
			}

			// 提交转换
			st.Previous = st.Current
			st.Current = r.To
			st.Duration = 0
			st.EnterTime = s.machine.elapsed

			// 转换副作用回调（消费一次性请求组件）
			if r.Apply != nil {
				r.Apply(w, e, st)
				// Apply 通常移除请求组件（触发迁移），重新获取 State 指针
				st = getState(w, e, s.stateID)
				if st == nil {
					continue
				}
			}

			// 进入新状态节点
			s.machine.Node(st.Current).OnEnter(w, e, st)
			// TODO: 发送「进入状态」Event（占位，属事件模块职责）
			// OnEnter 可能迁移/销毁实体，重新获取 State 指针
			st = getState(w, e, s.stateID)
			if st == nil {
				continue
			}
		}

		// 4) 运行当前状态节点（每帧，不发送 Event）
		s.machine.Node(st.Current).OnUpdate(w, e, st, dt)
	}
}
