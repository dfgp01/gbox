package reflector

import (
	"reflect"
)

// 简单代表一下数据类型
type ObjType int

const (
	Invalid ObjType = iota //不支持的类型，如chan、func、type interface、unsafe.pointer等
	Any                    //定义时为interace{}类型，运行时不确定
	Pointer                //指針類型，僅支持指針指向struct{}
	Bool
	Number
	String
	Struct
	Slice //slice或array
	Map   //key和value都支持的類型即可
)

// 反射对象定義信息
type TypeObject struct {
	defAny bool         //是否定義時是interface{}
	tp     ObjType      //封裝類型
	rt     reflect.Type //反射對象
	symbol string

	sub *TypeObject //pointer、slice會有

	name   string               //struct-name，field-name
	sf     *reflect.StructField //持有field引用
	fields []*TypeObject        //為struct時，字段列表，僅收錄支持的類型

	key *TypeObject //為map時的key類型
	val *TypeObject //為map時的val類型
}

func (t *TypeObject) Iter(h func(name, symbol string, tp ObjType)) {
	iter(t, h)
}

func iter(t *TypeObject, h func(name, symbol string, tp ObjType)) {
	switch t.tp {
	case Pointer, Slice:
		iter(t.sub, h)
	case Map:
		iter(t.key, h)
		iter(t.val, h)
	case Struct:
		for _, field := range t.fields {
			iter(field, h)
		}
	default:
	}
	h(t.name, t.symbol, t.tp)
}

// 反射對象定義信息
func ReflectTypeObject(v interface{}) *TypeObject {
	if v == nil {
		return nil
	}
	tp := reflect.TypeOf(v)
	return buildTypeObject(tp)
}
