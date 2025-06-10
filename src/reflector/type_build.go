package reflector

import (
	"reflect"
	"sync"
)

type (

	// 反射对象定義信息
	TypeObjectOld struct {
		_any   bool //是否定義時是interface{}
		tp     Type //定義類型
		subTp  Type
		rt     reflect.Type //反射引用
		symbol string

		sub *TypeObject //pointer、slice時的下級對象

		name     string               //struct-name，field-name
		fullName string               //pkg+struct-name
		sf       *reflect.StructField //若爲struct-field時，持有field引用
		fields   []*TypeObject        //為struct時，字段列表，僅收錄支持的類型

		key  *TypeObject //為map時的key類型
		val  *TypeObject //為map時的val類型
		keyT Type
		valT Type
	}
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

type typeMapper struct {
	sync.RWMutex
	m map[string]*TypeObject
}

func (tm *typeMapper) get(key string) *TypeObject {
	tm.RLock()
	defer tm.RUnlock()
	if o, ok := tm.m[key]; ok {
		return o
	}
	return nil
}

func (tm *typeMapper) exists(key string) bool {
	if tm.get(key) != nil {
		return true
	}
	tm.Lock()
	tm.m[key] = &TypeObject{}
	tm.Unlock()
	return false
}

var (
	// 目前只預存struct，避免無限遞歸
	tm = typeMapper{m: make(map[string]*TypeObject)}
)

func isTypeIn(t Type, ts ...Type) bool {
	if len(ts) <= 0 {
		return true
	}
	for _, v := range ts {
		if t == v {
			return true
		}
	}
	return false
}

func isTypeNotIn(t Type, ts ...Type) bool {
	if len(ts) <= 0 {
		return false
	}
	for _, v := range ts {
		if t == v {
			return false
		}
	}
	return true
}

func isBase(tk reflect.Kind) bool {
	return tk == reflect.Bool ||
		tk == reflect.String ||
		tk == reflect.Struct ||
		isNumber(tk)
}

func isNumber(tk reflect.Kind) bool {
	return isInt(tk) || isFloat(tk)
}

func isInt(tk reflect.Kind) bool {
	return tk == reflect.Int ||
		tk == reflect.Int8 ||
		tk == reflect.Int16 ||
		tk == reflect.Int32 ||
		tk == reflect.Int64 ||
		tk == reflect.Uint ||
		tk == reflect.Uint8 ||
		tk == reflect.Uint16 ||
		tk == reflect.Uint32 ||
		tk == reflect.Uint64
}

func isFloat(tk reflect.Kind) bool {
	return tk == reflect.Float32 ||
		tk == reflect.Float64 ||
		tk == reflect.Complex64 ||
		tk == reflect.Complex128
}

// 类型設定
func refType(t reflect.Type) Type {
	if t == nil {
		//interface{} or type interface，but not sure
		return Invalid
	}
	k := t.Kind()
	if k == reflect.Interface {
		//函數入口処的參數為any類型，則Kind()就會等於reflect.Interface
		//而參數為interface{}類型，則t=<nil>
		return Any
	}
	if k == reflect.Bool {
		return Bool
	} else if k == reflect.String {
		return String
	} else if k == reflect.Struct {
		return Struct
	} else if isNumber(k) {
		return Number
	} else if k == reflect.Slice || k == reflect.Array {
		return Slice
	} else if k == reflect.Map {
		return Map
	} else if k == reflect.Pointer {
		//指针的下级只能是struct
		// if t.Elem().Kind() == reflect.Struct {
		// 	return Pointer
		// }
	}
	//other invalid
	return Invalid
}
