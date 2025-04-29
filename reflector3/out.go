package reflector3

import (
	"fmt"
	"reflect"
	"strings"
)

func PrintDef(t reflect.Type) {
	printDefinition(t, "reflect.Type")
}

func PrintValDef(v *reflect.Value) bool {
	if !v.IsValid() {
		// panic if !v.IsValid() call: v.Type(), v.CanInterface(), v.IsNil(), v.IsZero(), v.Interface())
		fmt.Printf("reflect.Value.Kind() is invalid\n")
		return false
	}
	printDefinition(v.Type(), "reflect.Value.Type()")
	return true
}

func printDefinition(t reflect.Type, desc string) {
	if t == nil {
		fmt.Printf("%s is nil\n", desc)
		return
	}
	base := fmt.Sprintf("%s: String=%s, Name=%v, Kind=%v, FullPath=%v", desc, t.String(), t.Name(), t.Kind(), t.PkgPath())
	switch t.Kind() {
	case reflect.Map:
		base += fmt.Sprintf(", Key=%v", t.Key().String())
		fallthrough
	case reflect.Array, reflect.Slice, reflect.Pointer:
		base += fmt.Sprintf(", Elem=%v", t.Elem().String())
	case reflect.Struct:
		fields := make([]string, 0)
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			fields = append(fields, field.Name)
		}
		base += fmt.Sprintf(", fields=%v", fields)
	default:
		fmt.Printf("%s\n", base)
	}
}

func PrintValue(v *reflect.Value) {
	if !PrintValDef(v) {
		return
	}
	base := fmt.Sprintf("reflect.Value: String=%s, Kind=%v, CanInterface=%v, Value=%v", v.String(), v.Kind(), v.CanInterface(), v.Interface())

	switch v.Kind() {
	case reflect.Pointer:
		base += fmt.Sprintf(", Elem=%v", v.Elem().String())
	case reflect.Array, reflect.Slice, reflect.String:
		base += fmt.Sprintf(", Len=%v", v.Len())
	case reflect.Map:
		base += fmt.Sprintf(", Len=%v, KeyLen=%v", v.Len(), len(v.MapKeys()))
	case reflect.Struct:
		base += fmt.Sprintf(", FieldLen=%v", v.NumField())
	}
	fmt.Printf("%s\n", base)
}

// 臨時的
func (r *Node) Print(title string) {

	fmt.Printf("var=%s, stack=%s\n", title, r.GetStackName())
	PrintDef(r.obj.RefTp())
	PrintValue(r.obj.RefVal())
}

// 臨時測試方法
func (n *Node) GetStackName() string {
	var stack []string
	for n := n; n != nil; n = n.parent {
		stack = append(stack, n.Name)
	}
	// 將stack反過來
	for i, j := 0, len(stack)-1; i < j; i, j = i+1, j-1 {
		stack[i], stack[j] = stack[j], stack[i]
	}
	return strings.Join(stack, ".")
}
