package reflector3

type (
	Node struct {
		Tp     Type   //類型
		Name   string //reflect.Type.Name(), 一般用於struct, struct-field
		Index  int    //上級是slice才有，其餘為-1
		Key    any    //上級是map才有，其餘為nil
		Len    int    //slice、string才有，其餘為-1
		Value  any    //實際内容
		parent *Node  //上級節點，root的parent為nil

		obj Object //當前對象引用
	}

	Caller struct {
		currNode *Node //當前節點
		handler  func(n *Node)
	}
)

// forward 進入下一層級，根據類型進行不同處理
// newIfNil 如果沒有value，是否創建新值
func (c *Caller) Forward(newIfNil bool) {

	var (
		obj = c.currNode.obj
	)

	//是否有效對象
	if obj.DefType() == Invalid {
		return
	}

	// 是否可以進入下一層
	if !isTypeIn(obj.DefType(), Slice, Map, Struct, Pointer) {
		return
	}
	// 判斷value情況
	if obj.Value() == nil && !newIfNil {
		return
	}

	// 進入下一層
	c.stepInNext()
}

// 進入當前節點
func (c *Caller) stepInCurrent() {
	switch obj := c.currNode.obj.(type) {
	case *AnyObject:
		c.stepIn(obj, &Node{
			Tp:    Any,
			Name:  "Null",
			Index: -1,
			Key:   nil,
			Value: nil,
		})
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

func (c *Caller) stepInNext() {
	switch obj := c.currNode.obj.(type) {
	case *AnyObject:
		c.step(obj, &Node{
			Tp:    Any,
			Name:  "Null",
			Index: -1,
			Key:   nil,
			Value: nil,
		})
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

func (c *Caller) step(currObj Object, newNode *Node) {
	newNode.obj = currObj

	// add to stack rear
	parent := c.currNode
	c.currNode = newNode

	// 防止root的parent是自己
	if currObj != parent.obj {
		newNode.parent = parent
	}

	c.handler(newNode)
	// pop from stack
	c.currNode = parent
}

func Iterator(v interface{}, handler func(n *Node)) {

	// 若 v 是any定義，那麽rt為空
	root := NewRefObject(v)

	// new iter
	caller := &Caller{
		handler: handler,
		currNode: &Node{
			obj: root,
		},
	}

	caller.stepInNext()
}
