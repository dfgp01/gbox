package states

import (
	"errors"
	"fmt"
	"reflect"
	"sort"

	def "gbox/def/game/states"
	internal "gbox/internal/game/states"
)

// Node types are implemented internally; aliases preserve the pkg API.
type BaseNode = internal.BaseNode
type LeafNode = internal.LeafNode
type SerialNode = internal.SerialNode
type ParallelNode = internal.ParallelNode
type SelectorNode = internal.SelectorNode

func NewLeafNode(id def.StateID, name string) *LeafNode { return internal.NewLeafNode(id, name) }

func NewSerialNode(id def.StateID, name string, children ...def.Node) *SerialNode {
	return internal.NewSerialNode(id, name, children...)
}

func NewParallelNode(id def.StateID, name string, children ...def.Node) *ParallelNode {
	return internal.NewParallelNode(id, name, children...)
}

func NewSelectorNode(id def.StateID, name string, children ...def.Node) *SelectorNode {
	return internal.NewSelectorNode(id, name, children...)
}

// FunctionalGuards adapts a function to the Guards contract.
type FunctionalGuards struct {
	Check def.Guard // Predicate evaluated against the current FSM context.
}

func (g FunctionalGuards) Evaluate(ctx def.FSMContext) (bool, []int) {
	if g.Check == nil {
		return true, nil
	}
	return g.Check(ctx), nil
}

func (FunctionalGuards) Children() []def.Guards { return nil }

// All succeeds when every child succeeds. An empty All is true.
type All struct {
	Subguards []def.Guards // Guards that must all pass.
}

func (g All) Children() []def.Guards { return append([]def.Guards(nil), g.Subguards...) }
func (g All) Evaluate(ctx def.FSMContext) (bool, []int) {
	for index, child := range g.Subguards {
		if child == nil {
			return false, []int{index}
		}
		passed, path := child.Evaluate(ctx)
		if !passed {
			return false, append([]int{index}, path...)
		}
	}
	return true, nil
}

// Any succeeds when at least one child succeeds. An empty Any is false.
type Any struct {
	Subguards []def.Guards // Guards searched until one passes.
}

func (g Any) Children() []def.Guards { return append([]def.Guards(nil), g.Subguards...) }
func (g Any) Evaluate(ctx def.FSMContext) (bool, []int) {
	for index, child := range g.Subguards {
		if child == nil {
			continue
		}
		passed, path := child.Evaluate(ctx)
		if passed {
			return true, append([]int{index}, path...)
		}
	}
	return false, nil
}

// Not negates its only child. Invalid arity evaluates to false.
type Not struct {
	Subguards []def.Guards // Exactly one guard whose result is negated.
}

func (g Not) Children() []def.Guards { return append([]def.Guards(nil), g.Subguards...) }
func (g Not) Evaluate(ctx def.FSMContext) (bool, []int) {
	if len(g.Subguards) != 1 || g.Subguards[0] == nil {
		return false, nil
	}
	passed, path := g.Subguards[0].Evaluate(ctx)
	return !passed, path
}

// Graph is an immutable, validated state graph shared by FSM instances.
type Graph struct {
	root  def.Node                 // Entry node used to define this graph.
	nodes map[def.StateID]def.Node // Validated node lookup shared by FSM instances.
}

func NewGraph(root def.Node) (*Graph, error) {
	if root == nil {
		return nil, errors.New("states: root node is nil")
	}
	graph := &Graph{root: root, nodes: make(map[def.StateID]def.Node)}
	stack := []def.Node{root}
	for len(stack) > 0 {
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if node == nil || isNilNode(node) {
			return nil, errors.New("states: graph contains a nil node")
		}
		id := node.ID()
		if id == def.InvalidStateID {
			return nil, fmt.Errorf("states: node %q has invalid ID", node.Name())
		}
		if existing, ok := graph.nodes[id]; ok {
			if !sameNode(existing, node) {
				return nil, fmt.Errorf("states: duplicate node ID %d", id)
			}
			continue
		}
		if node.Name() == "" {
			return nil, fmt.Errorf("states: node %d has an empty name", id)
		}
		if len(node.Children()) > 0 {
			if _, ok := node.(childScheduler); !ok {
				return nil, fmt.Errorf("states: node %q has children but no child scheduler", node.Name())
			}
		}
		graph.nodes[id] = node

		children := node.Children()
		names := make(map[string]struct{}, len(children))
		for _, child := range children {
			if child == nil {
				return nil, fmt.Errorf("states: node %q has a nil child", node.Name())
			}
			if _, ok := names[child.Name()]; ok {
				return nil, fmt.Errorf("states: duplicate child name %q under %q", child.Name(), node.Name())
			}
			names[child.Name()] = struct{}{}
			stack = append(stack, child)
		}
		if target := node.DefaultTo(); target != nil {
			stack = append(stack, target)
		}
		for _, edge := range node.Transitions() {
			if edge.To == nil {
				return nil, fmt.Errorf("states: node %q has a transition with no target", node.Name())
			}
			stack = append(stack, edge.To)
		}
	}
	return graph, nil
}

