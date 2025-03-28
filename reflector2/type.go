package reflector2

import (
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

func (t Type) String() string {
	switch t {
	case Invalid:
		return "Invalid"
	case Any:
		return "Any"
	case Pointer:
		return "Pointer"
	case Bool:
		return "Bool"
	case Number:
		return "Number"
	case String:
		return "String"
	case Struct:
		return "Struct"
	case Slice:
		return "Slice"
	case Map:
		return "Map"
	}
	return ""
}

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
		return Any
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
		return Pointer
	}
	//other invalid
	return Invalid
}
