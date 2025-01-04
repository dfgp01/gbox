package component

import (
	"fmt"
	"sync"
)

type (
	Id   int
	Type int

	Component interface {
		ID() Id
		Type() Type
	}

	Base struct {
		id Id
		tp Type
	}
)

func (b *Base) ID() Id        { return b.id }
func (b *Base) Type() Id      { return Id(b.tp) }
func (b *Base) Error() string { return fmt.Sprintf("exists component id:%d, type:%d", b.id, b.tp) }

type ComponentManager struct {
	mu  sync.RWMutex
	ids map[Id]Component
	tps map[Type]Component
}

func (m *ComponentManager) Add(c Component) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	//todo 可能要優化
	if _, ok := m.ids[c.ID()]; ok {
		return &Base{id: c.ID(), tp: c.Type()}
	}
	if _, ok := m.tps[c.Type()]; ok {
		return &Base{id: c.ID(), tp: c.Type()}
	}
	m.ids[c.ID()] = c
	m.tps[c.Type()] = c
	return nil
}

func (m *ComponentManager) DelFromType(tp Type) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if c, ok := m.tps[tp]; !ok {
		return
	} else {
		delete(m.tps, tp)
		delete(m.ids, c.ID())
	}
}

func (m *ComponentManager) DelFromId(id Id) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if c, ok := m.ids[id]; !ok {
		return
	} else {
		delete(m.tps, c.Type())
		delete(m.ids, id)
	}
}

func (m *ComponentManager) Del(c Component) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.ids, c.ID())
	delete(m.tps, c.Type())
}

func (m *ComponentManager) GetById(id Id) Component {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.ids[id]
}

func (m *ComponentManager) GetByType(tp Type) Component {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.tps[tp]
}

func NewComponentManager() *ComponentManager {
	return &ComponentManager{
		ids: make(map[Id]Component),
		tps: make(map[Type]Component),
	}
}

var inst *ComponentManager

func init() {
	inst = NewComponentManager()
}

func GlobalComponentManager() *ComponentManager {
	return inst
}
