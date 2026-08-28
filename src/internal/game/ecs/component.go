package ecs

import game "gbox/def/game"

// 本文件存放“泛型便捷函数”的默认实现；对外由 pkg/game 包装提供。

// Add 为实体添加组件 T（零值），返回指向该组件的指针（可直接赋值初始化）。
// 组件类型未注册会自动注册；实体已有该组件则不覆盖；实体不存在时返回 nil。
func Add[T any](w game.World, e game.EntityID) *T {
	var zero T
	id := w.Add(e, zero)
	p := w.Get(e, id)
	if p == nil {
		return nil
	}
	return p.(*T)
}
