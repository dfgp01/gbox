package ecs

import "sort"

// System 统一系统接口：方法为 Update。
// 所有业务系统实现该方法；调度相关属性（优先级、软开关）由基类 BaseSystem 提供，
// 业务系统内嵌 BaseSystem 即可获得。
type System interface {
	Update(w World, dt float32)
}

// BaseSystem 系统基类（预留占位）：提供优先级与软开关属性。
// 业务系统内嵌使用，例如：
//
//	type MoveSystem struct{ ecs.BaseSystem }
//	func (s *MoveSystem) Update(w ecs.World, dt float32) { ... }
//
// 约定：运行期只通过 Enable/Disable 切换软开关。
type BaseSystem struct {
	priority int  // 优先级：越大越先执行（同优先级保持注册顺序）
	enable   bool // 软开关：false 跳过（默认启用，见 NewBaseSystem）
}

// NewBaseSystem 创建一个默认启用的系统基类。
func NewBaseSystem() *BaseSystem { return &BaseSystem{enable: true} }

// Priority 返回优先级（越大越先执行）。
func (b *BaseSystem) Priority() int { return b.priority }

// SetPriority 设置优先级。
func (b *BaseSystem) SetPriority(p int) { b.priority = p }

// Enable 启用系统（运行期软开关）。
func (b *BaseSystem) Enable() { b.enable = true }

// Disable 禁用系统（运行期软开关，代替运行期移除）。
func (b *BaseSystem) Disable() { b.enable = false }

// Enabled 系统是否处于启用状态。
func (b *BaseSystem) Enabled() bool { return b.enable }

// SystemManager 按优先级降序执行所有系统。
// 构建期：Add 后立即稳定排序；运行期：只读遍历，仅按 Enabled 标记跳过。
type SystemManager struct {
	systems []System
}

// NewSystemManager 创建一个空调度器。
func NewSystemManager() *SystemManager {
	return &SystemManager{}
}

// Add 添加系统（任意时刻可调用），并立即按优先级降序稳定排序。
// 若系统内嵌了 BaseSystem（实现了 Priority/Enable），则默认启用并按优先级排序；
// 否则按默认优先级 0、默认启用处理。
func (m *SystemManager) Add(s System) {
	if s == nil {
		return
	}
	if e, ok := s.(interface{ Enable() }); ok {
		e.Enable() // 添加即启用
	}
	m.systems = append(m.systems, s)
	sort.SliceStable(m.systems, func(i, j int) bool {
		return sysPriority(m.systems[i]) > sysPriority(m.systems[j])
	})
}

// sysPriority 读取系统优先级（未实现 Priority 时视为 0）。
func sysPriority(s System) int {
	if p, ok := s.(interface{ Priority() int }); ok {
		return p.Priority()
	}
	return 0
}

// Update 按优先级顺序执行所有启用（Enabled）的系统。
// 热循环不含任何锁/原子操作；线程安全由外层调用方保证。
func (m *SystemManager) Update(w World, dt float32) {
	for _, s := range m.systems {
		if en, ok := s.(interface{ Enabled() bool }); ok && !en.Enabled() {
			continue // 软开关：跳过
		}
		s.Update(w, dt)
	}
}
