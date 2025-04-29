package reflector3

import (
	"fmt"
	"reflect"
)

type (
	// 先定義反射對象的基本信息
	Object interface {
		DefType() Type          //定義時的類型
		ValType() Type          //實際值的類型	TODO 可能只需要一個
		RefTp() reflect.Type    //反射值的實際refVal.Type()
		RefVal() *reflect.Value //反射值refValue
		DefAny() bool           //是否定義為any類型
		Value() any             //實際值
		Valid() bool            //是否有效類型
		Len() int               //數據長度，slice、map、string的len()
		Empty() bool            //是否為空，number=0, string="", struct{}, map=nil, slice=nil, pointer=nil
	}

	// 基礎類型反射對象
	RefObject struct {
		defAny   bool           //是否定義為any類型（TODO 好像意義不大）
		emptyVar bool           //是否為空變量（TODO 同上）
		refTp    reflect.Type   //反射值的實際refVal.Type()
		refVal   *reflect.Value //反射值refValue
	}

	// 未知類型，refTp=nil
	AnyObject struct {
		RefObject
	}

	// 不支持的類型
	InvalidObject struct {
		RefObject
	}

	// 基礎類型反射對象
	BaseRefObject struct {
		RefObject
		tp Type //基礎類型
	}

	// 指針類型反射對象
	PtrRefObject struct {
		RefObject
		elem Object //指向的下一個對象
	}

	// struct-field封裝對象
	StructFieldObject struct {
		Object                        //實際内容
		fieldDef *reflect.StructField //field定義
	}

	// struct類型反射對象
	StructRefObject struct {
		RefObject
		fields []*StructFieldObject //field對象列表
	}

	// slice類型反射對象
	SliceRefObject struct {
		RefObject
		elems []Object //slice内的對象
	}

	// map的key-value對象
	MapEntryObject struct {
		RefObject
		key   Object //map的key
		value Object //map的value
	}

	// map類型反射對象
	MapRefObject struct {
		RefObject
		keys  []Object //map的key
		elems []Object //map的value
	}
)

/**
*	common implements
 */

func (r *RefObject) DefType() Type          { return refType(r.refTp) }
func (r *RefObject) ValType() Type          { return refType(r.refVal.Type()) }
func (r *RefObject) RefTp() reflect.Type    { return r.refTp }
func (r *RefObject) RefVal() *reflect.Value { return r.refVal }
func (r *RefObject) DefAny() bool           { return r.defAny }
func (r *RefObject) Value() any             { return r.refVal.Interface() }
func (r *RefObject) Valid() bool            { return true } // 缺省處理
func (r *RefObject) Len() int               { return 0 }    // 缺省處理

// 缺省處理
func (r *RefObject) Empty() bool {
	//return r.refVal.CanInterface() && r.refVal.Len() == 0
	return true
}

/**
*	invalid implements
 */

func (r *InvalidObject) DefType() Type {
	if r.defAny {
		return Any
	}
	return Invalid
}

func (r *InvalidObject) ValType() Type {
	return Invalid
}

func (r *InvalidObject) Valid() bool {
	return false
}

/**
*	base implements
 */

func (r *BaseRefObject) Empty() bool {
	return r.refVal.IsZero()
}

func (r *BaseRefObject) Len() int {
	if r.tp == String {
		return r.refVal.Len()
	}
	return 0
}

/**
*	ptr implements
 */

func (r *PtrRefObject) Empty() bool {
	return r.refVal.IsNil()
}

/**
*	struct implements
		todo 空struct{}問題， Empty()和Valid()的處理
*/

func (r *StructRefObject) Empty() bool {
	return r.refVal.IsZero()
}

func (r *StructRefObject) Len() int {
	// 有效field數量
	count := 0
	for _, field := range r.fields {
		if field.Valid() {
			count++
		}
	}
	return count
}

/**
*	slice implements
 */

func (r *SliceRefObject) Empty() bool {
	return r.refVal.IsNil()
}

func (r *SliceRefObject) Len() int {
	return r.refVal.Len()
}

/**
*	map implements
 */

func (r *MapRefObject) Empty() bool {
	return r.refVal.IsNil()
}

func (r *MapRefObject) Len() int {
	return r.refVal.Len()
}

func NewRefObject(v interface{}) Object {
	// var v any，那麽rt會是nil
	rt, rv := reflect.TypeOf(v), reflect.ValueOf(v)
	return buildReflector(rt, &rv)
}

func buildReflector(rt reflect.Type, rv *reflect.Value) Object {
	in := RefObject{
		refTp:  rt,
		refVal: rv,
	}

	if rt == nil {
		o := &AnyObject{
			RefObject: in,
		}
		o.defAny = true
		o.emptyVar = true
		return o
	}

	switch tp := refType(rt); tp {
	case Invalid:
		return &InvalidObject{in}
	case Any:
		// 嘗試能不能走到這裏？？？
		fmt.Println("sssssssssssssssssssssssssssssssssssssss", rt.String())
		return nil
	case Bool, Number, String:
		return &BaseRefObject{
			RefObject: in,
			tp:        tp,
		}
	case Pointer:
		nextVal := rv.Elem()
		return &PtrRefObject{
			RefObject: in,
			elem:      buildReflector(rt.Elem(), &nextVal),
		}
	case Slice:
		elems := make([]Object, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			sliceVal := rv.Index(i)
			elems[i] = buildReflector(rt.Elem(), &sliceVal)
		}
		return &SliceRefObject{
			RefObject: in,
			elems:     elems,
		}
	case Struct:
		fields := make([]*StructFieldObject, 0)
		for i := 0; i < rt.NumField(); i++ {
			fieldDef := rt.Field(i)
			fieldVal := rv.Field(i)
			fields = append(fields, &StructFieldObject{
				Object:   buildReflector(fieldDef.Type, &fieldVal),
				fieldDef: &fieldDef,
			})
		}
		return &StructRefObject{
			RefObject: in,
			fields:    fields,
		}
	case Map:
		keys := make([]Object, 0)
		elems := make([]Object, 0)
		for _, key := range rv.MapKeys() {
			mapKey := buildReflector(rt.Key(), &key)
			mapVal := rv.MapIndex(key)
			keys = append(keys, mapKey)
			elems = append(elems, buildReflector(rt.Elem(), &mapVal))
		}
		return &MapRefObject{
			RefObject: in,
			keys:      keys,
			elems:     elems,
		}
	}

	return nil
}

// 獲取struct的有效field定義，rt為struct的reflect.Type
func (r *RefObject) fieldDefs(rt reflect.Type) []*reflect.StructField {
	if refType(rt) != Struct {
		return nil
	}

	fields := make([]*reflect.StructField, 0)
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		if !isValidType(field.Type) {
			continue
		}
		if field.Anonymous {
			fields = append(fields, r.fieldDefs(field.Type)...)
		} else {
			fields = append(fields, &field)
		}
	}
	return fields
}
