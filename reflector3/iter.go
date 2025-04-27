package reflector3

import (
	"strings"
)

type (
	Node struct {
		Tp     Type   //類型
		Name   string //struct, struct-field才有
		Index  int    //slice、array才有
		Key    any    //map才有
		Value  any    //實際内容
		parent *Node  //上級節點

		obj Object //當前對象，臨時
	}

	Caller struct {
		currNode *Node  //當前節點
		currObj  Object //當前對象
		handler  func(n *Node)
	}
)

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

func (c *Caller) CanForward() bool {
	return isTypeIn(c.currObj.DefType(), Slice, Map, Struct, Pointer) && c.currObj.Value() != nil
}

// forward 進入下一層級，根據類型進行不同處理
func (c *Caller) Forward() {

	//是否有效對象
	if c.currObj.DefType() == Invalid {
		return
	}

	// 可以進入下一層的類型，基礎類型直接退回
	if !c.CanForward() {
		return
	}

	// 進入下一層
	c.stepIn(c.currObj)
}

func (c *Caller) stepIn(currObj Object) {
	switch obj := currObj.(type) {
	case *InvalidObject:
		c.step(obj, &Node{
			Tp:    Invalid,
			Name:  obj.refTp.Name() + "(" + obj.refTp.String() + ")",
			Index: -1,
			Key:   nil,
			Value: nil,
		})

	case *BaseRefObject:
		c.step(obj, &Node{
			Tp:    obj.DefType(),
			Name:  obj.refTp.Name() + "(" + obj.refTp.String() + ")",
			Index: -1,
			Key:   nil,
			Value: obj.Value(),
		})

	case *PtrRefObject:
		c.step(obj, &Node{
			Tp:    obj.DefType(),
			Name:  obj.refTp.Name() + "(*" + obj.refTp.String() + ")",
			Index: -1,
			Key:   nil,
			Value: obj.Value(),
		})

	case *StructRefObject:
		for _, field := range obj.fields {
			c.step(obj, &Node{
				Tp:    field.DefType(),
				Name:  field.fieldDef.Name + " - " + obj.refTp.Name() + "(*" + obj.refTp.String() + ")",
				Index: -1,
				Key:   nil,
				Value: field.Value(),
			})
		}
	case *SliceRefObject:
		for i, elem := range obj.elems {
			c.step(obj, &Node{
				Tp:    elem.DefType(),
				Name:  obj.refTp.Name() + "(" + obj.refTp.String() + ")",
				Index: i,
				Key:   nil,
				Value: elem.Value(),
			})
		}
	case *MapRefObject:
		for k, v := range obj.keys {
			c.step(obj, &Node{
				Tp:    v.DefType(),
				Name:  obj.refTp.Name() + "(" + obj.refTp.String() + ")",
				Index: -1,
				Key:   v.Value(),
				Value: obj.elems[k].Value(),
			})
		}
	default:
		return
	}
}

func (c *Caller) step(currObj Object, currNode *Node) {
	currNode.obj = currObj
	c.currObj = currObj

	// add to stack rear
	parent := c.currNode
	currNode.parent = parent

	c.handler(currNode)
	// pop from stack
	c.currNode = parent
}

func Iterator(v interface{}, handler func(n *Node)) {

	// 若 v 是any定義，那麽rt為空
	root := NewRefObject(v)

	// new iter
	caller := &Caller{
		handler: handler,
	}

	caller.stepIn(root)
}
