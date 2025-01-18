```
    // 不用函數，直接反射，上下兩行的結果也是一樣的
	var b int
	t, v = reflect.TypeOf(b), reflect.ValueOf(b)
	tk = t.Kind()
	vk = v.Kind()
	vb = v.IsValid()
	zb = v.IsZero()
	//nb = v.IsNil()	//panic
	fmt.Printf("var b int | %v | %v | %v | %v | %v | %v | %v \n", t, v, tk, vk, vb, zb, pn)

	a = b
	t, v = reflect.TypeOf(a), reflect.ValueOf(a)
	tk = t.Kind()
	vk = v.Kind()
	vb = v.IsValid()
	zb = v.IsZero()
	//nb = v.IsNil() //panic
	fmt.Printf("a = b | %v | %v | %v | %v | %v | %v | %v \n", t, v, tk, vk, vb, zb, pn)
```

| 变量赋值 | t-of | v-of | t-kind | v-kind | valid | zero | nil |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `var a interface{}` | `<nil>` | `<invalid reflect.Value>` | `!` | `invalid` | `false` | `!` | `!` |
| `var b int` | `int` | `0` | `int` | `int` | `true` | `true` | `!` |
| `a = b` | `int` | `0` | `int` | `int` | `true` | `true` | `!` |
| `var c *int` | `*int` | `<nil>` | `ptr` | `ptr` | `true` | `true` | `true` |
| `a = c` | `*int` | `<nil>` | `ptr` | `ptr` | `true` | `true` | `true` |
| `var d []int` | `[]int` | `[]` | `slice` | `slice` | `true` | `true` | `true` |
| `a = d` | `[]int` | `[]` | `slice` | `slice` | `true` | `true` | `true` |
| `var e MyInterface` | `<nil>` | `<invalid reflect.Value>` | `!` | `invalid` | `false` | `!` | `!` |
| `a = e` | `<nil>` | `<invalid reflect.Value>` | `!` | `invalid` | `false` | `!` | `!` |
| `var f MyInt` | `main.MyInt` | `0` | `int` | `int` | `true` | `true` | `!` |
| `a = f` | `main.MyInt` | `0` | `int` | `int` | `true` | `true` | `!` |


```
    //至於如何判斷 var a 定義的時候是interface{} ？，只能使用斷言

    if _, ok := a.(interface{}); ok {
        //是interface{}類型
    }else{
        //不是interface{}類型
    }
```

| 變量賦值 | t-of | v-of | t-kind | v-kind | valid | zero | nil |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `interface{}` | `<nil>` | `<invalid reflect.Value>` | `!` | `invalid` | `false` | `!` | `!` |
| `int` | `int` | `0` | `int` | `int` | `true` | `true` | `!` |
| `string` | `string` | `""` | `string` | `string` | `true` | `true` | `!` |
| `struct{}` | `struct {}` | `{}` | `struct` | `struct` | `true` | `true` | `!` |
| `User` | `main.User` | `{0 }` | `struct` | `struct` | `true` | `true` | `!` |
| `MyInt` | `main.MyInt` | `0` | `int` | `int` | `true` | `true` | `!` |
| `*int` | `*int` | `<nil>` | `ptr` | `ptr` | `true` | `true` | `true` |
| `*User` | `*main.User` | `<nil>` | `ptr` | `ptr` | `true` | `true` | `true` |
| `&User{}` | `*main.User` | `&{0 }` | `ptr` | `ptr` | `true` | `false` | `false` |
| `[]int` | `[]int` | `[]` | `slice` | `slice` | `true` | `true` | `true` |
| `map` | `map[interface {}]interface {}` | `map[]` | `map` | `map` | `true` | `true` | `true` |
| `type Action interface {Run()}` | `<nil>` | `<invalid reflect.Value>` | `!` | `invalid` | `false` | `!` | `!` |
| `User is Action` | `*main.User` | `&{0 }` | `interface` | `interface` | `true` | `false` | `false` |