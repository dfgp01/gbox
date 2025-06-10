package internal

type (
	MyStructA struct {
		Id   int
		Name string
		Ids  []int

		// inner struct
		// Recursive MyStructB	無法編譯，提示Recursive，改MyStructA也一樣
		// Recursive *MyStructB //可以編譯
		Recursive *MyStructA //可以編譯

		// Wrong Type
		Ch chan int
		Fn func(int, string) (bool, error)
	}
	MyStructB struct {
		Vals MyStructA
	}
)

var (

	// normal type
	_any       any
	_interface interface{}
	_int       int
	_ints      []int
	_intp      []*int
	//_str       string
	//_bool      bool
	_map map[string]int

	// clz
	_ptrClz *MyStructA
	_clz    MyStructA

	// slice clz
	_sliceClz  []MyStructA
	_sliceClzP []*MyStructA

	// map clz
	_mapClz  map[string]MyStructA
	_mapClzP map[bool][]*MyStructA
)

var (

	// Wrong Type

	wrongInterfaceRun interface {
		Run()
	}
	wrongFunc  func(int, string) (bool, error)
	wrongChan  chan int
	wrongSlice []*int
	wrongMap   map[chan int]func(int, string) (bool, error)
)
