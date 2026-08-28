package game

import (
	"sort"

	def "gbox/def/game"
)

// BaseSystem 系统基类（公共类）：提供优先级与软开关。
// 业务系统内嵌使用，例如：
//
//	type MoveSystem struct{ game.BaseSystem }
//	func (s *MoveSystem) Update(w game.World, dt float32) { ... }
//
// 约定：运行期只通过 Enable/Disable 切换软开关。
type BaseSystem struct {
	priority int  // 优先级：越大越先执行（同优先级保持注册顺序）
	disable  bool // 软开关
}

// NewBaseSystem 创建一个默认启用的系统基类。
func NewBaseSystem(priority int) *BaseSystem { return &BaseSystem{priority: priority} }

// Priority 返回优先级（越大越先执行）。
func (b *BaseSystem) Priority() int { return b.priority }

// SetPriority 设置优先级。
func (b *BaseSystem) SetPriority(p int) { b.priority = p }

// Enable 启用系统（运行期软开关）。
func (b *BaseSystem) Enable() { b.disable = false }

// Disable 禁用系统（运行期软开关，代替运行期移除）。
func (b *BaseSystem) Disable() { b.disable = true }

// Enabled 系统是否处于启用状态。
func (b *BaseSystem) Enabled() bool { return !b.disable }

// SystemGroup 系统组：仿照 ComponentMapper，专门维护一组 System。
// 可添加系统并按优先级降序稳定排序；一个系统组对应一组系统的调度集合。
//
// 默认提供全局唯一的系统组（DefaultSystemGroup）；也可通过 GetSystemGroup(key)
// 按 int key 获取多个系统组，用于不同模块/玩法各自调度。
type SystemGroup struct {
	systems []def.System // 按优先级降序排列
}

// NewSystemGroup 创建一个空的系统组。
func NewSystemGroup() *SystemGroup {
	return &SystemGroup{}
}

// systemGroup 全局默认系统组（全局独一份）。
var systemGroup = NewSystemGroup()

// systemGroups 按 int key 区分的多个系统组（可选，用于多组调度）。
var systemGroups = make(map[int]*SystemGroup)

// DefaultSystemGroup 返回全局默认系统组（全局独一份）。
func DefaultSystemGroup() *SystemGroup {
	return systemGroup
}

// GetSystemGroup 获取指定 key 的系统组；不存在则创建并缓存。
// key 由使用方约定，例如按模块划分不同的系统组。
func GetSystemGroup(key int) *SystemGroup {
	if g, ok := systemGroups[key]; ok {
		return g
	}
	g := NewSystemGroup()
	systemGroups[key] = g
	return g
}

// Add 添加一个系统，并立即按优先级降序稳定排序（同优先级保持添加顺序）。
// 添加即启用（Enable）。系统需实现 def.System（通常内嵌 BaseSystem）。
func (g *SystemGroup) Add(s def.System) {
	if s == nil {
		return
	}
	s.Enable() // 添加即启用
	g.systems = append(g.systems, s)
	sort.SliceStable(g.systems, func(i, j int) bool {
		return g.systems[i].Priority() > g.systems[j].Priority()
	})
}

// Update 按优先级顺序执行所有启用（Enabled）的系统。
// 热循环不含任何锁/原子操作；线程安全由外层调用方保证。
func (g *SystemGroup) Update(w def.World, dt float32) {
	for _, s := range g.systems {
		if !s.Enabled() {
			continue // 软开关：跳过
		}
		s.Update(w, dt)
	}
}

// Len 已添加的系统数量。
func (g *SystemGroup) Len() int {
	return len(g.systems)
}