func (g *Graph) Root() def.Node { return g.root }

func (g *Graph) Node(id def.StateID) (def.Node, bool) {
	node, ok := g.nodes[id]
	return node, ok
}

func sameNode(left, right def.Node) bool {
	leftValue := reflect.ValueOf(left)
	rightValue := reflect.ValueOf(right)
	if leftValue.Type() != rightValue.Type() {
		return false
	}
	if leftValue.Type().Comparable() {
		return leftValue.Interface() == rightValue.Interface()
	}
	if leftValue.Kind() == reflect.Pointer {
		return leftValue.Pointer() == rightValue.Pointer()
	}
	return false
}

func isNilNode(node def.Node) bool {
	value := reflect.ValueOf(node)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

// FSM holds mutable state for one player or unit. Graph definitions are shared.
type FSM struct {
	Current   def.StateID // ID of the active node.
	Previous  def.StateID // ID of the node exited by the last transition.
	EnterTime float32     // State-machine clock time when Current was entered.
	Duration  float32     // Time spent in Current.

	graph      *Graph                       // Shared immutable graph definition.
	ctx        def.FSMContext               // Mutable per-instance callback context.
	elapsed    float32                      // Accumulated state-machine clock time.
	entered    bool                         // Whether the current node has received OnEnter.
	runtime    map[def.StateID]*nodeRuntime // Per-node composite scheduling state.
	branchPath []int                        // Branch path selected during the current tick.
	proposals  []transitionProposal         // Transition requests awaiting arbitration this tick.
}

type nodeRuntime = internal.NodeRuntime

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
		runtime: make(map[def.StateID]*nodeRuntime),
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
			f.ctx = proposal.action(f.ctx)
		}
		f.Previous = f.Current
		f.Current = proposal.target.ID()
		f.Duration = 0
		f.EnterTime = f.elapsed
		f.ctx.Duration = 0
		f.runtime = make(map[def.StateID]*nodeRuntime)
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

func (f *FSM) collectTransitions(node def.Node, guardsPassed bool, guardPath []int) {
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
		passed, path := edge.Guards.Evaluate(f.ctx)
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
	return node.OnEnter(ctx)
}

func (f *FSM) exitNode(node def.Node, ctx def.FSMContext) def.FSMContext {
	if runtime := f.runtime[node.ID()]; runtime != nil {
		for index, child := range node.Children() {
			if runtime.ChildStarted(index) {
				ctx = f.exitNode(child, ctx)
				runtime.SetChildStarted(index, false)
			}
		}
	}
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
			f.collectTransitions(node, false, guardPath)
			f.branchPath = append([]int(nil), guardPath...)
			return ctx, def.Failure, false
		}
	}
	var status def.Status
	ctx, status = node.OnTick(ctx)
	for _, addon := range node.Addons() {
		if addon != nil {
			ctx = addon(ctx)
		}
	}
	f.collectTransitions(node, true, guardPath)
	if len(guardPath) > 0 {
		f.branchPath = append([]int(nil), guardPath...)
	}
	if len(node.Children()) == 0 {
		return ctx, status, true
	}
	ctx, status = f.tickChildren(node, ctx)
	return ctx, status, true
}

type childScheduler interface {
	internal.ChildScheduler
}

func (f *FSM) tickChildren(node def.Node, ctx def.FSMContext) (def.FSMContext, def.Status) {
	children := node.Children()
	runtime := f.runtime[node.ID()]
	if runtime == nil {
		runtime = internal.NewNodeRuntime(len(children))
		f.runtime[node.ID()] = runtime
	}
	scheduler, ok := node.(childScheduler)
	if !ok {
		return ctx, def.Failure
	}
	return scheduler.ScheduleChildren(runtime, f.tickNode, f.startChild, f.finishChild, ctx)
}

func (f *FSM) startChild(node def.Node, runtime *nodeRuntime, index int, ctx def.FSMContext) def.FSMContext {
	if !runtime.ChildStarted(index) {
		ctx = f.enterNode(node, ctx)
		runtime.SetChildStarted(index, true)
	}
	return ctx
}

func (f *FSM) finishChild(node def.Node, runtime *nodeRuntime, index int, ctx def.FSMContext) def.FSMContext {
	if runtime.ChildStarted(index) {
		ctx = f.exitNode(node, ctx)
		runtime.SetChildStarted(index, false)
	}
	return ctx
}
