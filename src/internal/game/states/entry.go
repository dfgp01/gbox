package states

import (
	"fmt"
	def "gbox/def/game/states"
)

func Check(node def.Node) error {
	// 合法检查
	if node == nil {
		return fmt.Errorf("node is nil")
	}
	if node.ID() <= def.InvalidStateID {
		return fmt.Errorf("node %q has invalid ID", node.Name())
	}
	if node.Name() == "" {
		return fmt.Errorf("node %q has empty name", node.Name())
	}

	//节点类型检查
	switch node.Type() {
	case def.Leaf, def.Terminal:
		//需要在这里检查终态
		if node.Type() == def.Terminal {
			return nil
		}
		// 周期检查
		if node.Guards() == nil || node.DefaultTo() == nil {
			return fmt.Errorf("leaf node %q must have guards and default transition", node.Name())
		}
		// 迁移检查
		if node.Transitions() == nil || len(node.Transitions()) == 0 {
			return fmt.Errorf("leaf node %q must have at least one transition", node.Name())
		}
	case def.Serial, def.Parallel, def.Selector:
		if len(node.Children()) == 0 {
			return fmt.Errorf("composite node %q has no children", node.Name())
		}
	default:
		return fmt.Errorf("node %q has unknown type %v", node.Name(), node.Type())
	}
	return nil
}

func IsTerminal(node def.Node) bool {
	return node != nil && node.DefaultTo() == nil
}

func IsLeaf(node def.Node) bool {
	return node != nil && node.Type() == def.Leaf
}

func IsHierarchical(node def.Node) bool {
	return node != nil && node.Type() != def.Leaf
}

// 保留这个命名
func IsComposite(node def.Node) bool {
	return false
}
