package reflector

import (
	"fmt"
	"reflect"
	"sync"
)

var (

	// 目前只預存struct，避免無限遞歸
	typeMapper   = make(map[string]*TypeObject)
	typeMapperMu sync.RWMutex

	_int      int
	_int8     int8
	_int16    int16
	_int32    int32
	_int64    int64
	_uint     uint
	_uint8    uint8
	_uint16   uint16
	_uint32   uint32
	_uint64   uint64
	_float32  float32
	_float64  float64
	_ints     []int
	_int8s    []int8
	_int16s   []int16
	_int32s   []int32
	_int64s   []int64
	_uints    []uint
	_uint8s   []uint8
	_uint16s  []uint16
	_uint32s  []uint32
	_uint64s  []uint64
	_float32s []float32
	_float64s []float64
	_bool     bool
	_string   string
	_bytes    []byte
)

func init() {
	//缓存基础类型反射
	hot(_int)
	hot(_int8)
	hot(_int16)
	hot(_int32)
	hot(_int64)

	hot(_uint)
	hot(_uint8)
	hot(_uint16)
	hot(_uint32)
	hot(_uint64)

	hot(_float32)
	hot(_float64)

	hot(_ints)
	hot(_int8s)
	hot(_int16s)
	hot(_int32s)
	hot(_int64s)

	hot(_uints)
	hot(_uint8s)
	hot(_uint16s)
	hot(_uint32s)
	hot(_uint64s)

	hot(_float32s)
	hot(_float64s)

	hot(_bool)
	hot(_string)
	hot(_bytes)
}

func hot(v interface{}) {
	//buildTypeObject(reflect.TypeOf(v))
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
func refType(t reflect.Type) ObjType {
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
		if t.Elem().Kind() == reflect.Struct {
			return Pointer
		}
	}
	//other invalid
	return Invalid
}

func nameSymbol(tp ObjType, rt reflect.Type) string {

	switch tp {
	case Invalid:
		return "!"
	case Any:
		return "{%s}"
	case Pointer:
		return "*%s"
	case Slice:
		return "[]%s"
	case Map:
		return "<%s,%s>"
	case Struct:
		return fmt.Sprintf("%s/%s", rt.PkgPath(), rt.Name())
	default:
		return rt.Name()
	}
}

// 遞歸構建反射定義對象
func buildTypeObject(rt reflect.Type) *TypeObject {

	//先判斷是否支持的類型
	obj := &TypeObject{rt: rt}
	obj.tp = refType(rt)
	//名稱佔位符
	obj.symbol = nameSymbol(obj.tp, rt)

	if obj.tp == Invalid {
		//todo 目前無法確定是否定義的interface
		// if _, ok := v.(interface{}); ok {
		// 	obj.tp = Any
		// 	obj.defAny = true
		// } else {
		// 	//不支持的類型
		// 	obj.symbol = ""
		// }
		return obj
	}

	if o, ok := typeMapper[obj.symbol]; ok {
		return o
	}

	//先占位，防struct{}無限遞歸
	if obj.tp == Struct {
		typeMapperMu.RLock()
		if o, ok := typeMapper[obj.symbol]; ok {
			return o
		}
		typeMapperMu.RUnlock()
		typeMapperMu.Lock()
		typeMapper[obj.symbol] = obj
		typeMapperMu.Unlock()
	}

	//根據具體類型遞歸
	switch obj.tp {
	case Pointer, Slice:

		sub := buildTypeObject(obj.rt.Elem())
		obj.sub = sub
		obj.symbol = fmt.Sprintf(obj.symbol, sub.symbol)
		if sub.tp == Invalid {
			//錯誤往上傳遞
			obj.tp = Invalid
		}

	case Struct:

		obj.name = rt.Name()
		for i := 0; i < rt.NumField(); i++ {
			sf := rt.Field(i)
			fieldObj := buildTypeObject(sf.Type)
			if fieldObj.tp != Invalid {
				fieldObj.sf = &sf
				fieldObj.name = sf.Name
				obj.fields = append(obj.fields, fieldObj)
			}
		}

	case Map:

		keyT := buildTypeObject(obj.rt.Key())
		valT := buildTypeObject(obj.rt.Elem())
		obj.key, obj.val = keyT, valT
		obj.symbol = fmt.Sprintf(obj.symbol, keyT.symbol, valT.symbol)
		if keyT.tp == Invalid || valT.tp == Invalid {
			//錯誤往上傳遞
			obj.tp = Invalid
		}

	}

	return obj
}
