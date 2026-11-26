package states

import (
	"errors"
	"fmt"
	"sort"

	def "gbox/def/game/states"
	internal "gbox/internal/game/states"
)

// Node types are implemented internally; aliases preserve the public API.
type LeafNode = internal.LeafNode
type TerminalNode = internal.TerminalNode
type SerialNode = internal.SerialNode
type ParallelNode = internal.ParallelNode
type SelectorNode = internal.SelectorNode

func NewLeafNode(id def.StateID, name string) *LeafNode { return internal.NewLeafNode(id, name) }

func NewTerminalNode(id def.StateID, name string) *TerminalNode {
	return internal.NewTerminalNode(id, name)
}

func NewSerialNode(id def.StateID, name string, children ...def.Node) *SerialNode {
	return internal.NewSerialNode(id, name, children...)
}

func NewParallelNode(id def.StateID, name string, children ...def.Node) *ParallelNode {
	return internal.NewParallelNode(id, name, children...)
}

func NewSelectorNode(id def.StateID, name string, children ...def.Node) *SelectorNode {
	return internal.NewSelectorNode(id, name, children...)
}

// Guard types are implemented internally; aliases preserve the public API.
type FunctionalGuards = internal.FunctionalGuards
type All = internal.All
type Any = internal.Any
type Not = internal.Not

// FSM holds mutable state for one player or unit. Graph definitions are shared.
type FSM struct {
	Current   def.StateID // ID of the active node.
	Previous  def.StateID // ID of the node exited by the last transition.
	EnterTime float32     // State-machine clock time when Current was entered.
	Duration  float32     // Time spent in Current.

	graph      *Graph               // Shared immutable graph definition.
	ctx        def.FSMContext       // Mutable per-instance callback context.
	elapsed    float32              // Accumulated state-machine clock time.
	entered    bool                 // Whether the current node has received OnEnter.
	runtime    map[def.StateID]any  // Per-node composite scheduling state.
	branchPath []int                // Branch path selected during the current tick.
	proposals  []transitionProposal // Transition requests awaiting arbitration this tick.
}

type transitionProposal struct {
	weight    int        // Priority of the matching explicit edge.
	isDefault bool       // Whether this proposal came from a node's fallback target.
	target    def.Node   // Requested destination node.
	action    def.Action // Context transformation applied if this proposal wins.
	path      []int      // Guard branch path associated with this proposal.
}

func NewFSM(graph *Graph, initial def.StateID, data any) (*FSM, error) {
	if graph == nil {
		return nil, errors.New("states: graph is nil")
	}
	if _, ok := graph.Node(initial); !ok {
		return nil, fmt.Errorf("states: initial node %d is not in graph", initial)
	}
	return &FSM{
		Current: initial,
		graph:   graph,
		ctx:     def.FSMContext{Data: data},
		runtime: make(map[def.StateID]any),
	}, nil
}

func (f *FSM) Context() def.FSMContext { return f.ctx }

func (f *FSM) SetContext(ctx def.FSMContext) { f.ctx = ctx }

func (f *FSM) Tick(dt float32) (def.TickResult, error) {
	if dt < 0 {
		return def.TickResult{}, errors.New("states: dt must not be negative")
	}
	if f.graph == nil {
		return def.TickResult{}, errors.New("states: FSM has no graph")
	}
	node, ok := f.graph.Node(f.Current)
	if !ok {
		return def.TickResult{}, fmt.Errorf("states: current node %d is not in graph", f.Current)
	}

	f.elapsed += dt
	f.ctx.Time = f.elapsed
	f.ctx.Delta = dt
	f.ctx.Execution = f
	f.ctx.Events = nil
	f.ctx.BranchPath = nil
	f.branchPath = nil
	f.proposals = f.proposals[:0]
	if !f.entered {
		f.EnterTime = f.elapsed - dt
		f.ctx.Duration = 0
		f.ctx = f.enterNode(node, f.ctx)
		f.entered = true
	}
	f.Duration += dt
	f.ctx.Duration = f.Duration

	status := def.Running
	passed := true
	f.ctx, status, passed = f.tickNodeGuarded(node, f.ctx)
	proposal, transitioned := f.selectTransition()
	if transitioned {
		f.ctx = f.exitNode(node, f.ctx)
		if proposal.action != nil {
			f.ctx = proposal.action.OnEnter(f.ctx)
		}
		f.Previous = f.Current
		f.Current = proposal.target.ID()
		f.Duration = 0
		f.EnterTime = f.elapsed
		f.ctx.Duration = 0
		f.runtime = make(map[def.StateID]any)
		f.ctx = f.enterNode(proposal.target, f.ctx)
		f.entered = true
		f.branchPath = append([]int(nil), proposal.path...)
	} else if len(node.Transitions()) == 0 && node.DefaultTo() == nil && len(f.proposals) == 0 && passed {
		status = def.Success
	}
	f.ctx.BranchPath = append([]int(nil), f.branchPath...)

	result := def.TickResult{
		Current:      f.Current,
		Previous:     f.Previous,
		Status:       status,
		Transitioned: transitioned,
		Events:       append([]any(nil), f.ctx.Events...),
		BranchPath:   append([]int(nil), f.ctx.BranchPath...),
	}
	return result, nil
}

