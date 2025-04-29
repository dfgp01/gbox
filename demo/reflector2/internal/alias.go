package internal

type (
	StructAlias struct{}

	AsMyStruct  MyStructA
	AsMyStructP *MyStructA
)

var (
	interAlias    InterfaceAlias
	anyAlias      AnyAlias
	intAlias      IntAlias
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
