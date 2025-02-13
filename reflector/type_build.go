package reflector

import (
	"fmt"
	"reflect"
	"sync"
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
}

func (tm *typeMapper) exists(key string) bool {
	if tm.get(key) {
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

// 遞歸構建反射定義對象
func buildTypeObject(rt reflect.Type) *TypeObject {

	//先判斷是否支持的類型
	obj := &TypeObject{rt: rt}
	obj.tp = refType(rt)

	if obj.tp == Invalid {
		return obj
	}

	//根據具體類型遞歸
	switch obj.tp {
	case Pointer, Slice:
		sub := buildTypeObject(obj.rt.Elem())
		obj.sub = sub
		if sub.tp == Invalid {
			//往上傳遞不支持的類型
			obj.tp = Invalid
		}
	case Struct:
		//防止無限遞歸
		obj.name = rt.Name()
		for i := 0; i < rt.NumField(); i++ {
			sf := rt.Field(i)
			fieldObj := buildTypeObject2(sf.Type)
			if fieldObj.tp != Invalid {
				fieldObj.sf = &sf
				fieldObj.name = sf.Name
				obj.fields = append(obj.fields, fieldObj)
			}
		}
	case Map:
		keyT := buildTypeObject2(obj.rt.Key())
		valT := buildTypeObject2(obj.rt.Elem())
		obj.key, obj.val = keyT, valT
		obj.symbol = fmt.Sprintf(obj.symbol, keyT.symbol, valT.symbol)
		if keyT.tp == Invalid || valT.tp == Invalid {
			//錯誤往上傳遞
			obj.tp = Invalid
		}

	}

	return obj
}

// 遞歸構建反射值對象
func buildValueObject(rv reflect.Value) *ValueObject {
	return &ValueObject{rv: rv}
}
