package reflector

import (
	"fmt"
	"reflect"
)

// 简单代表一下数据类型
type Type int

const (
	Invalid Type = iota //不支持的类型，如chan、func、type interface、unsafe.pointer等
	Any                 //定义时为interace{}类型，运行时不确定
	Pointer             //指針類型，僅支持指向struct{}
	Bool
	Number
	String
	Struct
	Slice //slice或array
	Map   //key和value都支持的類型即可
)

type (
	Def struct {
		tp    Type         //定義類型
		refTp reflect.Type //反射refType

		//struct, struct-field
		name     string               // struct-name，field-name，TODO 注意別名情況
		fullName string               // pkg/struct-name
		sf       *reflect.StructField //若爲struct-field時，持有field引用

		//map use
		mapKeyT Type //map的key類型
		mapValT Type //map的value類型

		//slice, pointer 的下級類型
		step Type
	}

	Val struct {
		tp     Type           //實際内容的類型
		refVal *reflect.Value //反射refValue

		step *Val //pointer 的下級内容

		list    []*RefObject //slice的内容, struct的fields
		mapKeys []*RefObject //map的keys
		mapVals []*RefObject //map的values
	}

	// 反射对象包裝
	RefObject struct {
		def   *Def
		val   *Val
		index int //當前遍歷的index
	}

	Iter struct {
		root    *RefObject //根節點
		curr    *RefObject //當前引用
		prep    *RefObject //上一級引用
		handler func(w *RefObject, forward func()) bool
	}
)

func Iterator(v interface{}, handler func(w *RefObject) bool) {
	rt, rv := reflect.TypeOf(v), reflect.ValueOf(v)

	w := &RefObject{
		def: buildTypeObject(rt),
		val: buildValueObject(&rv),
	}

	//TODO 截至2025.03.08，需要實現以下：
	//將Iter對象傳進handler，裏面有當前RefObject，
	//可以通過一些輔助func()如IsEmpty()、Len()等獲取當前RefObject的狀態
	//handler返回false時，跳出當前iter，進入鄰節點繼續
	//通過自行調用farward()，進入下一級的RefObject，這裏會進棧，結束後再繼續後面的邏輯，參考middleware
}

func buildTypeObject(rt reflect.Type) *Def {

	def := &Def{
		tp:    refType(rt),
		refTp: rt,
	}

	//淺層創建，不遞歸
	switch def.tp {
	case Pointer, Slice:
		def.step = refType(rt.Elem())
	case Struct:
		//TODO 注意單純 struct{}問題
		def.name = rt.Name()
		def.fullName = fmt.Sprintf("%s/%s", rt.PkgPath(), rt.Name())
	case Map:
		def.mapKeyT = refType(rt.Key())
		def.mapValT = refType(rt.Elem())

	default:
		//TODO 注意別名情況
		fmt.Println(rt.Name(), rt.Kind())
	}

	return def

}

func buildValueObject(rv *reflect.Value) *Val {

	val := &Val{
		tp:     refType(rv.Type()),
		refVal: rv,
	}

	return val
}

func (r *RefObject) DefAny() bool {
	return r.def.tp == Any
}

func (r *RefObject) ValidType() bool {
	switch r.def.tp {
	case Pointer, Slice:
		return r.def.step != Invalid
	case Map:
		return r.def.mapKeyT != Invalid && r.def.mapValT != Invalid
	default:
		return r.def.tp != Invalid
	}
}

func (r *RefObject) ValidValueType() bool {
	return r.val.tp != Invalid
}

// 對象是否有效
func (r *RefObject) Valid() bool {
	if r.DefAny() {
		return r.ValidValueType()
	}
	return r.ValidType()
}
