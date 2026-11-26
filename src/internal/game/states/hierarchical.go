package states

import def "gbox/def/game/states"

// childRuntime stores mutable scheduling data per FSM, never on a shared node.
type childRuntime struct {
	active   int
	started  []bool
	finished []bool
	failed   bool
}

func newChildRuntime(count int) *childRuntime {
	return &childRuntime{started: make([]bool, count), finished: make([]bool, count)}
}

// SerialNode advances to the next child after each successful child.
type SerialNode struct {
	nodeBase
	children []def.Node
}

var _ def.Node = (*SerialNode)(nil)

func NewSerialNode(id def.StateID, name string, children ...def.Node) *SerialNode {
	return &SerialNode{
		nodeBase: nodeBase{NodeID: id, NodeName: name},
		children: append([]def.Node(nil), children...),
	}
}

func (n *SerialNode) Type() def.NodeType   { return def.Serial }
func (n *SerialNode) Addons() []def.Action { return nil }
func (n *SerialNode) Children() []def.Node { return append([]def.Node(nil), n.children...) }

func (n *SerialNode) OnEnter(ctx def.FSMContext) def.FSMContext {
	if ctx.Execution != nil {
		ctx.Execution.SetState(n.ID(), newChildRuntime(len(n.children)))
	}
	return ctx
}

func (n *SerialNode) OnTick(ctx def.FSMContext) (def.FSMContext, def.Status) {
	runtime, ok := childState(ctx, n.ID())
	if !ok || ctx.Execution == nil {
		return ctx, def.Failure
	}
	if runtime.active >= len(n.children) {
		return ctx, def.Success
	}
	index := runtime.active
	child := n.children[index]
	ctx = startChild(ctx, runtime, child, index)
	ctx, status := ctx.Execution.TickNode(child, ctx)
	if status == def.Running {
		return ctx, def.Running
	}
	ctx = finishChild(ctx, runtime, child, index)
	if status == def.Failure {
		return ctx, def.Failure
	}
	runtime.active++
	if runtime.active == len(n.children) {
		return ctx, def.Success
	}
	return ctx, def.Running
}

func (n *SerialNode) OnExit(ctx def.FSMContext) def.FSMContext {
	return exitStartedChildren(ctx, n.ID(), n.children)
}

// SelectorNode advances after failures and succeeds on the first successful child.
type SelectorNode struct {
	nodeBase
	children []def.Node
}

var _ def.Node = (*SelectorNode)(nil)

func NewSelectorNode(id def.StateID, name string, children ...def.Node) *SelectorNode {
	return &SelectorNode{
		nodeBase: nodeBase{NodeID: id, NodeName: name},
		children: append([]def.Node(nil), children...),
	}
}

func (n *SelectorNode) Type() def.NodeType   { return def.Selector }
func (n *SelectorNode) Addons() []def.Action { return nil }
func (n *SelectorNode) Children() []def.Node { return append([]def.Node(nil), n.children...) }

func (n *SelectorNode) OnEnter(ctx def.FSMContext) def.FSMContext {
	if ctx.Execution != nil {
		ctx.Execution.SetState(n.ID(), newChildRuntime(len(n.children)))
	}
	return ctx
}

func (n *SelectorNode) OnTick(ctx def.FSMContext) (def.FSMContext, def.Status) {
	runtime, ok := childState(ctx, n.ID())
	if !ok || ctx.Execution == nil {
		return ctx, def.Failure
	}
	if runtime.active >= len(n.children) {
		return ctx, def.Failure
	}
	index := runtime.active
	child := n.children[index]
	ctx = startChild(ctx, runtime, child, index)
	ctx, status := ctx.Execution.TickNode(child, ctx)
	if status == def.Running {
		return ctx, def.Running
	}
	ctx = finishChild(ctx, runtime, child, index)
	if status == def.Success {
		return ctx, def.Success
	}
	runtime.active++
	if runtime.active == len(n.children) {
		return ctx, def.Failure
	}
	return ctx, def.Running
}

func (n *SelectorNode) OnExit(ctx def.FSMContext) def.FSMContext {
	return exitStartedChildren(ctx, n.ID(), n.children)
}

// ParallelNode ticks all unfinished children in declaration order.
type ParallelNode struct {
	nodeBase
	children []def.Node
}

var _ def.Node = (*ParallelNode)(nil)

func NewParallelNode(id def.StateID, name string, children ...def.Node) *ParallelNode {
	return &ParallelNode{
		nodeBase: nodeBase{NodeID: id, NodeName: name},
		children: append([]def.Node(nil), children...),
	}
}

func (n *ParallelNode) Type() def.NodeType   { return def.Parallel }
func (n *ParallelNode) Addons() []def.Action { return nil }
func (n *ParallelNode) Children() []def.Node { return append([]def.Node(nil), n.children...) }

func (n *ParallelNode) OnEnter(ctx def.FSMContext) def.FSMContext {
	if ctx.Execution != nil {
		ctx.Execution.SetState(n.ID(), newChildRuntime(len(n.children)))
	}
	return ctx
}

func (n *ParallelNode) OnTick(ctx def.FSMContext) (def.FSMContext, def.Status) {
	runtime, ok := childState(ctx, n.ID())
	if !ok || ctx.Execution == nil {
		return ctx, def.Failure
	}
	if runtime.failed {
		return ctx, def.Failure
	}
	allFinished := true
	for index, child := range n.children {
		if runtime.finished[index] {
			continue
		}
		allFinished = false
		ctx = startChild(ctx, runtime, child, index)
		var status def.Status
		ctx, status = ctx.Execution.TickNode(child, ctx)
		if status == def.Running {
			continue
		}
		ctx = finishChild(ctx, runtime, child, index)
		runtime.finished[index] = true
		if status == def.Failure {
			runtime.failed = true
			return ctx, def.Failure
		}
	}
	if allFinished || allTrue(runtime.finished) {
		return ctx, def.Success
	}
	return ctx, def.Running
}

func (n *ParallelNode) OnExit(ctx def.FSMContext) def.FSMContext {
	return exitStartedChildren(ctx, n.ID(), n.children)
}

func childState(ctx def.FSMContext, id def.StateID) (*childRuntime, bool) {
	if ctx.Execution == nil {
		return nil, false
	}
	value, ok := ctx.Execution.State(id)
	if !ok {
		return nil, false
	}
	runtime, ok := value.(*childRuntime)
	return runtime, ok
}

func startChild(ctx def.FSMContext, runtime *childRuntime, child def.Node, index int) def.FSMContext {
	if !runtime.started[index] {
		ctx = ctx.Execution.EnterNode(child, ctx)
		runtime.started[index] = true
	}
	return ctx
}

func finishChild(ctx def.FSMContext, runtime *childRuntime, child def.Node, index int) def.FSMContext {
	if runtime.started[index] {
		ctx = ctx.Execution.ExitNode(child, ctx)
		runtime.started[index] = false
	}
	return ctx
}

func exitStartedChildren(ctx def.FSMContext, id def.StateID, children []def.Node) def.FSMContext {
	runtime, ok := childState(ctx, id)
	if !ok || ctx.Execution == nil {
		return ctx
	}
	for index, child := range children {
		if runtime.started[index] {
			ctx = ctx.Execution.ExitNode(child, ctx)
			runtime.started[index] = false
		}
	}
	ctx.Execution.SetState(id, nil)
	return ctx
}

func allTrue(values []bool) bool {
	for _, value := range values {
		if !value {
			return false
		}
	}
	return true
}
