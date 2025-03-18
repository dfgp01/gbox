package reflector2

import (
	"fmt"
)

type (
	Caller struct {
		//root    *RefObject //根節點
		curr *RefObject //當前引用
		//prep    *RefObject //上一級引用
		handler func(w *RefObject) bool
	}
)

func Iterator(v interface{}, handler func(w *RefObject) bool) {

	// 若 v 是any定義，那麽rt為空
	root := NewRefObject(v)

	// new iter
	caller := &Caller{
		//root:    root,
		//curr:    root,
		handler: handler,
	}

	caller.step(root)
}

func (c *Caller) step(next *RefObject) bool {
	c.curr = next
	return c.handler(next)
}

// forward 進入下一層級，根據類型進行不同處理
func (c *Caller) Forward() {

	var (
		obj = c.curr
	)

	//是否有效對象
	if !obj.ValidVal() {
		return
	}

	// 可以進入下一層的類型，基礎類型直接退回
	tp := obj.canStep()
	if tp == Invalid {
		return
	}

	//iter.prep = iter.curr
	//iter.prep.index = 0

	switch tp {
	case Pointer:
		next := buildPtr(obj)
		c.step(next)
	case Struct:
		//TODO 注意單純 struct{}問題
		buildStruct(obj)
		for _, next := range obj.val.list {
			if !c.step(next) {
				break
			}
		}
	case Slice:
		buildSlice(obj)
		for _, next := range obj.val.list {
			if !c.step(next) {
				break
			}
		}
	case Map:
		buildMap(obj)
		for k, v := range obj.val.kv {
			if !c.step(v) {
				fmt.Println(k.val.refVal.String())
				break
			}
		}
	default:
		//基礎類型，不需要進入，結束
		return
	}
}
