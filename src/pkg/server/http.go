package server

import (
	"context"
	"fmt"
	"gbox/common"
	isrv "gbox/internal/server"
	"sync"
)

// ---------------------------------------------------------------------------
// 类型定义 — 纯公共接口，不依赖任何框架
// ---------------------------------------------------------------------------

// Context 是 HTTP 请求上下文接口，由 common 层定义，internal 层实现。
type Context = common.HttpContext

// Handler 是通用的 HTTP 处理器签名。
type Handler func(ctx Context)

// ---------------------------------------------------------------------------
// Route — 路由定义
// ---------------------------------------------------------------------------

// Route 定义一条 HTTP 路由。Method 支持 GET / POST / PUT / DELETE / PATCH /
// HEAD / OPTIONS / ANY。
type Route struct {
	Method  string
	Path    string
	Handler Handler
}

// ---------------------------------------------------------------------------
// HttpServerConfig — 对外暴露的配置别名
// ---------------------------------------------------------------------------

// HttpServerConfig HTTP 服务配置，可直接从 JSON 解析。
type HttpServerConfig = common.HttpServerConfig

// ---------------------------------------------------------------------------
// HttpServerComponent — 多实例 HTTP 服务管理器
// ---------------------------------------------------------------------------

// HttpServerComponent 管理多个命名 HttpServer 实例。
// 使用全局单例 Default 即可，无需自行创建。
type HttpServerComponent struct {
	mu      sync.RWMutex
	servers map[string]*isrv.HttpServer
}

// NewHttpServerComponent 创建一个组件实例。
func NewHttpServerComponent() *HttpServerComponent {
	return &HttpServerComponent{
		servers: make(map[string]*isrv.HttpServer),
	}
}

// Default 是全局默认的 HttpServerComponent 单例。
var Default = NewHttpServerComponent()

// ---------------------------------------------------------------------------
// 实例管理
// ---------------------------------------------------------------------------

// get 获取指定名称的 server，不存在时 panic。
func (c *HttpServerComponent) get(name string) *isrv.HttpServer {
	c.mu.RLock()
	srv := c.servers[name]
	c.mu.RUnlock()
	if srv == nil {
		panic(fmt.Sprintf("HttpServerComponent: server %q not found, call Load() first", name))
	}
	return srv
}

// Load 创建或替换一个命名 HttpServer 实例。
func (c *HttpServerComponent) Load(name string, cfg *HttpServerConfig) *HttpServerComponent {
	c.mu.Lock()
	c.servers[name] = isrv.NewHttpServer(cfg)
	c.mu.Unlock()
	return c
}

// Config 返回指定 server 的配置副本。
func (c *HttpServerComponent) Config(name string) HttpServerConfig {
	return c.get(name).Config()
}

// ---------------------------------------------------------------------------
// 中间件
// ---------------------------------------------------------------------------

// Use 注册全局中间件到指定 server。
func (c *HttpServerComponent) Use(name string, middleware ...Handler) *HttpServerComponent {
	srv := c.get(name)
	for _, m := range middleware {
		srv.UseHandler(m)
	}
	return c
}

// ---------------------------------------------------------------------------
// Router — 简易路由链
// ---------------------------------------------------------------------------

// Router 接收一组路由定义或中间件，自动识别并注册到指定 server：
//   - Route   → 注册为具体路由（GET/POST/PUT/DELETE 等）
//   - Handler → 注册为全局中间件
func (c *HttpServerComponent) Router(name string, items ...interface{}) *HttpServerComponent {
	srv := c.get(name)
	for _, item := range items {
		switch v := item.(type) {
		case Route:
			srv.Handle(v.Method, v.Path, v.Handler)
		case Handler:
			srv.UseHandler(v)
		}
	}
	return c
}

// ---------------------------------------------------------------------------
// 直接路由方法 — 在指定 server 上注册单条路由
// ---------------------------------------------------------------------------

func (c *HttpServerComponent) GET(name, path string, handler Handler) *HttpServerComponent {
	c.get(name).Handle("GET", path, handler)
	return c
}

func (c *HttpServerComponent) POST(name, path string, handler Handler) *HttpServerComponent {
	c.get(name).Handle("POST", path, handler)
	return c
}

func (c *HttpServerComponent) PUT(name, path string, handler Handler) *HttpServerComponent {
	c.get(name).Handle("PUT", path, handler)
	return c
}

func (c *HttpServerComponent) DELETE(name, path string, handler Handler) *HttpServerComponent {
	c.get(name).Handle("DELETE", path, handler)
	return c
}

func (c *HttpServerComponent) PATCH(name, path string, handler Handler) *HttpServerComponent {
	c.get(name).Handle("PATCH", path, handler)
	return c
}

func (c *HttpServerComponent) Any(name, path string, handler Handler) *HttpServerComponent {
	c.get(name).Handle("ANY", path, handler)
	return c
}

// ---------------------------------------------------------------------------
// 便捷路由构造器 — 返回 Route 以便在 Router() 中使用
// ---------------------------------------------------------------------------

func GET(path string, handler Handler) Route {
	return Route{Method: "GET", Path: path, Handler: handler}
}

func POST(path string, handler Handler) Route {
	return Route{Method: "POST", Path: path, Handler: handler}
}

func PUT(path string, handler Handler) Route {
	return Route{Method: "PUT", Path: path, Handler: handler}
}

func DELETE(path string, handler Handler) Route {
	return Route{Method: "DELETE", Path: path, Handler: handler}
}

func PATCH(path string, handler Handler) Route {
	return Route{Method: "PATCH", Path: path, Handler: handler}
}

func Any(path string, handler Handler) Route {
	return Route{Method: "ANY", Path: path, Handler: handler}
}

// ---------------------------------------------------------------------------
// IService 生命周期 — 代理到指定 server
// ---------------------------------------------------------------------------

func (c *HttpServerComponent) Start(ctx context.Context, name string) error {
	return c.get(name).Start(ctx)
}

func (c *HttpServerComponent) Listen(ctx context.Context, name string) error {
	return c.get(name).Listen(ctx, 0)
}

func (c *HttpServerComponent) Stop(ctx context.Context, name string) error {
	return c.get(name).Stop(ctx)
}

// ---------------------------------------------------------------------------
// 包级便捷函数 — 直接操作 Default 单例
// ---------------------------------------------------------------------------

func LoadDefault(cfg *HttpServerConfig) { Default.Load("default", cfg) }
