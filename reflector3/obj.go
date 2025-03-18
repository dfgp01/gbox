package reflector2

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
		refTp  reflect.Type         //反射refType
		refVal *reflect.Value       //反射refValue
		field  *reflect.StructField //field引用，僅為struct-field時
	}
)

// 臨時的
func (r *RefObject) Print() {
	fmt.Printf("RefObject Info:\n")
	fmt.Printf("  Definition:\n")
	fmt.Printf("    Type: %v\n", r.def.tp.String())
	fmt.Printf("    Name: %v\n", r.def.name)
	fmt.Printf("    Full Path: %v\n", r.def.full)

	switch r.def.tp {
	case Pointer, Slice:
		fmt.Printf("    Element Type: %v\n", r.def.step.String())
	case Map:
		fmt.Printf("    Map Key Type: %v\n", r.def.mapKeyT.String())
		fmt.Printf("    Map Value Type: %v\n", r.def.mapValT.String())
	case Struct:
		if r.def.refTp != nil {
			fmt.Printf("    Fields Count: %d\n", r.def.refTp.NumField())
		}
	}

	if r.def.sf != nil {
		fmt.Printf("    Field Name: %s, Tag: %s\n", r.def.name, r.def.sf.Tag)
	}

	// 以下是 val 部分

	fmt.Printf("  Value:\n")
	fmt.Printf("    Type: %v\n", r.val.tp.String())

	if r.val.refVal.IsValid() {
		switch r.val.tp {
		case Map:
			fmt.Printf("    Map Entries: %d\n", r.Len())
		case String, Slice:
			fmt.Printf("    Length: %d\n", r.Len())
		case Struct:
			fmt.Printf("    Available Fields Count: %d\n", r.Len())
		}

		if r.val.refVal.CanInterface() {
			fmt.Printf("    Value: %v\n", r.val.refVal.Interface())
		}
	}

	fmt.Printf("    Has Pointer Element: %v\n", r.val.step != nil)
	fmt.Printf("    List Items: %d\n", len(r.val.list))
	fmt.Printf("    Map Entries: %d\n", len(r.val.kv))
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
		if !IsValidType(field.Type) {
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

// 獲取struct的有效field，rv為struct的reflect.Value
func (r *RefObject) fields(rv *reflect.Value) []*reflect.StructField {
	if refType(rv.Type()) != Struct {
		return nil
	}
	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)

		// if field.CanInterface() {
		// 	fields = append(fields, field)
		// }
	}

	return r.refVal.NumField()
}

func NewRefObject(v interface{}) *RefObject {
	// var v any，那麽rt會是nil
	rt, rv := reflect.TypeOf(v), reflect.ValueOf(v)
	return newRefObject(rt, &rv)
}

func newRefObject(rt reflect.Type, rv *reflect.Value) *RefObject {
	ref := &RefObject{
		refTp:  rt,
		refVal: rv,
	}
	return ref
}

func newTypeObject(rt reflect.Type) *Def {

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
		def.full = fmt.Sprintf("%s/%s", rt.PkgPath(), rt.Name())
	case Map:
		def.mapKeyT = refType(rt.Key())
		def.mapValT = refType(rt.Elem())
	default:
		//TODO 注意別名情況
	}

	return def

}

func newValueObject(rv *reflect.Value) *Val {

	//fmt.Println("111111111111", rv.IsValid())
	//fmt.Println("222222222222", rv.IsZero())
	//fmt.Println("333333333333", rv.IsNil())

	val := &Val{
		refVal: rv,
	}

	if !rv.IsValid() {
		val.tp = Invalid
	} else if rv.IsNil() {
		val.nl = true
	} else {
		val.tp = refType(rv.Type())
	}

	return val
}

func buildPtr(s *RefObject) *RefObject {
	v := s.refVal.Elem()
	return newRefObject(v.Type(), &v)
}

// 構建struct對象，s為struct的RefObject
func buildStruct(s *RefObject) []*RefObject {
	//將合法的field加入list
	var fields []*RefObject
	for i := 0; i < s.refTp.NumField(); i++ {
		field := s.refTp.Field(i)
		fieldVal := s.refVal.Field(i)
		fieldObj := newRefObject(field.Type, &fieldVal)
		fieldObj.field = &field
		if fieldObj.ValidVal() {
			fields = append(fields, fieldObj)
		}
	}
	return fields
}

// 構建slice對象，s為slice的RefObject
func buildSlice(s *RefObject) []*RefObject {
	var list []*RefObject
	for i := 0; i < s.refVal.Len(); i++ {
		sliceVal := s.refVal.Index(i)
		sliceObj := newRefObject(sliceVal.Type(), &sliceVal)
		if sliceObj.Valid() {
			s.val.list = append(s.val.list, sliceObj)
		}
	}
}

// 構建map對象，m為map的RefObject
func buildMap(m *RefObject) {
	// 將map-val的refVal，加入list
	m.val.kv = make(map[*RefObject]*RefObject)
	for _, key := range m.val.refVal.MapKeys() {
		mapKey := newRefObject(m.def.refTp.Key(), &key)
		v := m.val.refVal.MapIndex(key)
		mapVal := newRefObject(m.def.refTp.Elem(), &v)
		if mapKey.Valid() && mapVal.Valid() {
			m.val.kv[mapKey] = mapVal
		}
	}
}
