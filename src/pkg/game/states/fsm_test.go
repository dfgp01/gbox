package states

import (
	"testing"

	def "gbox/def/game/states"
)

func TestGuardCombinatorsReportBranches(t *testing.T) {
	ctx := def.FSMContext{Data: 3}
	guards := Any{Subguards: []def.Guards{
		FunctionalGuards{Check: func(def.FSMContext) bool { return false }},
		Any{Subguards: []def.Guards{
			FunctionalGuards{Check: func(def.FSMContext) bool { return false }},
			FunctionalGuards{Check: func(ctx def.FSMContext) bool { return ctx.Data.(int) == 3 }},
		}},
	}}
	passed, path := guards.Evaluate(ctx)
	if !passed || len(path) != 2 || path[0] != 1 || path[1] != 1 {
		t.Fatalf("Evaluate() = (%v, %v), want true and branch path [1 1]", passed, path)
	}

	if passed, _ := (All{}).Evaluate(ctx); !passed {
		t.Fatal("empty All must pass")
	}
	if passed, _ := (Any{}).Evaluate(ctx); passed {
		t.Fatal("empty Any must fail")
	}
	if passed, _ := (Not{}).Evaluate(ctx); passed {
		t.Fatal("Not with invalid arity must fail")
	}
}

func TestTickCommitsAtMostOneTransition(t *testing.T) {
	third := NewLeafNode(3, "third")
	second := NewLeafNode(2, "second")
	second.Edges = []def.Transition{{To: third}}
	first := NewLeafNode(1, "first")
	first.Edges = []def.Transition{
		{Weight: 1, Guards: FunctionalGuards{Check: func(def.FSMContext) bool { return true }}, To: third},
		{Weight: 10, Guards: FunctionalGuards{Check: func(def.FSMContext) bool { return true }}, To: second},
	}
	graph, err := NewGraph(first)
	if err != nil {
		t.Fatal(err)
	}
	fsm, err := NewFSM(graph, first.ID(), nil)
	if err != nil {
		t.Fatal(err)
	}
	result, err := fsm.Tick(1)
	if err != nil {
		t.Fatal(err)
	}
	if result.Current != second.ID() || !result.Transitioned {
		t.Fatalf("first tick = %+v, want transition to second", result)
	}
	result, err = fsm.Tick(1)
	if err != nil {
		t.Fatal(err)
	}
	if result.Current != third.ID() || !result.Transitioned {
		t.Fatalf("second tick = %+v, want transition to third", result)
	}
}

func TestNodeGuardFailureUsesDefaultButEdgeGuardFailureDoesNot(t *testing.T) {
	defaultNode := NewLeafNode(3, "default")
	blocked := NewLeafNode(2, "blocked")
	start := NewLeafNode(1, "start")
	start.NodeGuards = FunctionalGuards{Check: func(def.FSMContext) bool { return false }}
	start.FailTo = defaultNode
	start.Edges = []def.Transition{{
		Guards: FunctionalGuards{Check: func(def.FSMContext) bool { return false }},
		To:     blocked,
	}}
	graph, err := NewGraph(start)
	if err != nil {
		t.Fatal(err)
	}
	fsm, err := NewFSM(graph, start.ID(), nil)
	if err != nil {
		t.Fatal(err)
	}
	result, err := fsm.Tick(0.25)
	if err != nil {
		t.Fatal(err)
	}
	if result.Current != defaultNode.ID() || !result.Transitioned {
		t.Fatalf("node guard failure = %+v, want default transition", result)
	}

	plainStart := NewLeafNode(4, "plain-start")
	plainStart.Edges = []def.Transition{{
		Guards: FunctionalGuards{Check: func(def.FSMContext) bool { return false }},
		To:     blocked,
	}}
	plainGraph, err := NewGraph(plainStart)
	if err != nil {
		t.Fatal(err)
	}
	fsm, err = NewFSM(plainGraph, plainStart.ID(), nil)
	if err != nil {
		t.Fatal(err)
	}
	result, err = fsm.Tick(0.25)
	if err != nil {
		t.Fatal(err)
	}
	if result.Current != plainStart.ID() || result.Transitioned {
		t.Fatalf("edge guard failure = %+v, want no transition", result)
	}
}

