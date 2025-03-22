package reflector3

import (
	"fmt"
	"reflect"
)

type (
	// 先定義反射對象的基本信息
	Object interface {
		DefType() Type   //定義時的類型
		DefAny() bool    //是否定義為any類型
		ValidType() bool //定義類型是否有效

		ValType() Type //實際值的類型

		Len() int    //數據長度，slice、map、string的len()
		Empty() bool //是否為空，number=0, string="", struct{}, map=nil, slice=nil, pointer=nil
		Value() any  //實際值
	}

	// 反射对象包裝
	RefObject struct {
		defAny bool                 //是否定義為any類型（好像意義不大）
		refTp  reflect.Type         //反射值的實際refVal.Type()
		refVal *reflect.Value       //反射值refValue
		field  *reflect.StructField //field引用，僅為struct-field時

		next   *RefObject                //為pointer時，指向的下一個對象
		list   []*RefObject              //為slice時，對象列表
		kv     map[*RefObject]*RefObject //為map時，key-value對象
		fields []*RefObject              //為struct時，field對象列表
	}
)

func PrintDefinition(t reflect.Type) {
	fmt.Printf("  Definition:\n")
	fmt.Printf("    Type: %v\n", t.String())
	fmt.Printf("    Kind: %v\n", t.Kind())
	fmt.Printf("    Name: %v\n", t.Name())
	fmt.Printf("    FullPath: %v\n", t.PkgPath())
}

func PrintValue(v *reflect.Value) {
	fmt.Printf("  Value:\n")
	fmt.Printf("    IsValid: %v\n", v.IsValid())
	if v.IsValid() {
		// panic if !v.IsValid()
		fmt.Printf("    CanInterface: %v\n", v.CanInterface())
		fmt.Printf("    IsNil: %v\n", v.IsNil())
		fmt.Printf("    IsZero: %v\n", v.IsZero())
		fmt.Printf("    Type: %v\n", v.Type().String())
		fmt.Printf("    Value: %v\n", v.Interface())
	}
}

// 臨時的
func (r *RefObject) Print(title string) {
	fmt.Printf("************** %s :\n", title)
	defer func() {
		// 以下是 val 部分
		PrintValue(r.refVal)
		fmt.Printf("\n")
	}()
	if r.refTp == nil {
		fmt.Printf("   defAny: %v\n", r.defAny)
		return
	}

	PrintDefinition(r.refTp)

	switch refType(r.refTp) {
	case Pointer:
		fmt.Printf(" *Element\n")
		PrintDefinition(r.refTp.Elem())
	case Slice:
		fmt.Printf(" []Element\n")
		PrintDefinition(r.refTp.Elem())
	case Map:
		fmt.Printf(" Map Key\n")
		PrintDefinition(r.refTp.Key())
		fmt.Printf(" Map Value\n")
		PrintDefinition(r.refTp.Elem())
	case Struct:
		for i := 0; i < r.refTp.NumField(); i++ {
			field := r.refTp.Field(i)
			fmt.Printf(" Field Name: %s, Tag: %s\n", field.Name, field.Tag)
			PrintDefinition(field.Type)
		}
	}
}

// 長度，slice、map、string的len()，struct的有效field數量
func (r *RefObject) Len() int {
	tp := refType(r.refVal.Type())
	switch tp {
	case String, Slice, Map:
		return r.refVal.Len()
	case Struct:
		return len(r.fieldDefs(r.refVal.Type()))
	default:
		return 0
	}
}

// 獲取struct的有效field定義，rv為struct的reflect.Value
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

func NewRefObject(v interface{}) *RefObject {
	// var v any，那麽rt會是nil
	rt, rv := reflect.TypeOf(v), reflect.ValueOf(v)
	return buildRefObject(rt, &rv)
}

func buildRefObject(rt reflect.Type, rv *reflect.Value) *RefObject {
	ref := &RefObject{
		refTp:  rt,
		refVal: rv,
	}
	if rt == nil {
		ref.defAny = true
		return ref
	}
	//ref.refTp = rv.Type()
	//ref.refVal = rv

	switch refType(ref.refTp) {
	case Pointer:
		ref.next = buildPtr(ref)
	case Slice:
		ref.list = buildSlice(ref)
	case Struct:
		ref.fields = buildStruct(ref)
	case Map:
		ref.kv = buildMap(ref)
	}

	return ref
}

func buildPtr(prep *RefObject) *RefObject {
	v := prep.refVal.Elem()
	return buildRefObject(prep.refTp.Elem(), &v)
}

// 構建slice對象
func buildSlice(prep *RefObject) []*RefObject {
	var list []*RefObject
	for i := 0; i < prep.refVal.Len(); i++ {
		sliceVal := prep.refVal.Index(i)
		sliceObj := buildRefObject(prep.refTp.Elem(), &sliceVal)
		if sliceObj.ValidVal() {
			list = append(list, sliceObj)
		}
	}
	return list
}

// 構建struct對象
func buildStruct(prep *RefObject) []*RefObject {
	//將合法的field加入list
	var fields []*RefObject
	for i := 0; i < prep.refTp.NumField(); i++ {
		field := prep.refTp.Field(i)
		fieldVal := prep.refVal.Field(i)
		fieldObj := buildRefObject(field.Type, &fieldVal)
		fieldObj.field = &field
		fields = append(fields, fieldObj)
	}
	return fields
}

func buildMap(prep *RefObject) map[*RefObject]*RefObject {
	m := make(map[*RefObject]*RefObject)
	for _, key := range prep.refVal.MapKeys() {
		mapKey := buildRefObject(prep.refTp.Key(), &key)
		mapVal := prep.refVal.MapIndex(key)
		mapObj := buildRefObject(mapVal.Type(), &mapVal)
		m[mapKey] = mapObj
	}
	return m
}
