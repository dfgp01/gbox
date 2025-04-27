package internal

import "gbox/reflector3"

/**
* 盡可能覆蓋不支持類型的場景
**/

type (
	Fn      func()
	Ch      chan int
	MapFn   map[int]Fn
	MapCh   map[int]Ch
	SlcFn   []Fn
	SlcCh   []Ch
	PtrFn   *Fn
	PtrCh   *Ch
	MapFnk  map[*Fn]int
	MapChk  map[Ch]int
	AliasFn func()
	AliasCh chan int
	PtrAny  *any
	TwinPtr **int
)

func UnsupportedRun(_var any, title string) {
	reflector3.Iterator(_var, func(n *reflector3.Node) {
		n.Print(title)
	})
}
