package reflector3

import (
	"fmt"
	"reflect"
)

func PrintDef(t reflect.Type) {
	printDefinition(t, "Def")
}

func PrintValDef(t reflect.Type) {
	printDefinition(t, "Act-Def")
}

func printDefinition(t reflect.Type, desc string) {
	if t == nil {
		fmt.Printf("%s type is nil\n", desc)
		return
	}
	fmt.Printf("%s %s:\n", desc, t.String())
	switch t.Kind() {
	case reflect.Array, reflect.Slice, reflect.Pointer:
		fmt.Printf("-- Type=%v, Kind=%v, Name=%v, FullPath=%v Elem=%v", t.String(), t.Kind(), t.Name(), t.PkgPath(), t.Elem().String())
	case reflect.Map:
		fmt.Printf("-- Type=%v, Kind=%v, Name=%v, FullPath=%v Key=%v Elem=%v", t.String(), t.Kind(), t.Name(), t.PkgPath(), t.Key().String(), t.Elem().String())
	case reflect.Struct:
		fields := make([]string, 0)
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			fields = append(fields, field.Name)
		}
		fmt.Printf("-- Type=%v, Kind=%v, Name=%v, FullPath=%v, fields=%v", t.String(), t.Kind(), t.Name(), t.PkgPath(), fields)
	default:
		fmt.Printf("-- Type=%v, Kind=%v, Name=%v, FullPath=%v", t.String(), t.Kind(), t.Name(), t.PkgPath())
	}
	fmt.Printf("\n")
}

func PrintValue(v *reflect.Value) {
	fmt.Printf("%s Value IsValid: %v\n", v.String(), v.IsValid())
	if !v.IsValid() {
		return
	}
	// panic if !v.IsValid()
	fmt.Printf("-- CanInterface=%v, IsNil=%v, IsZero=%v, Value=%v\n", v.CanInterface(), v.IsNil(), v.IsZero(), v.Interface())

	switch v.Kind() {
	case reflect.Pointer:
		fmt.Printf("-- Kind=%v, Elem=%v", v.Kind(), v.Elem().String())
	case reflect.Array, reflect.Slice:
		fmt.Printf("-- Kind=%v, Len=%v", v.Kind(), v.Len())
	case reflect.Map:
		fmt.Printf("-- Kind=%v, Keys=%v", v.Kind(), len(v.MapKeys()))
	case reflect.Struct:
		fmt.Printf("-- Kind=%v, Fields=%v", v.Kind(), v.NumField())
	default:
		fmt.Printf("-- Kind=%v", v.Kind())
	}
	fmt.Printf("\n")
}

// 臨時的
func (r *Node) Print(title string) {

	fmt.Printf("var %s -> %s\n", title, r.GetStackName())
	PrintDef(r.obj.RefTp())
	PrintValDef(r.obj.RefVal().Type())
	PrintValue(r.obj.RefVal())
}