func TestTerminalNodeCompletes(t *testing.T) {
	terminal := NewLeafNode(1, "terminal")
	graph, err := NewGraph(terminal)
	if err != nil {
		t.Fatal(err)
	}
	fsm, err := NewFSM(graph, terminal.ID(), nil)
	if err != nil {
		t.Fatal(err)
	}
	result, err := fsm.Tick(0)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != def.Success {
		t.Fatalf("terminal status = %v, want Success", result.Status)
	}
}

func TestHierarchicalSchedulers(t *testing.T) {
	t.Run("serial", func(t *testing.T) {
		first := NewLeafNode(2, "first")
		first.Tick = func(ctx def.FSMContext) (def.FSMContext, def.Status) { return ctx, def.Success }
		second := NewLeafNode(3, "second")
		second.Tick = func(ctx def.FSMContext) (def.FSMContext, def.Status) { return ctx, def.Success }
		parent := NewSerialNode(1, "sequence", first, second)
		fsm := newTestFSM(t, parent)
		result, err := fsm.Tick(0)
		if err != nil || result.Status != def.Running {
			t.Fatalf("first tick = %+v, %v; want Running", result, err)
		}
		result, err = fsm.Tick(0)
		if err != nil || result.Status != def.Success {
			t.Fatalf("second tick = %+v, %v; want Success", result, err)
		}
	})

	t.Run("selector", func(t *testing.T) {
		first := NewLeafNode(2, "fails")
		first.Tick = func(ctx def.FSMContext) (def.FSMContext, def.Status) { return ctx, def.Failure }
		second := NewLeafNode(3, "succeeds")
		second.Tick = func(ctx def.FSMContext) (def.FSMContext, def.Status) { return ctx, def.Success }
		parent := NewSelectorNode(1, "selector", first, second)
		fsm := newTestFSM(t, parent)
		result, err := fsm.Tick(0)
		if err != nil || result.Status != def.Running {
			t.Fatalf("first tick = %+v, %v; want Running", result, err)
		}
		result, err = fsm.Tick(0)
		if err != nil || result.Status != def.Success {
			t.Fatalf("second tick = %+v, %v; want Success", result, err)
		}
	})

	t.Run("parallel", func(t *testing.T) {
		first := NewLeafNode(2, "first")
		first.Tick = func(ctx def.FSMContext) (def.FSMContext, def.Status) { return ctx, def.Running }
		second := NewLeafNode(3, "second")
		second.Tick = func(ctx def.FSMContext) (def.FSMContext, def.Status) { return ctx, def.Running }
		low := NewLeafNode(4, "low-weight-target")
		high := NewLeafNode(5, "high-weight-target")
		first.Edges = []def.Transition{{
			Weight: 5,
			Guards: FunctionalGuards{Check: func(def.FSMContext) bool { return true }},
			To:     low,
		}}
		second.Edges = []def.Transition{{
			Weight: 9,
			Guards: FunctionalGuards{Check: func(def.FSMContext) bool { return true }},
			To:     high,
		}}
		parent := NewParallelNode(1, "parallel", first, second)
		fsm := newTestFSM(t, parent)
		result, err := fsm.Tick(0)
		if err != nil || result.Current != high.ID() || !result.Transitioned {
			t.Fatalf("tick = %+v, %v; want one transition to the higher-weight target", result, err)
		}
	})
}

func newTestFSM(t *testing.T, root def.Node) *FSM {
	t.Helper()
	graph, err := NewGraph(root)
	if err != nil {
		t.Fatal(err)
	}
	fsm, err := NewFSM(graph, root.ID(), nil)
	if err != nil {
		t.Fatal(err)
	}
	return fsm
}

func TestSharedGraphKeepsInstanceContextsSeparate(t *testing.T) {
	first := NewLeafNode(1, "idle")
	first.Tick = func(ctx def.FSMContext) (def.FSMContext, def.Status) {
		data := ctx.Data.(int)
		ctx.Data = data + 1
		return ctx, def.Running
	}
	graph, err := NewGraph(first)
	if err != nil {
		t.Fatal(err)
	}
	left, err := NewFSM(graph, first.ID(), 10)
	if err != nil {
		t.Fatal(err)
	}
	right, err := NewFSM(graph, first.ID(), 20)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := left.Tick(0.1); err != nil {
		t.Fatal(err)
	}
	if _, err := right.Tick(0.1); err != nil {
		t.Fatal(err)
	}
	if left.Context().Data != 11 || right.Context().Data != 21 {
		t.Fatalf("contexts = %v and %v, want 11 and 21", left.Context().Data, right.Context().Data)
	}
}
