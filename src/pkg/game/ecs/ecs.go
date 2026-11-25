// Package game 提供游戏领域（ECS/FSM）对外 API 层（pkg/game）。
//
// 分层约定：
//   - 公共定义（接口/枚举/公共类）在 def/game；
//   - 本包（pkg/game）提供对外工厂与流程：入参出参以 def/game 为主，
//     底层委托 internal/game/ecs 的默认实现；
//   - 外部使用者 import gbox/pkg/game（API）与 gbox/def/game（类型）即可。
package game

import (
	"context"
	"time"

	def "gbox/def/game/ecs"
	"gbox/internal/game/ecs"
	internal "gbox/internal/game/ecs"
)

// ---- 工厂：World / 状态机系统 ----

// NewWorld 创建一个默认 World 实例（Archetype 列式存储的默认实现）。
// 使用全局默认组件映射器（DefaultComponentMapper）作为组件解析器，
// 组件注册一次即可，所有 World 副本共用同一组件组。
func NewWorld() def.World {
	return ecs.NewWorldWithRegistry(componentMapper)
}

// NewWorldFrom 使用指定组件映射器作为组件解析器创建 World（组件组由其决定）。
// 传 nil 时使用全局默认组件映射器。
func NewWorldFrom(m *ComponentMapper) def.World {
	if m == nil {
		m = componentMapper
	}
	return ecs.NewWorldWithRegistry(m)
}

// StateMachineSystem 状态机系统接口：在 def.System 基础上，额外提供运行期调整优先级的能力。
// 说明：def/game.System 只暴露只读的 Priority()；需要运行期调整优先级时使用本接口。
type StateMachineSystem interface {
	def.System
	// SetPriority 设置优先级（越大越先执行，同优先级保持注册顺序）。
	SetPriority(p int)
}

// NewStateMachineSystem 创建驱动指定状态机的系统（实现 StateMachineSystem）。
// 自带优先级与软开关，可 SetPriority / Enable / Disable 进行调度。
func NewStateMachineSystem(m *def.StateMachine) StateMachineSystem {
	return internal.NewStateMachineSystem(m)
}

// GameStart 开始游戏主循环。
// 新开一条线程（goroutine）运行主循环，确保内部单线程运行，无法从其他地方进入，
// 以免破坏线程安全。暂定固定 60fps 定时 tick，每帧依次执行：
//
//	beforeUpdate(dt) → sysGroup.Update(world, dt) → afterUpdate(dt)
//
// 参数说明：
//   - c: 上下文控制，建议用超时控制的，外部可以通过 cancel 取消游戏主循环，
//     也可以通过最大超时结束，避免死线程。
//   - sysGroup: 系统组；传 nil 时使用全局默认系统组（DefaultSystemGroup）。
//   - beforeUpdate: 帧前钩子（系统组更新前执行，可为 nil），dt 为固定帧间隔（秒）。
//   - afterUpdate: 帧后钩子（系统组更新后执行，可为 nil），dt 为固定帧间隔（秒）。
func GameStart(c context.Context, sysGroup *SystemGroup, beforeUpdate func(dt float32), afterUpdate func(dt float32)) {
	if sysGroup == nil {
		sysGroup = DefaultSystemGroup()
	}
	if c == nil {
		c = context.Background()
	}

	// 暂定写死 60fps：固定帧间隔与 tick 间隔一致
	const fps = 60
	frameDt := float32(1) / fps
	interval := time.Second / fps

	w := NewWorld() // 默认世界（与全局默认组件组共用注册）

	// 新开线程（goroutine）运行主循环，内部保持单线程
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-c.Done():
				return // 外部 cancel 或超时结束
			case <-ticker.C:
				if beforeUpdate != nil {
					beforeUpdate(frameDt)
				}
				sysGroup.Update(w, frameDt)
				if afterUpdate != nil {
					afterUpdate(frameDt)
				}
			}
		}
	}()
}
