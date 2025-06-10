package internal

/**
* 盡可能提供更多的複雜類型場景
**/

type (
	MapStringAnyAlias map[string]any
	MapAnyAnyAlias    map[any]any
)

var (
	MapStringAnyVar      map[string]any
	MapStringAnyAliasVar MapStringAnyAlias
	MapAnyAnyVar         map[any]any
	MapAnyAnyAliasVar    MapAnyAnyAlias
)

func RunMapStringAny() {

	VarRun(MapStringAnyVar, "MapStringAnyVar empty-value")

	MapStringAnyVar = make(map[string]any)
	MapStringAnyVar["int-key"] = 100
	MapStringAnyVar["string-key"] = "string-value"
	MapStringAnyVar["bool-key"] = true
	MapStringAnyVar["slice-key"] = []int{1, 2, 3}

	VarRun(MapStringAnyVar, "MapStringAnyVar has-value")
}

func RunMapStringAnyAlias() {
	VarRun(MapStringAnyAliasVar, "MapStringAnyAliasVar empty-value")

	MapStringAnyAliasVar = make(map[string]any)
	MapStringAnyAliasVar["int-key"] = 100
	MapStringAnyAliasVar["string-key"] = "string-value"
	MapStringAnyAliasVar["bool-key"] = true
	MapStringAnyAliasVar["slice-key"] = []int{1, 2, 3}

	VarRun(MapStringAnyAliasVar, "MapStringAnyAliasVar has-value")
}

func RunAllComplex() {
	RunMapStringAny()
	RunMapStringAnyAlias()
}
