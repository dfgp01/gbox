package ecs

import game "gbox/def/game"

// ============================================================
// 有限状态机（FSM）默认实现：StateMachineSystem
//
// 公共定义（StateID / State / StateContext / StateNode /
// TransitionRule / StateMachine）位于 def/game/fsm.go；
// 本文件只提供“驱动状态机运转”的默认 System 实现（internal）。
//
// 每帧执行策略（对每个拥有 State 组件的实体）：
//   1) 检查“退出此节点的标记”（即发生了转换）→ 执行 OnExit，并发送相关 Event；
//   2) 检查“进入此节点的标记”（转换已提交 / 实体首次运行）→ 执行 OnEnter，
//      并发送相关 Event；
//   3) 检查“运行标记” → 执行 OnUpdate，不发送 Event。
//
// StateMachineSystem 首次 Update 时会惰性解析 State 组件 ID
// （RegisterComponent 幂等），因此无需手动预注册 State。
//
// “发送相关 Event”属于事件模块（core 包）职责，本模块只留占位注释。
// ============================================================

// StateMachineSystem 驱动状态机运转的 System。
//
// 由于节点回调（OnEnter/OnUpdate/OnExit）是使用者自定义的 OOP 代码，
// 很可能在回调中对“当前实体”增删组件（触发 Archetype 迁移），
// 因此本系统先收集“拥有 State 的实体快照”，再逐个处理，
// 避免在查询迭代过程中被自身回调破坏。
type StateMachineSystem struct {
	priority int  // 优先级：越大越先执行（同优先级保持注册顺序）
	disable  bool // 软开关
	machine  *game.StateMachine
	stateID  game.ComponentID // 惰性解析的 State 组件 ID（首次 Update 时确定）
	resolved bool
	buf      []game.EntityID // 每帧复用的实体快照缓冲
}

// Priority 返回优先级（越大越先执行）。
func (s *StateMachineSystem) Priority() int { return s.priority }

// SetPriority 设置优先级。
func (s *StateMachineSystem) SetPriority(p int) { s.priority = p }

// Enable 启用系统（运行期软开关）。
func (s *StateMachineSystem) Enable() { s.disable = false }

// Disable 禁用系统（运行期软开关）。
func (s *StateMachineSystem) Disable() { s.disable = true }

// Enabled 系统是否处于启用状态。
func (s *StateMachineSystem) Enabled() bool { return !s.disable }

// NewStateMachineSystem 创建一个驱动指定状态机的系统。
// 自带优先级与软开关，可直接 Add 到系统组（pkg/game.SystemGroup），并通过
// SetPriority / Enable / Disable 进行调度，例如：
//
//	fsm := game.NewStateMachineSystem(m)
//	fsm.SetPriority(50)
//	group.Add(fsm)
func NewStateMachineSystem(m *game.StateMachine) *StateMachineSystem {
	return &StateMachineSystem{machine: m}
}

// SetMachine 替换/绑定状态机。
func (s *StateMachineSystem) SetMachine(m *game.StateMachine) { s.machine = m }

// getState 读取实体上的 State 组件指针；实体不存在或没有该组件时返回 nil。
func getState(w game.World, e game.EntityID, id game.ComponentID) *game.State {
	p := w.Get(e, id)
	if p == nil {
		return nil
	}
	return p.(*game.State)
}

// Update 实现 System 接口。
func (s *StateMachineSystem) Update(w game.World, dt float32) {
	if s.machine == nil {
		return
	}
	s.machine.Tick(dt)

	// 惰性解析 State 组件 ID（RegisterComponent 幂等：已注册则直接返回，无副作用）。
	// def/game.World 不暴露组件注册（统一走 pkg/game.ComponentMapper），
	// 内部实现 world 仍提供 RegisterComponent，通过内部契约取用。
	if !s.resolved {
		if reg, ok := w.(componentRegistrar); ok {
			s.stateID = reg.RegisterComponent(game.State{})
			s.resolved = true
		} else {
			panic("ecs: StateMachineSystem 需要支持组件注册的 World（请用 pkg/game.NewWorld / NewWorldFrom 创建）")
		}
	}

	// 1) 收集本帧需要处理的实体快照（避免回调增删组件破坏迭代）
	s.buf = s.buf[:0]
	w.Query(game.Mask(s.stateID), func(e game.EntityID, q game.Query) bool {
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
		if !st.Entered() {
			st.MarkEntered()
			st.EnterTime = s.machine.Elapsed()
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
			st.EnterTime = s.machine.Elapsed()

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
