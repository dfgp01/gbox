package internal

import (
	"fmt"
	"reflect"
)

//	種草錨點
//	問題：
//		rt := reflect.ValueOf(v).Type()
//		rtt := reflect.TypeOf(v1)
//		is rt == rtt ??????

//		reflect.TypeOf(v1).Kind()，裏面的interface究竟指的是什麽

//	1、注意空接口問題，是 interface 還是 interface{}

// 自定義接口
type Action interface {
	Run()
}

// 自定義類型
type MyInt int

// 自定義結構
type User struct {
	Id   int
	Name string
}

func (u *User) Run() {
	u.Name = "run"
}

func RunNative() {

	//	欄位：變量賦值、t-of、v-of、t-kind、v-kind、valid、zero、nil
	fmt.Println("變量賦值 | | t-name | t-of | v-of | t-kind | v-kind | valid | zero | nil")

	// interface{}變量
	var a interface{}
	reflectVar("interface{}", a)

	// 基礎類型
	var b int
	reflectVar("int", b)
	var b2 string
	reflectVar("string", b2)
	var b3 struct{}
	reflectVar("struct{}", b3)
	var b4 User
	reflectVar("User", b4)
	var b5 MyInt
	reflectVar("MyINt", b5)

	// 指針類型
	var c *int
	reflectVar("*int", c)
	var c2 *User
	reflectVar("*User", c2)
	var c3 = &User{}
	reflectVar("&User{}", c3)

	// 數組/切片
	var d []int
	reflectVar("[]int", d)

	// map
	var e map[interface{}]interface{}
	reflectVar("map", e)

	// 自定義接口
	var f Action
	reflectVar("type Action interface {Run()}", f)
	var f1 Action
	f1 = Action(&User{})
	reflectVar("User is Action", f1)

}

func reflectVar(title string, a interface{}) {
	//從目前情況來看
	//儅	interface{}、type interface 時，t-of=<nil>，v-of=<invalid>, v-kind=invalid, valid=false，如何區分兩者？

	var (
		t          reflect.Type
		v          reflect.Value
		tk, vk     reflect.Kind
		vb, zb, nb bool
		pn         = "!"
	)

	t, v = reflect.TypeOf(a), reflect.ValueOf(a)
	vk = v.Kind()
	vb = v.IsValid()

	// 變量為 interface{}、type interface時，t==nil，vk=invalid，此時isZero()和isNil()不可用
	// 判斷 !v.IsValid(), v.Kind() == reflect.Invalid, reflect.TypeOf(a) == nil 都可以
	if !vb {
		fmt.Printf("%s | %v | %v | %v | %v | %v | %v | %v | %v \n", title, pn, t, v, pn, vk, vb, pn, pn)
		return
	}

	//其餘情況，tk都有值
	tk = t.Kind()

	//變量為 int 之類的基礎類型時，isNil()不可用
	if isBase(tk) {
		zb = v.IsZero()
		fmt.Printf("%s | %v | %v | %v | %v | %v | %v | %v | %v \n", title, t.Name(), t, v, tk, vk, vb, zb, pn)
		return
	}

	//其他類型
	zb = v.IsZero()
	nb = v.IsNil()
	fmt.Printf("%s | %v | %v | %v | %v | %v | %v | %v | %v \n", title, t.Name(), t, v, tk, vk, vb, zb, nb)

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

func ReflectTest2() {
	var a interface{}
	var b int = 5
	var c string = "hi"

	ta, va := reflect.TypeOf(a), reflect.ValueOf(a)
	tb, vb := reflect.TypeOf(b), reflect.ValueOf(b)
	tc, vc := reflect.TypeOf(c), reflect.ValueOf(c)

	fmt.Println("1", ta, tb, tc, va, vb, vc)

	a = b
	ta2, va2 := reflect.TypeOf(a), reflect.ValueOf(a)
	fmt.Println("2", ta == ta2, ta2 == tb, va2 == vb)

	a = c
	ta3, va3 := reflect.TypeOf(a), reflect.ValueOf(a)
	fmt.Println("3", ta == ta3, ta3 == tc, va3 == vc)
}

type MyStruct struct {
	FieldA int
	FieldB interface{}
}

func ReflectTest3() {
	var a MyStruct
	ta1, va1 := reflect.TypeOf(a).Field(1), reflect.ValueOf(a).Field(1)
	fmt.Println("1", ta1, va1, va1.Type())

	a.FieldB = 100
	_, va2 := reflect.TypeOf(a).Field(1), reflect.ValueOf(a).Field(1)
	fmt.Println("2", va1.Type() == va2.Type(), va1.Type().Kind(), va1.Kind(), va2.Type().Kind(), va2.Kind())

	a.FieldB = "100"
	ta3, va3 := reflect.TypeOf(a).Field(1), reflect.ValueOf(a).Field(1)
	fmt.Println("3", ta1.Type == ta3.Type, va1.Type() == va3.Type())
}
