package internal

var (
	Str          string
	WrongVar     []*int
	HackVar      struct{}
	SlicePStruct []*User
	SimpleMap    map[string]int
	NormalMap    map[any][]byte
	HardMap      map[bool][]*User
)

func RunFrame(varType any) {
	// t := reflector.ReflectTypeObject(varType)
	// t.Iter(func(name, symbol string, tp reflector.ObjType) {
	// 	fmt.Printf("name:%v, tp:%v, symbol:%v\n", name, tp, symbol)
	// })
}
