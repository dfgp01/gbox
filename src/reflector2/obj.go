package reflector2

import (
	"fmt"
	"reflect"
)

type (
	Def struct {
		tp    Type         //定義類型
		refTp reflect.Type //反射refType

		//struct, struct-field
		name string               // struct-name，field-name，TODO 注意別名情況
		full string               // pkg/struct-name
		sf   *reflect.StructField //若爲struct-field時，持有field引用

		//map use
		mapKeyT Type //map的key類型
		mapValT Type //map的value類型

		//slice, pointer 的下級類型
		step Type
	}

	Val struct {
		nl     bool           //是否為nil
		tp     Type           //實際内容的類型
		refVal *reflect.Value //反射refValue

		step *RefObject //pointer 的下級内容

		list []*RefObject              //slice的内容, struct的fields
		kv   map[*RefObject]*RefObject //map的内容

	}

	// 反射对象包裝
	RefObject struct {
		def *Def
		val *Val
	}
)

// 臨時的
func (r *RefObject) Print(title string) {
	fmt.Printf("%s Info:\n", title)
	fmt.Printf("  Definition:\n")
	fmt.Printf("    Type: %v\n", r.def.tp)
	fmt.Printf("    Name: %v\n", r.def.name)
	fmt.Printf("    Full Path: %v\n", r.def.full)

	switch r.def.tp {
	case Pointer, Slice:
		fmt.Printf("    Element Type: %v\n", r.def.step)
	case Map:
		fmt.Printf("    Map Key Type: %v\n", r.def.mapKeyT)
		fmt.Printf("    Map Value Type: %v\n", r.def.mapValT)
	case Struct:
		if r.def.refTp != nil {
			fmt.Printf("    Fields Count: %d\n", r.def.refTp.NumField())
		}
	}

	if r.def.sf != nil {
		fmt.Printf("    Field Name: %s\n", r.def.sf.Name)
		if len(r.def.sf.Tag) > 0 {
			fmt.Printf("    Field Tag: %s\n", r.def.sf.Tag)
		}
	}

	fmt.Printf("  Value:\n")
	fmt.Printf("    Type: %v\n", r.val.tp)
	if r.val.nl {
		fmt.Printf("    Is Nil: true\n")
	}

	if r.val.refVal != nil && r.val.refVal.IsValid() {
		switch r.val.tp {
		case String, Slice, Map:
			fmt.Printf("    Length: %d\n", r.Len())
		case Struct:
			if r.val.refVal.Type().NumField() > 0 {
				fmt.Printf("    Fields Count: %d\n", r.val.refVal.Type().NumField())
			}
		}

		if r.val.refVal.CanInterface() {
			fmt.Printf("    Value: %v\n", r.val.refVal.Interface())
		}
	}

	fmt.Printf("    Has Pointer Element: %v\n", r.val.step != nil)
	fmt.Printf("    List Items: %d\n", len(r.val.list))
	fmt.Printf("    Map Entries: %d\n", len(r.val.kv))
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

// 數據長度，只有slice, map, string有效
func (r *RefObject) Len() int {
	switch r.val.tp {
	case String:
		return r.val.refVal.Len()
	case Map:
		return len(r.val.kv)
	case Slice, Struct:
		return len(r.val.list)
	default:
		return 0
	}
}

func newRefObject(rt reflect.Type, rv *reflect.Value) *RefObject {
	ref := &RefObject{
		def: newTypeObject(rt),
		val: newValueObject(rv),
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

func buildPtr(s *RefObject) {
	v := s.val.refVal.Elem()
	s.val.step = newRefObject(s.def.refTp.Elem(), &v)
}

// 構建struct對象，s為struct的RefObject
func buildStruct(s *RefObject) {
	//將合法的field加入list
	for i := 0; i < s.def.refTp.NumField(); i++ {
		field := s.def.refTp.Field(i)
		fieldVal := s.val.refVal.Field(i)
		fieldObj := newRefObject(field.Type, &fieldVal)
		if fieldObj.Valid() {
			fieldObj.def.name = field.Name
			fieldObj.def.sf = &field
			s.val.list = append(s.val.list, fieldObj)
		}
	}
}

// 構建slice對象，s為slice的RefObject
func buildSlice(s *RefObject) {
	// 將slice-val的refVal，加入list
	for i := 0; i < s.val.refVal.Len(); i++ {
		sliceVal := s.val.refVal.Index(i)
		sliceObj := newRefObject(s.def.refTp.Elem(), &sliceVal)
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
