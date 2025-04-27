package internal

type (
	InterfaceAlias interface{}
	AnyAlias       any
	IntsAlias      []int
	StructAlias    struct{}

	AsMyStruct  MyStructA
	AsMyStructP *MyStructA
)

var (
	interAlias    InterfaceAlias
	anyAlias      AnyAlias
	intAlias      IntAlias
	intsAlias     IntsAlias
	structAlias   StructAlias
	myStructAlias MyStructA
)

type (
	// Wrong Alias Type

	WrongInterfaceAlias interface {
		Run()
	}
	WrongFuncAlias       func(int, string) (bool, error)
	WrongChanAlias       chan int
	WrongMapAlias        map[int]chan int
	WrongSliceChanAlias  []chan int
	WrongSliceChanAlias2 []WrongChanAlias
)

var (
	// Wrong Alias Var

	wrongInterfaceAlias  WrongInterfaceAlias
	wrongFuncAlias       WrongFuncAlias
	wrongChanAlias       WrongChanAlias
	wrongMapAlias        WrongMapAlias
	wrongSliceChanAlias  WrongSliceChanAlias
	wrongSliceChanAlias2 WrongSliceChanAlias2
)

func AliasUnsupportType() {
	// 以下的Value都支持：
	// IsValid: true
	// CanInterface: true
	// IsNil: true
	// IsZero: true

	// Type: internal.WrongFuncAlias
	// Kind: func
	// Name: WrongFuncAlias
	// FullPath: gbox/demo/reflector2/internal
	nativeAnyDef(wrongInterfaceAlias, "wrongInterfaceAlias")

	// Name and FullPath is nil
	nativeAnyDef(wrongFuncAlias, "wrongFuncAlias")

	nativeAnyDef(wrongChanAlias, "wrongChanAlias")

	//Name: MyChan FullPath: gbox/demo/reflector2/internal
	nativeAnyDef(wrongMapAlias, "wrongMapAlias")

	nativeAnyDef(wrongSliceChanAlias, "wrongSliceChanAlias")

	nativeAnyDef(wrongSliceChanAlias2, "wrongSliceChanAlias2")

}

// 帶別名，未賦值
func AliasRun() {
	nativeAnyDef(interAlias, "interAlias")
	nativeAnyDef(anyAlias, "anyAlias")
	nativeAnyDef(intAlias, "intAlias")
	nativeAnyDef(intsAlias, "intsAlias")
	nativeAnyDef(structAlias, "structAlias")
	nativeAnyDef(myStructAlias, "myStructAlias")
}