func (f *FSM) collectTransitions(node def.Node, guardsPassed bool, guardPath []int, ctx def.FSMContext) {
	if !guardsPassed {
		if target := node.DefaultTo(); target != nil && target.ID() != node.ID() {
			f.proposals = append(f.proposals, transitionProposal{
				isDefault: true,
				target:    target,
				path:      append([]int(nil), guardPath...),
			})
		}
		return
	}
	edges := node.Transitions()
	sort.SliceStable(edges, func(i, j int) bool { return edges[i].Weight > edges[j].Weight })
	for _, edge := range edges {
		if edge.Guards == nil {
			if edge.To.ID() != node.ID() {
				f.proposals = append(f.proposals, transitionProposal{
					weight: edge.Weight, target: edge.To, action: edge.Action,
				})
			}
			return
		}
		passed, path := edge.Guards.Evaluate(ctx)
		if passed {
			if edge.To.ID() != node.ID() {
				f.proposals = append(f.proposals, transitionProposal{
					weight: edge.Weight, target: edge.To, action: edge.Action,
					path: append([]int(nil), path...),
				})
			}
			return
		}
	}
}

func (f *FSM) selectTransition() (transitionProposal, bool) {
	if len(f.proposals) == 0 {
		return transitionProposal{}, false
	}
	selected := f.proposals[0]
	for _, proposal := range f.proposals[1:] {
		if selected.isDefault && !proposal.isDefault || selected.isDefault == proposal.isDefault && proposal.weight > selected.weight {
			selected = proposal
		}
	}
	return selected, true
}

func (f *FSM) enterNode(node def.Node, ctx def.FSMContext) def.FSMContext {
	ctx.Execution = f
	return node.OnEnter(ctx)
}

func (f *FSM) exitNode(node def.Node, ctx def.FSMContext) def.FSMContext {
	ctx.Execution = f
	return node.OnExit(ctx)
}

func (f *FSM) tickNode(node def.Node, ctx def.FSMContext) (def.FSMContext, def.Status) {
	ctx, status, _ := f.tickNodeGuarded(node, ctx)
	return ctx, status
}

func (f *FSM) tickNodeGuarded(node def.Node, ctx def.FSMContext) (def.FSMContext, def.Status, bool) {
	var guardPath []int
	if guards := node.Guards(); guards != nil {
		var passed bool
		passed, guardPath = guards.Evaluate(ctx)
		if !passed {
			f.collectTransitions(node, false, guardPath, ctx)
			f.branchPath = append([]int(nil), guardPath...)
			return ctx, def.Failure, false
		}
	}
	var status def.Status
	ctx, status = node.OnTick(ctx)
	for _, addon := range node.Addons() {
		if addon != nil {
			ctx, _ = addon.OnTick(ctx)
		}
	}
	f.collectTransitions(node, true, guardPath, ctx)
	if len(guardPath) > 0 {
		f.branchPath = append([]int(nil), guardPath...)
	}
	if len(node.Children()) == 0 {
		return ctx, status, true
	}
	return ctx, status, true
}

// EnterNode implements def.NodeExecution for composite node lifecycles.
func (f *FSM) EnterNode(node def.Node, ctx def.FSMContext) def.FSMContext {
	return f.enterNode(node, ctx)
}

// TickNode implements def.NodeExecution and evaluates a child without resolving a second transition.
func (f *FSM) TickNode(node def.Node, ctx def.FSMContext) (def.FSMContext, def.Status) {
	ctx.Execution = f
	ctx, status, _ := f.tickNodeGuarded(node, ctx)
	return ctx, status
}

// ExitNode implements def.NodeExecution for composite node lifecycles.
func (f *FSM) ExitNode(node def.Node, ctx def.FSMContext) def.FSMContext {
	return f.exitNode(node, ctx)
}

// State returns an opaque runtime value owned by this FSM and node ID.
func (f *FSM) State(id def.StateID) (any, bool) {
	state, ok := f.runtime[id]
	return state, ok
}

// SetState stores or removes opaque runtime data for a node.
func (f *FSM) SetState(id def.StateID, state any) {
	if state == nil {
		delete(f.runtime, id)
		return
	}
	f.runtime[id] = state
}

var _ def.NodeExecution = (*FSM)(nil)
