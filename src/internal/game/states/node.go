package states

import def "gbox/def/game/states"

// nodeBase contains only the structural fields shared by all node categories.
type nodeBase struct {
	NodeID     def.StateID
	NodeName   string
	NodeGuards def.Guards
}

func (n *nodeBase) ID() def.StateID { return n.NodeID }

func (n *nodeBase) Name() string { return n.NodeName }

func (n *nodeBase) Guards() def.Guards { return n.NodeGuards }

func (*nodeBase) Children() []def.Node { return nil }

func (*nodeBase) Transitions() []def.Transition { return nil }

func (*nodeBase) DefaultTo() def.Node { return nil }

// leafBehavior implements lifecycle behavior shared by ordinary and terminal leaves.
type leafBehavior struct {
	Actions    []def.Action
	NodeAddons []def.Action
}

func (b *leafBehavior) OnEnter(ctx def.FSMContext) def.FSMContext {
	for _, action := range b.Actions {
		if action != nil {
			ctx = action.OnEnter(ctx)
		}
	}
	return ctx
}

func (b *leafBehavior) OnTick(ctx def.FSMContext) (def.FSMContext, def.Status) {
	status := def.Running
	for _, action := range b.Actions {
		if action == nil {
			continue
		}
		var actionStatus def.Status
		ctx, actionStatus = action.OnTick(ctx)
		if actionStatus != def.Running {
			status = actionStatus
		}
	}
	return ctx, status
}

func (b *leafBehavior) OnExit(ctx def.FSMContext) def.FSMContext {
	for index := len(b.Actions) - 1; index >= 0; index-- {
		action := b.Actions[index]
		if action != nil {
			ctx = action.OnExit(ctx)
		}
	}
	return ctx
}

func (b *leafBehavior) Addons() []def.Action {
	return append([]def.Action(nil), b.NodeAddons...)
}

// LeafNode is an executable leaf with guarded outgoing transitions.
type LeafNode struct {
	nodeBase
	leafBehavior
	FailTo def.Node
	Edges  []def.Transition
}

var _ def.Node = (*LeafNode)(nil)

func NewLeafNode(id def.StateID, name string) *LeafNode {
	return &LeafNode{nodeBase: nodeBase{NodeID: id, NodeName: name}}
}

func (n *LeafNode) Type() def.NodeType { return def.Leaf }

func (n *LeafNode) DefaultTo() def.Node { return n.FailTo }

func (n *LeafNode) Transitions() []def.Transition {
	return append([]def.Transition(nil), n.Edges...)
}

func (n *LeafNode) OnEnter(ctx def.FSMContext) def.FSMContext {
	return n.leafBehavior.OnEnter(ctx)
}

func (n *LeafNode) OnTick(ctx def.FSMContext) (def.FSMContext, def.Status) {
	return n.leafBehavior.OnTick(ctx)
}

func (n *LeafNode) OnExit(ctx def.FSMContext) def.FSMContext {
	return n.leafBehavior.OnExit(ctx)
}

// TerminalNode is an executable leaf that cannot transition to another node.
type TerminalNode struct {
	nodeBase
	leafBehavior
}

var _ def.Node = (*TerminalNode)(nil)

func NewTerminalNode(id def.StateID, name string) *TerminalNode {
	return &TerminalNode{nodeBase: nodeBase{NodeID: id, NodeName: name}}
}

func (n *TerminalNode) Type() def.NodeType { return def.Terminal }

func (n *TerminalNode) OnEnter(ctx def.FSMContext) def.FSMContext {
	return n.leafBehavior.OnEnter(ctx)
}

func (n *TerminalNode) OnExit(ctx def.FSMContext) def.FSMContext {
	return n.leafBehavior.OnExit(ctx)
}

func (n *TerminalNode) OnTick(ctx def.FSMContext) (def.FSMContext, def.Status) {
	ctx, status := n.leafBehavior.OnTick(ctx)
	if status == def.Running {
		return ctx, def.Success
	}
	return ctx, status
}
