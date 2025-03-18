package internal

type (
	InterfaceAlias interface{}
	AnyAlias       any
	IntAlias       int
	IntsAlias      []int
	StructAlias    struct{}

	AsMyStruct  MyStructA
	AsMyStructP *MyStructA

	// UnsupportType

	MyInterface interface {
		Run()
	}
	MyFunc func(int, string) (bool, error)
	MyChan chan int
)

var (
	interA  InterfaceAlias
	anyA    AnyAlias
	intA    IntAlias
	intsA   IntsAlias
	structA StructAlias
	myA     MyStructA

	// UnsupportType

	interfaceRun MyInterface
	funcA        MyFunc
	chanA        MyChan
)
