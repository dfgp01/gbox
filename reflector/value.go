package reflector

import (
	"fmt"
	"reflect"
)

type (

	// 反射对象值信息
	ValueObject struct {
		_type *TypeObject    //實際類型
		rv    *reflect.Value //反射對象
		sub   *ValueObject   //pointer會有
		list  []*ValueObject //slice的内容, struct的fields，pointer的[0]，map的[0]=key, [1]=val
		mKey  *ValueObject   //map-entry-key
		mVal  *ValueObject   //map-entry-val
	}
)

// 反射對象定義信息
func ReflectValueObject(v interface{}) *ValueObject {
	val := reflect.ValueOf(v)
	return buildValueObject(&val)
}

// 遞歸構建反射值對象
func buildValueObject(rv *reflect.Value) *ValueObject {

	//先判斷是否支持的類型
	obj := &ValueObject{
		rv:    rv,
		_type: buildTypeObject(rv.Type()),
	}
	if obj._type.tp == Invalid {
		return obj
	}

	//根據具體類型遞歸
	switch obj._type.tp {
	case Pointer, Slice:
		rv2 := obj.rv.Elem()
		sub := buildValueObject(&rv2)
		obj.sub = sub
		if sub._type.tp == Invalid {
			//往上傳遞不支持的類型
			obj._type.tp = Invalid
		}
	case Struct:
		//防止無限遞歸
		obj.name = rv.Name()
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

func buildStructValue(structVO *ValueObject) {
	// add fields
	rv := structVO.rv
	for i := 0; i < rv.NumField(); i++ {
		sf := rv.Field(i)
		fieldObj := buildTypeObject(sf.Type)
		if fieldObj.tp != Invalid {
			fieldObj.sf = &sf
			fieldObj.name = sf.Name
			structTO.fields = append(structTO.fields, fieldObj)
		}
	}
}
