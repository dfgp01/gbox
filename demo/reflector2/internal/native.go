package internal

import "gbox/reflector3"

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

// 常規測試，未賦值
func NativeAny() {
	reflector3.Iterator(_any, func(w *reflector3.RefObject) bool {
		w.Print("any")
		return true
	})
}

// 入口定義的any類型
func nativeAnyDef(a any, title string) {
	reflector3.Iterator(a, func(w *reflector3.RefObject) bool {
		w.Print(title)
		return true
	})
}

// 常規類型，未賦值
func NativeNormalRun() {
	nativeAnyDef(_any, "any")
	nativeAnyDef(_interface, "interface")

	// Name=int FullPath= IsValid=true panic->IsNil
	nativeAnyDef(_int, "int")

	// Name=ints FullPath= IsValid=true
	nativeAnyDef(_ints, "ints")

	// Name=intp FullPath= IsValid=true
	nativeAnyDef(_intp, "intp")

	nativeAnyDef(_map, "map")
	nativeAnyDef(_ptrClz, "ptrClz")
	nativeAnyDef(_clz, "clz")
	nativeAnyDef(_sliceClz, "sliceClz")
	nativeAnyDef(_sliceClzP, "sliceClzP")
	nativeAnyDef(_mapClz, "mapClz")
	nativeAnyDef(_mapClzP, "mapClzP")
}

// 不支持類型，未賦值
func NativeUnsupportType() {

	// 以下的Value都支持：
	// IsValid: true
	// CanInterface: true
	// IsNil: true
	// IsZero: true

	// only wrongInterfaceRun -> reflect.Type is nil
	nativeAnyDef(wrongInterfaceRun, "wrongInterfaceRun")
	nativeAnyDef(wrongFunc, "wrongFunc")

	nativeAnyDef(wrongChan, "wrongChan")
	nativeAnyDef(wrongSlice, "wrongSlice")
	nativeAnyDef(wrongMap, "wrongMap")
}
