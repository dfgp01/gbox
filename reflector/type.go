package reflector

import (
	"fmt"
	"reflect"
)

type (

	// 反射对象定義信息
	TypeObject struct {
		defAny bool         //是否定義時是interface{}
		tp     Type         //定義類型
		rt     reflect.Type //反射引用
		symbol string

		sub *TypeObject //pointer、slice時的下級對象

		name   string               //struct-name，field-name
		sf     *reflect.StructField //若爲struct-field時，持有field引用
		fields []*TypeObject        //為struct時，字段列表，僅收錄支持的類型

		key *TypeObject //為map時的key類型
		val *TypeObject //為map時的val類型
	}

	// 反射对象值信息
	ValueObject struct {
		tp   Type           //實際類型
		rv   reflect.Value  //反射對象
		sub  *ValueObject   //pointer會有
		list []*ValueObject //slice, struct會有
		mKey *ValueObject   //map-entry-key
		mVal *ValueObject   //map-entry-val
	}

	// 反射对象包裝
	RefObject struct {
		t     *TypeObject
		v     *ValueObject
		l     int //struct-fields、slice、map-entry的長度
		index int
	}
	RefWrapper struct {
		root    *RefObject //根節點
		curr    *RefObject //當前引用
		prep    *RefObject //上一級引用
		handler func(w *RefWrapper, forward func()) bool
	}
)

// 反射對象定義信息
func ReflectTypeObject(v interface{}) *TypeObject {
	if v == nil {
		return nil
	}
	tp := reflect.TypeOf(v)
	return buildTypeObject(tp)
}

// 反射對象定義信息
func ReflectValueObject(v interface{}) *ValueObject {
	if v == nil {
		return nil
	}
	val := reflect.ValueOf(v)
	return buildValueObject(val)
}

func Iter(v interface{}, handler func(w *RefWrapper, forward, step func()) bool) {
	if v == nil {
		return
	}
	t, rv := reflect.TypeOf(v), reflect.ValueOf(v)
	w := &RefWrapper{
		root: &RefObject{
			t: buildTypeObject(t),
			v: buildValueObject(rv),
		},
		handler: handler,
	}

	w.curr = w.root

	//root 層開始
	w.forward()
}

func build(rt reflect.Type, rv reflect.Value) *RefObject {
	o := &RefObject{
		t: buildTypeObject(t),
		v: buildValueObject(rv),
	},
}

// 單步跳進（確認進入迭代）
func (w *RefWrapper) forward() {
	curr := w.curr

	//interface 判斷
	if curr.t.tp == Any {
		v := curr.v.rv.Interface()
		if _, ok := v.(interface{}); ok {
			curr.t.defAny = true
		} else {
			//不支持的類型
			curr.t.tp = Invalid
		}
	}

	tp := curr.t.tp
	if typeNotIn(tp, Pointer, Slice, Map, Struct) {
		w.handler(w, w.forward, w.step)
		return
	}
	w.prep = curr

	switch tp {
	case Pointer:
		curr.t.sub
		curr.v.rv.Elem()
	case Slice:
		curr.t.sub
		curr.v.rv.Index(i)
	case Map:
	case Struct:

	}

	w.handler(w, w.forward, w.step)
}

// 單步進入
func (w *RefWrapper) step() {
	w.handler(w, w.forward)
}

func (w *RefWrapper) Out() {
	tp := w.curr.t.tp
	vp := w.curr.v.tp
	fmt.Printf("Type out1 tp:%v, vp:%v  \n", tp, vp)
}
