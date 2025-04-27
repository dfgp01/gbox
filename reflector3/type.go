package reflector3

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
	switch t.Kind() {
	case reflect.Interface: //todo 注意這個
		return Any
	case reflect.Bool:
		return Bool
	case reflect.String:
		return String
	case reflect.Struct:
		return Struct
	case reflect.Slice, reflect.Array:
		return Slice
	case reflect.Map:
		return Map
	case reflect.Pointer:
		return Pointer
	default:
		if isNumber(t.Kind()) {
			return Number
		}
		return Invalid
	}
}

// 校驗有效類型
func isValidType(t reflect.Type) bool {
	tp := refType(t)
	switch tp {
	case Pointer, Slice:
		return isValidType(t.Elem())
	case Map:
		return isValidType(t.Key()) && isValidType(t.Elem())
	}
	return tp != Invalid
}

// 檢查定義類型（遞歸到底）
func (r *RefObject) ValidDefType() bool {
	return isValidType(r.refTp)
}

// 對象是否有效
func (r *RefObject) ValidVal() bool {

	// 檢查有效類型 和 數據有效性
	return isValidType(r.refVal.Type()) && r.refVal.CanInterface()
}

// 是否可以進入下一層
func (r *RefObject) canStep() bool {
	//todo panic
	tp := refType(r.refVal.Type())
	return isTypeIn(tp, Slice, Map, Struct, Pointer)
}
