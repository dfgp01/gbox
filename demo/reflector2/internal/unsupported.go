package internal

import "gbox/reflector3"

/**
* 盡可能覆蓋不支持類型的場景
**/

type (
	AliasFn func()
	AliasCh chan int
)

var (
	Fn      func()
	FnAlias AliasFn
	Ch      chan int
	ChAlias AliasCh
	MapFn   map[int]func()
	MapCh   map[int]chan int
	SlcFn   []func()
	SlcCh   []chan int
	PtrFn   *func()
	PtrCh   *chan int
	MapFnk  map[*func()]int
	MapChk  map[chan int]int

	PtrAny  *any
	TwinPtr **int
)

func UnsupportedRun(_var any, title string) {
	reflector3.Iterator(_var, func(n *reflector3.Node) {
		n.Print(title)
	})
}

func UnsupportedRunALl() {
	UnsupportedRun(Fn, "Fn")
	UnsupportedRun(FnAlias, "FnAlias")
}
