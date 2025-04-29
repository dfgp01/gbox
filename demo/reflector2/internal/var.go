package internal

import "gbox/reflector3"

/**
* 盡可能覆蓋簡單變量的場景
**/

type (
	InterfaceAlias interface{}
	AnyAlias       any
	IntAlias       int
)

var (
	InterfaceVar      interface{}
	InterfaceAliasVar InterfaceAlias
	AnyVar            any
	AnyAliasVar       AnyAlias
	IntVar            int
	IntAliasVar       IntAlias
	IntsVar           []int
	IntpVar           *int
	IntpsVar          []*int
)

func VarRun(_var any, title string) {
	reflector3.Iterator(_var, func(n *reflector3.Node) {
		n.Print(title)
	})
}

func VarRunAlias() {
	//VarRun(InterfaceAliasVar, "InterfaceAliasVar")
	var alias AnyAlias = 1
	VarRun(alias, "alias-any-int")
}

func VarRunAll() {
	// VarRun(InterfaceVar, "InterfaceVar")
	// VarRun(InterfaceAliasVar, "InterfaceAliasVar")
	// VarRun(AnyVar, "AnyVar")
	// VarRun(AnyAliasVar, "AnyAliasVar")
	// VarRun(IntVar, "IntVar")
	// VarRun(IntAliasVar, "IntAliasVar")
	// VarRun(IntsVar, "IntsVar")
	// VarRun(IntpVar, "IntpVar")
	VarRun(IntpsVar, "IntpsVar")
}
