package states

// StateID identifies a node in a state graph. Zero is reserved as invalid.
type StateID int

const InvalidStateID StateID = 0

// Status describes the result of ticking a node.
type Status uint8

const (
	Running Status = iota
	Success
	Failure
)

// FSMContext contains instance-specific data passed through node callbacks.
// Callback code should treat Data as immutable and return a changed context.
type FSMContext struct {
	Data       any     // Per-instance application data shared with node callbacks.
	Time       float32 // State-machine clock time in seconds.
	Delta      float32 // Elapsed time for the current tick in seconds.
	Duration   float32 // Time spent in the current state in seconds.
	Events     []any   // Event markers produced during the current tick.
	BranchPath []int   // Guard branch indexes selected or rejected during evaluation.
}

// Mark appends an event marker for the caller to handle after the current tick.
func (c FSMContext) Mark(event any) FSMContext {
	c.Events = append(c.Events, event)
	return c
}

// Action transforms an FSM context. It should not mutate external state.
type Action func(FSMContext) FSMContext

// Guard checks whether an operation is allowed without changing state.
type Guard func(FSMContext) bool

// Guards evaluates a composed guard and reports the selected branch path.
type Guards interface {
	Evaluate(FSMContext) (bool, []int)
	Children() []Guards
}

// Transition is an outgoing edge from a node.
type Transition struct {
	Weight int    // Priority used to arbitrate matching outgoing transitions.
	Guards Guards // Conditions required for this transition to match.
	Action Action // Context transformation applied when this transition is committed.
	To     Node   // Destination node.
}

// Node is an immutable node definition shared by FSM instances.
type Node interface {
	ID() StateID
	Name() string
	OnEnter(FSMContext) FSMContext
	OnTick(FSMContext) (FSMContext, Status)
	OnExit(FSMContext) FSMContext
	DefaultTo() Node
	Guards() Guards
	Children() []Node
	Transitions() []Transition
	Addons() []Action
}

// TickResult reports the completed work for one frame.
type TickResult struct {
	Current      StateID // State after this tick completes.
	Previous     StateID // State exited by the most recent committed transition.
	Status       Status  // Result reported by the active node or composite node.
	Transitioned bool    // Whether one transition was committed during this tick.
	Events       []any   // Event markers produced during this tick.
	BranchPath   []int   // Guard branch path reported during this tick.
}
