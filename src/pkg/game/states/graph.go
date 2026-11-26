package states

import (
	"errors"
	"fmt"
	def "gbox/def/game/states"
	"reflect"
)

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
		if err := validateNode(node); err != nil {
			return nil, err
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

func validateNode(node def.Node) error {
	switch node.Type() {
	case def.Leaf:
		if len(node.Children()) != 0 {
			return fmt.Errorf("states: leaf node %q cannot have children", node.Name())
		}
		if node.Guards() == nil || node.DefaultTo() == nil {
			return fmt.Errorf("states: leaf node %q requires guards and a default target", node.Name())
		}
		if len(node.Transitions()) == 0 {
			return fmt.Errorf("states: leaf node %q requires at least one outgoing transition", node.Name())
		}
	case def.Terminal:
		if len(node.Children()) != 0 || node.DefaultTo() != nil || len(node.Transitions()) != 0 {
			return fmt.Errorf("states: terminal node %q cannot have children or outgoing transitions", node.Name())
		}
	case def.Serial, def.Parallel, def.Selector:
		if len(node.Children()) == 0 {
			return fmt.Errorf("states: composite node %q requires at least one child", node.Name())
		}
	default:
		return fmt.Errorf("states: node %q has unknown type %d", node.Name(), node.Type())
	}
	return nil
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
