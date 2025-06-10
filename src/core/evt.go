package core

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

type (
	EventType  string
	HandleFunc func(event *Event) bool
	Middleware func(ctx context.Context, evt *Event, next func(ctx context.Context, evt *Event) error) error

	Event struct {
		Data            interface{}     // 自定義的數據
		StopPropagation bool            // 冒泡终止标记
		tp              EventType       // 事件類型
		ctx             context.Context // 上下文
	}

	EventHandler struct {
		handle   HandleFunc
		priority int // 监听器优先级
	}

	EventListener struct {
		in       chan *Event
		handlers []*EventHandler
		mdChain  func(ctx context.Context, evt *Event) error
		action   func(*Event, []HandleFunc) error
	}

	EventManager struct {
		listeners   map[EventType]*EventListener
		middlewares []Middleware
		pool        sync.Pool
	}
)

func (e *Event) Type() EventType {
	return e.tp
}

var (
	eventManager *EventManager
	initOnce     sync.Once
)

func GetEventManager() *EventManager {
	initOnce.Do(func() {
		eventManager = &EventManager{
			listeners: make(map[EventType]*EventListener),
			pool: sync.Pool{
				New: func() interface{} {
					return &Event{}
				},
			},
		}
	})
	return eventManager
}

func (em *EventManager) Use(md ...Middleware) {
	em.middlewares = append(em.middlewares, md...)
}

func (em *EventManager) RegisterListener(eventType EventType, handler HandleFunc, priority ...int) {
	var p int
	if len(priority) > 0 {
		p = priority[0]
	}
	eh := &EventHandler{handle: handler, priority: p}
	if listener, exists := em.listeners[eventType]; exists {
		listener.handlers = append(listener.handlers, eh)
		// 按优先级排序
		sort.Slice(listener.handlers, func(i, j int) bool {
			return listener.handlers[i].priority > listener.handlers[j].priority
		})
	} else {
		em.listeners[eventType] = &EventListener{
			in:       make(chan *Event, 10),
			handlers: []*EventHandler{eh},
		}
	}
}

func (em *EventManager) SetListenerAction(eventType EventType, action func(*Event, []HandleFunc) error) {
	if listener, exists := em.listeners[eventType]; exists {
		listener.action = action
	}
}

func (em *EventManager) SendEvent(eventType EventType, data interface{}, ctx context.Context) {
	event := em.pool.Get().(*Event)
	event.tp = eventType
	event.Data = data
	event.ctx = ctx // 将上下文存储在事件对象中
	event.StopPropagation = false

	listener, exists := em.listeners[eventType]
	if exists {
		listener.in <- event
	} else {
		em.pool.Put(event) // 如果没有监听器，放回池中
	}
}

func (em *EventManager) Start(ctx context.Context) {
	for _, listener := range em.listeners {
		em.buildMdChain(listener)
		go func(ls *EventListener) {
			for {
				select {
				case <-ctx.Done():
					// 上下文完成，停止处理事件
					fmt.Println("ctx is done")
					return
				case event := <-ls.in:
					em.Do(event, ls)
				}
			}
		}(listener)
	}
}

func (em *EventManager) buildMdChain(ls *EventListener) {

	if ls.action == nil {
		ls.action = Action2
	}

	var hFuncs []HandleFunc
	for _, h := range ls.handlers {
		hFuncs = append(hFuncs, h.handle)
	}

	//包裝listener處理器到棧尾
	nextCall := func(ctx context.Context, evt *Event) error {
		return ls.action(evt, hFuncs)
	}

	for i := len(em.middlewares) - 1; i >= 0; i-- {
		//當前md即尾部
		md := em.middlewares[i]
		// 创建一个新的闭包实例（kimi很輕鬆的解決了閉包地址問題呢），原則就是多用新變量
		next := nextCall
		//包裝尾部call給前一個md
		nextCall = func(ctx context.Context, evt *Event) error {
			return md(ctx, evt, next)
		}
	}
	ls.mdChain = nextCall
}

func (em *EventManager) Do(evt *Event, ls *EventListener) {
	defer em.pool.Put(evt)
	if err := ls.mdChain(evt.ctx, evt); err != nil {
		fmt.Println("Error handling event:", err)
	}
}

// 并行處理
func Action1(event *Event, handlers []HandleFunc) {
	var wg sync.WaitGroup
	for _, handler := range handlers {
		if event.StopPropagation {
			break // 如果事件标记为停止冒泡，则终止后续处理
		}

		wg.Add(1)
		go func(h HandleFunc, e *Event) {
			defer wg.Done()
			if !h(e) {
				e.StopPropagation = true // 如果处理函数返回 false，则标记停止冒泡
			}
		}(handler, event)

		// 检查事件的上下文是否被取消
		if event.ctx != nil && event.ctx.Err() != nil {
			break // 如果上下文被取消，则终止后续处理
		}
	}
	wg.Wait()
}

// 串行處理
func Action2(event *Event, handlers []HandleFunc) error {
	if event == nil {
		return nil
	}
	for _, handler := range handlers {
		if event.StopPropagation {
			break // 如果事件标记为停止冒泡，则终止后续处理
		}

		// 串行执行处理程序
		if !handler(event) {
			event.StopPropagation = true // 如果处理函数返回 false，则标记停止冒泡
		}
	}
	return nil
}
