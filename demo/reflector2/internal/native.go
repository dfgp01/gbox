package internal

import (
	"gbox/reflector2"
)

type (
	MyStructA struct {
		Id   int
		Name string
		Ids  []int
		Vals IntsAlias

		// inner struct
		// Recursive MyStructB	無法編譯，提示Recursive，改MyStructA也一樣
		// Recursive *MyStructB //可以編譯
		Recursive *MyStructA //可以編譯

		// UnsupportType
		Un chan int
		F  func(int, string) (bool, error)
		Mf MyFunc
	}
	MyStructB struct {
		Vals MyStructA
	}
)

var (

	// UnsupportType

	unInterfaceRun interface {
		Run()
	}
	unFunc func(int, string) (bool, error)
	unChan chan int

	unSlice []*int
	unMap   map[int]MyChan

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

// 常規測試，未賦值
func RunNative() {
	reflector2.Iterator(_any, func(w *reflector2.RefObject) bool {
		w.Print()
		return true
	})
}

// 入口定義的any類型
func RunNative2(a any) {
	reflector2.Iterator(a, func(w *reflector2.RefObject) bool {
		w.Print()
		return true
	})
}
