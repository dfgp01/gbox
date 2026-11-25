# Hierarchical State Machine

The reusable state graph and per-unit runtime are provided by:

- `gbox/def/game/states`: callback contracts and shared data types.
- `gbox/pkg/game/states`: public graph/FSM API and guard combinators.
- `gbox/internal/game/states`: node definitions and hierarchical scheduling implementations.

This package is independent from the ECS APIs and from the legacy FSM in `gbox/def/game`.

## Define a graph

Configure nodes before passing the root to `NewGraph`. Node definitions are
shared; do not capture mutable per-unit data in node behavior. Embed `LeafNode`
and override `OnEnter`, `OnTick`, `OnExit`, or `Addons` methods to customize a
leaf. Put per-unit data in `FSMContext.Data` instead.

```go
idle := states.NewLeafNode(1, "idle")
moving := states.NewLeafNode(2, "moving")

idle.Edges = []def.Transition{{
	Weight: 10,
	Guards: states.FunctionalGuards{Check: func(ctx def.FSMContext) bool {
		return ctx.Data.(PlayerContext).WantsToMove
	}},
	To: moving,
}}

graph, err := states.NewGraph(idle)
if err != nil {
	return err
}
playerFSM, err := states.NewFSM(graph, idle.ID(), PlayerContext{})
```

Outgoing transitions are checked by descending weight. Equal weights preserve
the node's declaration order. Guards should only inspect the context. Actions
transform and return a context; external side effects should be represented as
data or event markers for the caller to handle.

## Tick semantics

Call `FSM.Tick(dt)` once per frame. The current node enters once, then its
`OnTick` and addons run, followed by outgoing edge evaluation. At most one
transition is committed in a tick. The transition sequence is current `OnExit`,
edge action, state bookkeeping, then target `OnEnter`; the target is not ticked
until the next frame. A node-level guard failure selects `DefaultTo` when set.
If no edge guard passes, the FSM remains in its current node.

`FSMContext` is copied into callbacks and returned from them. Its `Events` field
is reset at the start of each tick; `Mark(event)` appends a marker for the caller
to process from `TickResult.Events`. `BranchPath` records the successful branch
selected by nested `Any` guards, or the failing branch for `All`.

## Guard combinators

`All` short-circuits on the first failure and is true when empty. `Any`
short-circuits on the first success and is false when empty. `Not` requires
exactly one child; invalid arity evaluates to false. `FunctionalGuards` wraps a
single pure guard function.

## Hierarchical nodes

`SerialNode`, `SelectorNode`, and `ParallelNode` each provide their own child
scheduling behavior; there is no scheduler-kind field on `Node`. Serial
advances after each successful child and fails on the first failed child.
Selector advances after failures and succeeds on the first successful child.
Parallel ticks each unfinished child in declaration order, succeeds when all
children succeed, and fails on the first failure. Child activation state is
stored by each FSM, not on the shared graph.

All node IDs must be nonzero and unique in the graph. Child names must be unique
among siblings. A graph should be treated as immutable after `NewGraph`.
During a hierarchical tick, each active node may contribute its first matching
outgoing transition; the FSM arbitrates those requests by weight and commits at
most one transition for the whole instance. Equal weights preserve traversal
order (parent before children, then child declaration order). A node's default
exit is a fallback and loses to any matching explicit edge.
When a child node's request wins, the transition target becomes the FSM's new
current node and the current composite tree is exited; this version does not
model local child-to-child transitions that preserve the composite parent.