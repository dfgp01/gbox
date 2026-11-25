package states

import def "gbox/def/game/states"

// NodeRuntime holds mutable child scheduling data for one FSM instance.
type NodeRuntime struct {
	active   int
	started  []bool
	finished []bool
	failed   bool
}

// NewNodeRuntime allocates per-instance scheduling data for a composite node.
func NewNodeRuntime(childCount int) *NodeRuntime {
	return &NodeRuntime{
		started:  make([]bool, childCount),
		finished: make([]bool, childCount),
	}
}

// ChildStarted reports whether a child has entered and not yet exited.
func (r *NodeRuntime) ChildStarted(index int) bool {
	return index >= 0 && index < len(r.started) && r.started[index]
}

// SetChildStarted updates a child's lifecycle state.
func (r *NodeRuntime) SetChildStarted(index int, started bool) {
	if index >= 0 && index < len(r.started) {
		r.started[index] = started
	}
}

// ChildTickFunc ticks a scheduled child node.
type ChildTickFunc func(def.Node, def.FSMContext) (def.FSMContext, def.Status)

// ChildLifecycleFunc invokes a child lifecycle callback and updates the context.
type ChildLifecycleFunc func(def.Node, *NodeRuntime, int, def.FSMContext) def.FSMContext

// ChildScheduler is implemented by each concrete hierarchical node type.
type ChildScheduler interface {
	ScheduleChildren(*NodeRuntime, ChildTickFunc, ChildLifecycleFunc, ChildLifecycleFunc, def.FSMContext) (def.FSMContext, def.Status)
}

// SerialNode advances to the next child after each successful child.
type SerialNode struct {
	LeafNode // Common node structure and default lifecycle behavior.
}

var _ def.Node = (*SerialNode)(nil)
var _ ChildScheduler = (*SerialNode)(nil)

func NewSerialNode(id def.StateID, name string, children ...def.Node) *SerialNode {
	return &SerialNode{LeafNode: LeafNode{BaseNode: BaseNode{
		NodeID: id, NodeName: name, NodeChildren: append([]def.Node(nil), children...),
	}}}
}

func (n *SerialNode) ScheduleChildren(runtime *NodeRuntime, tick ChildTickFunc, start, finish ChildLifecycleFunc, ctx def.FSMContext) (def.FSMContext, def.Status) {
	children := n.Children()
	if runtime.active >= len(children) {
		return ctx, def.Success
	}
	index := runtime.active
	ctx = start(children[index], runtime, index, ctx)
	ctx, status := tick(children[index], ctx)
	if status == def.Running {
		return ctx, def.Running
	}
	ctx = finish(children[index], runtime, index, ctx)
	if status == def.Failure {
		return ctx, def.Failure
	}
	runtime.active++
	if runtime.active == len(children) {
		return ctx, def.Success
	}
	return ctx, def.Running
}

// SelectorNode advances after failures and succeeds on the first successful child.
type SelectorNode struct {
	LeafNode // Common node structure and default lifecycle behavior.
}

var _ def.Node = (*SelectorNode)(nil)
var _ ChildScheduler = (*SelectorNode)(nil)

func NewSelectorNode(id def.StateID, name string, children ...def.Node) *SelectorNode {
	return &SelectorNode{LeafNode: LeafNode{BaseNode: BaseNode{
		NodeID: id, NodeName: name, NodeChildren: append([]def.Node(nil), children...),
	}}}
}

func (n *SelectorNode) ScheduleChildren(runtime *NodeRuntime, tick ChildTickFunc, start, finish ChildLifecycleFunc, ctx def.FSMContext) (def.FSMContext, def.Status) {
	children := n.Children()
	if runtime.active >= len(children) {
		return ctx, def.Failure
	}
	index := runtime.active
	ctx = start(children[index], runtime, index, ctx)
	ctx, status := tick(children[index], ctx)
	if status == def.Running {
		return ctx, def.Running
	}
	ctx = finish(children[index], runtime, index, ctx)
	if status == def.Success {
		return ctx, def.Success
	}
	runtime.active++
	if runtime.active == len(children) {
		return ctx, def.Failure
	}
	return ctx, def.Running
}

// ParallelNode ticks all unfinished children in declaration order.
type ParallelNode struct {
	LeafNode // Common node structure and default lifecycle behavior.
}

var _ def.Node = (*ParallelNode)(nil)
var _ ChildScheduler = (*ParallelNode)(nil)

func NewParallelNode(id def.StateID, name string, children ...def.Node) *ParallelNode {
	return &ParallelNode{LeafNode: LeafNode{BaseNode: BaseNode{
		NodeID: id, NodeName: name, NodeChildren: append([]def.Node(nil), children...),
	}}}
}

func (n *ParallelNode) ScheduleChildren(runtime *NodeRuntime, tick ChildTickFunc, start, finish ChildLifecycleFunc, ctx def.FSMContext) (def.FSMContext, def.Status) {
	children := n.Children()
	if runtime.failed {
		return ctx, def.Failure
	}
	allFinished := true
	for index, child := range children {
		if runtime.finished[index] {
			continue
		}
		allFinished = false
		ctx = start(child, runtime, index, ctx)
		var status def.Status
		ctx, status = tick(child, ctx)
		if status == def.Running {
			continue
		}
		ctx = finish(child, runtime, index, ctx)
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

func allTrue(values []bool) bool {
	for _, value := range values {
		if !value {
			return false
		}
	}
	return true
}
