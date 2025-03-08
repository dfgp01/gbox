package reflector

import (
	"fmt"
	"reflect"
)

type (

	// 反射对象定義信息
	TypeObject struct {
		tp Type         //定義類型
		rt reflect.Type //反射引用

		//struct, struct-field
		name     string               //struct-name，field-name
		fullName string               //pkg+struct-name
		sf       *reflect.StructField //若爲struct-field時，持有field引用
		//fields   []*TypeObject        //為struct時，字段列表，僅收錄支持的類型

		//pointer、slice時的下級對象
		sub *TypeObject

		//map use
		keyT Type
		valT Type
	}
)

func (TO *TypeObject) Valid() bool {
	switch TO.tp {
	case Invalid:
		return false
	case Bool, Number, String:
		return true
	case Pointer, Slice:
		return true
	}
	return false
}

// 反射定義
func ReflectTypeObject(v interface{}) *TypeObject {
	if v == nil {
		return nil
	}
	tp := reflect.TypeOf(v)
	return buildTypeObject(tp)
}

func buildTypeObject(rt reflect.Type) *TypeObject {

	to := &TypeObject{
		rt: rt,
		tp: refType(rt),
	}

	//淺層創建，不遞歸
	switch to.tp {
	//case Pointer, Slice:
	case Struct: //TODO 注意單純 struct{}問題
		//只要名字
		to.name = rt.Name()
		to.fullName = fmt.Sprintf("%s/%s", rt.PkgPath(), rt.Name())
	case Map:
		//只要key和val的類型
		to.keyT = refType(rt.Key())
		to.valT = refType(rt.Elem())
	}

	return to
}

func buildStructType(structTO *TypeObject) {
	// add fields
	rt := structTO.rt
	for i := 0; i < rt.NumField(); i++ {
		sf := rt.Field(i)
		fieldObj := buildTypeObject(sf.Type)
		if fieldObj.tp != Invalid {
			fieldObj.sf = &sf
			fieldObj.name = sf.Name
			structTO.fields = append(structTO.fields, fieldObj)
		}
	}
}
