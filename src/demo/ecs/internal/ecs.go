package internal

import (
	"fmt"

	"gbox/game/ecs"
)

// Pos / Vel 演示组件：纯数据，不含 EntityId。
type Pos struct {
	X, Y float32
}

type Vel struct {
	Vx, Vy float32
}

// RunECSBasic 演示 ECS 核心增删改查与掩码查询。
func RunECSBasic() {
	fmt.Println("===== ECS 基础演示 =====")

	// 创建世界（World 是接口；组件注册表按世界独立）
	w := ecs.NewWorld()

	// 在世界中注册组件类型（重复注册安全，返回同一 ID）
	idPos := ecs.RegisterComponent[Pos](w)
	idVel := ecs.RegisterComponent[Vel](w)

	// 1. Spawn：创建实体并附加组件值（任意组合、任意顺序）
	e1 := w.Spawn(Pos{X: 1, Y: 2}, Vel{Vx: 3, Vy: 4})
	e2 := w.Spawn(Pos{X: 10, Y: 20})
	fmt.Printf("spawn e1=%d e2=%d alive=%v count=%d\n",
		e1, e2, w.Alive(e1), w.Count())

	// 2. Get：通过组件 ID 取指针修改（Get 返回 *T 指针或 nil）
	p := w.Get(e1, idPos).(*Pos)
	p.X += 10
	fmt.Printf("after modify Pos(e1)=%+v\n", *(w.Get(e1, idPos).(*Pos)))

	// 3. Has / Add / Remove（均基于组件 ID）
	fmt.Printf("e1 Has[Pos]=%v Has[Vel]=%v\n", w.Get(e1, idPos) != nil, w.Get(e1, idVel) != nil)
	w.Remove(e1, idVel)
	fmt.Printf("after Remove[Vel] Has[Vel]=%v\n", w.Get(e1, idVel) != nil)
	// Add 方法：直接传组件数据，返回组件 ID；组件类型未注册会自动注册
	w.Add(e1, Vel{Vx: 5})
	fmt.Printf("after Add Vel(e1)=%+v\n", *(w.Get(e1, idVel).(*Vel)))

	// 4. Query：掩码查询（超集匹配：拥有 mask 中全部组件的实体），回调遍历
	//    这里查 Pos+Vel：e1 命中（Pos+Vel），e2 不命中（只有 Pos）。
	posVelMask := ecs.Mask(idPos, idVel)
	w.Query(posVelMask, func(e ecs.EntityID, q *ecs.Query) bool {
		pos := q.Get(idPos).(*Pos) // Get 返回 *T 指针，可原位修改
		vel := q.Get(idVel).(*Vel)
		pos.X += vel.Vx // 修改会写回存储
		pos.Y += vel.Vy
		fmt.Printf("query entity=%d pos=(%.1f,%.1f) vel=(%.1f,%.1f)\n",
			e, pos.X, pos.Y, vel.Vx, vel.Vy)
		return true
	})
	// 验证写回
	fmt.Printf("after query Pos(e1)=%+v\n", *(w.Get(e1, idPos).(*Pos)))

	// 5. Despawn：销毁实体
	w.Despawn(e1)
	fmt.Printf("after despawn e1 alive=%v Has[Pos]=%v count=%d\n",
		w.Alive(e1), w.Get(e1, idPos) != nil, w.Count())

	fmt.Println()
}
