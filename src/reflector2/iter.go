package reflector2

import (
	"fmt"
	"reflect"
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
	rt, rv := reflect.TypeOf(v), reflect.ValueOf(v)

	// 若 v 是any定義，那麽rt為空

	root := newRefObject(rt, &rv)

	// new iter
	caller := &Caller{
		//root:    root,
		//curr:    root,
		handler: handler,
	}

	caller.step(root)

	//TODO 截至2025.03.08，需要實現以下：
	//將Iter對象傳進handler，裏面有當前RefObject，
	//handler返回false時，跳出當前iter，進入鄰節點繼續
	//通過自行調用farward()，進入下一級的RefObject，這裏會進棧，結束後再繼續後面的邏輯，參考middleware
}

// forward 進入下一層級，根據類型進行不同處理
func (c *Caller) Forward() {

	var (
		obj = c.curr
		val = c.curr.val
	)

	//本身是否有效
	if !obj.Valid() {
		return
	}

	// 可以進入下一層的類型，基礎類型直接退回
	if !isTypeIn(val.tp, Pointer, Struct, Slice, Map) {
		return
		//iter.prep = iter.curr
		//iter.prep.index = 0
	}

	switch val.tp {
	case Pointer:
		buildPtr(obj)
		c.step(val.step)
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

func (c *Caller) step(next *RefObject) bool {
	c.curr = next
	return c.handler(next)
}
