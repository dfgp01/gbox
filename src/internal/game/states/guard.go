package states

import def "gbox/def/game/states"

// FunctionalGuards adapts a function to the Guards contract.
type FunctionalGuards struct {
	Check def.Guard // Predicate evaluated against the current FSM context.
}

var _ def.Guards = FunctionalGuards{}

func (g FunctionalGuards) Evaluate(ctx def.FSMContext) (bool, []int) {
	if g.Check == nil {
		return true, nil
	}
	return g.Check(ctx), nil
}

// All succeeds when every child succeeds. An empty All is true.
type All struct {
	Subguards []def.Guards // Guards that must all pass.
}

var _ def.Guards = All{}

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

var _ def.Guards = Any{}

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

var _ def.Guards = Not{}

func (g Not) Evaluate(ctx def.FSMContext) (bool, []int) {
	if len(g.Subguards) != 1 || g.Subguards[0] == nil {
		return false, nil
	}
	passed, path := g.Subguards[0].Evaluate(ctx)
	return !passed, path
}
