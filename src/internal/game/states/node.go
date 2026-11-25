package states

import def "gbox/def/game/states"

// BaseNode stores the structural definition shared by leaf and composite nodes.
type BaseNode struct {
	NodeID       def.StateID      // Identifier unique across the graph.
	NodeName     string           // Name unique among this node's siblings.
	FailTo       def.Node         // Fallback target when node guards fail; nil means no fallback.
	NodeGuards   def.Guards       // Conditions that gate ticking this node.
	NodeChildren []def.Node       // Children scheduled by a composite node.
	Edges        []def.Transition // Outgoing transitions evaluated after this node ticks.
}

func (n *BaseNode) ID() def.StateID { return n.NodeID }

func (n *BaseNode) Name() string { return n.NodeName }

func (n *BaseNode) DefaultTo() def.Node { return n.FailTo }

func (n *BaseNode) Guards() def.Guards { return n.NodeGuards }

func (n *BaseNode) Children() []def.Node { return append([]def.Node(nil), n.NodeChildren...) }

func (n *BaseNode) Transitions() []def.Transition {
	return append([]def.Transition(nil), n.Edges...)
}

// LeafNode is a leaf definition with default lifecycle behavior.
// Embed it and override lifecycle methods to customize a leaf without storing callbacks.
type LeafNode struct {
	BaseNode // Structural definition shared with composite nodes.
}

var _ def.Node = (*LeafNode)(nil)

func NewLeafNode(id def.StateID, name string) *LeafNode {
	return &LeafNode{BaseNode: BaseNode{NodeID: id, NodeName: name}}
}

func (n *LeafNode) OnEnter(ctx def.FSMContext) def.FSMContext { return ctx }

func (n *LeafNode) OnTick(ctx def.FSMContext) (def.FSMContext, def.Status) {
	return ctx, def.Running
}

func (n *LeafNode) OnExit(ctx def.FSMContext) def.FSMContext { return ctx }

func (n *LeafNode) Addons() []def.Action { return nil }
