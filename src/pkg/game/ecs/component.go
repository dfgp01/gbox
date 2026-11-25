package game

import (
	def "gbox/def/game"
	"gbox/internal/game/ecs"
	"reflect"
)

// ---- 组件映射器：组件静态注册（组件组） ----

// ComponentMapper 组件映射器：负责组件的静态注册——map 映射、ID 生成、注册时类型校验、
// 查找组件等，全部归它管。一个组件映射器对应一个“组件组”。
//
// 按业务分类，不同的组件映射器管理不同的组件组；一个游戏整体可以有多个组件组。
// 默认提供全局唯一的组件映射器（DefaultComponentMapper），其映射可被所有 World 副本
// 共用（组件注册一次即可）；也可通过 GetComponentMapper(key) 按 int key 获取多个
// 组件映射器，分别管理不同的组件组。
//
// ComponentMapper 同时实现了 internal 的组件解析契约（RegisterComponent / IDOf / TypeOf），
// 可直接作为 World 的组件解析器（见 NewWorld / NewWorldFrom）。
type ComponentMapper struct {
	byType map[reflect.Type]def.ComponentID // 组件类型 -> 组件 ID
	byID   map[def.ComponentID]reflect.Type // 组件 ID -> 组件类型
}

// RegisterComponent 注册一个组件类型，返回其组件 ID（幂等）。
// 注册时校验 v 必须是 struct 值类型（不允许指针/nil）；数量上限由调用方负责（默认 64）。
func (m *ComponentMapper) RegisterComponent(v any) def.ComponentID {
	t := reflect.TypeOf(v)
	if t == nil || t.Kind() != reflect.Struct {
		panic("game: 组件必须是值类型，不允许传指针或 nil")
	}
	if id, ok := m.byType[t]; ok {
		return id
	}
	id := def.ComponentID(len(m.byID)) + 1 // 从 1 开始分配，0 保留为无效哨兵
	m.byType[t] = id
	m.byID[id] = t
	return id
}

// RegisterComponents 批量注册多个组件类型，返回对应的组件 ID 列表（顺序与入参一致）。
func (m *ComponentMapper) RegisterComponents(vs ...any) []def.ComponentID {
	ids := make([]def.ComponentID, 0, len(vs))
	for _, v := range vs {
		ids = append(ids, m.RegisterComponent(v))
	}
	return ids
}

// IDOf 查找组件类型对应的 ID；未注册返回 (ComponentIDInvalid, false)。
func (m *ComponentMapper) IDOf(v any) (def.ComponentID, bool) {
	t := reflect.TypeOf(v)
	if t == nil || t.Kind() != reflect.Struct {
		return def.ComponentIDInvalid, false
	}
	id, ok := m.byType[t]
	return id, ok
}

// TypeOf 查找组件 ID 对应的类型；未注册返回 nil。
func (m *ComponentMapper) TypeOf(id def.ComponentID) reflect.Type {
	return m.byID[id]
}

// Len 已注册组件类型的数量。
func (m *ComponentMapper) Len() int {
	return len(m.byID)
}

// NewComponentMapper 创建一个新的组件映射器（独立组件组）。
func NewComponentMapper() *ComponentMapper {
	return &ComponentMapper{
		byType: make(map[reflect.Type]def.ComponentID),
		byID:   make(map[def.ComponentID]reflect.Type),
	}
}

// componentMapper 全局默认组件映射器（全局注册一次即可，所有 World 副本共用）。
var componentMapper = NewComponentMapper()

// componentMappers 多个组件映射器（可选，0=默认组件映射器，其他根据需要自己添加）。
var componentMappers []*ComponentMapper

func AddComponentMapper(m *ComponentMapper) {
	componentMappers = append(componentMappers, m)
}

func init() {
	AddComponentMapper(componentMapper)
}

// RegisterComponent 在全局默认组件映射器中注册组件类型 T，返回其组件 ID（幂等）。
func RegisterComponent[T any]() def.ComponentID {
	var zero T
	return componentMapper.RegisterComponent(zero)
}

// RegisterComponentIn 在指定组件映射器 m 中注册组件类型 T，返回其组件 ID（幂等）。
// 用于多组件组；传 nil 时使用全局默认组件映射器。
func RegisterComponentIn[T any](m *ComponentMapper) def.ComponentID {
	if m == nil {
		m = componentMapper
	}
	var zero T
	return m.RegisterComponent(zero)
}

// Add 为实体添加组件 T（零值），返回指向该组件的指针（可直接赋值初始化）。
// 组件类型未注册会自动注册；实体已有该组件则不覆盖；实体不存在时返回 nil。
func Add[T any](w def.World, e def.EntityID) *T {
	return ecs.Add[T](w, e)
}
