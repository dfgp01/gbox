package server

import (
	"context"
	"fmt"
	"gbox/common"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ---------------------------------------------------------------------------
// HttpOption — 函数式选项，用于 NewHttpServer 时注入自定义行为
// ---------------------------------------------------------------------------

// HttpOption 是 HttpServer 的函数式选项。
type HttpOption func(*HttpServer)

// WithReadTimeout 设置 http.Server 的读超时。
func WithReadTimeout(d time.Duration) HttpOption {
	return func(s *HttpServer) {
		if s.srv != nil {
			s.srv.ReadTimeout = d
		}
	}
}

// WithWriteTimeout 设置 http.Server 的写超时。
func WithWriteTimeout(d time.Duration) HttpOption {
	return func(s *HttpServer) {
		if s.srv != nil {
			s.srv.WriteTimeout = d
		}
	}
}

// ---------------------------------------------------------------------------
// HttpServer — 基于 gin 的 HTTP 服务，消息用 Msg 封装
// ---------------------------------------------------------------------------

// HttpServer is an HTTP server built on gin. Messages are represented as Msg.
type HttpServer struct {
	engine *gin.Engine
	srv    *http.Server
	cfg    *common.HttpServerConfig
	hdl    common.MsgHandler
	cancel context.CancelFunc
	done   chan struct{}
}

// NewHttpServer 创建 HttpServer 实例。cfg 可为 nil（使用默认配置）。
// 通过 opts 可传入额外选项，例如 WithReadTimeout / WithWriteTimeout。
func NewHttpServer(cfg *common.HttpServerConfig, opts ...HttpOption) *HttpServer {
	if cfg == nil {
		cfg = &common.HttpServerConfig{}
	}
	cfg.FillDefaults()

	gin.SetMode(cfg.Mode)
	engine := gin.New()
	engine.Use(gin.Recovery())

	s := &HttpServer{
		engine: engine,
		cfg:    cfg,
		srv: &http.Server{
			ReadTimeout:    time.Duration(cfg.ReadTimeout) * time.Second,
			WriteTimeout:   time.Duration(cfg.WriteTimeout) * time.Second,
			MaxHeaderBytes: cfg.MaxHeaderBytes,
		},
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Engine 返回内部 *gin.Engine，用于外部注册自定义路由。
func (s *HttpServer) Engine() *gin.Engine {
	return s.engine
}

// Config 返回当前配置副本。
func (s *HttpServer) Config() common.HttpServerConfig {
	return *s.cfg
}

// Use 注册 gin 中间件，是对 engine.Use() 的包装。
func (s *HttpServer) Use(middleware ...gin.HandlerFunc) *HttpServer {
	s.engine.Use(middleware...)
	return s
}

// UseHandler 注册中间件，handler 接收 common.HttpContext（包装了 gin.Context）。
// pkg 层通过此方法注册中间件，无需直接依赖 gin。
func (s *HttpServer) UseHandler(handler func(ctx common.HttpContext)) {
	s.engine.Use(func(gc *gin.Context) {
		handler(&ginContext{c: gc})
	})
}

// HandleMsg 注册 POST /msg 端点，请求体 JSON 映射为 Msg 并交给 handler 处理。
func (s *HttpServer) HandleMsg(hdl common.MsgHandler) *HttpServer {
	s.hdl = hdl
	s.engine.POST("/msg", func(c *gin.Context) {
		var msg common.Msg
		if err := c.ShouldBindJSON(&msg); err != nil {
			c.JSON(http.StatusBadRequest, common.NewMsg("", "", "error", []byte(err.Error())))
			return
		}
		msg.Timestamp = time.Now().UnixNano()
		if s.hdl != nil {
			s.hdl(msg)
		}
		c.JSON(http.StatusOK, msg)
	})
	return s
}

type ginContext struct {
	c *gin.Context
}

func (ctx *ginContext) JSON(code int, obj interface{})       { ctx.c.JSON(code, obj) }
func (ctx *ginContext) ShouldBindJSON(obj interface{}) error { return ctx.c.ShouldBindJSON(obj) }
func (ctx *ginContext) String(code int, str string)          { ctx.c.String(code, str) }
func (ctx *ginContext) GetHeader(key string) string          { return ctx.c.GetHeader(key) }

// Handle 注册一条路由。handler 接收 common.HttpContext（包装了 gin.Context）。
// pkg 层通过此方法注册路由，无需直接依赖 gin。
func (s *HttpServer) Handle(method, path string, handler func(ctx common.HttpContext)) {
	s.engine.Handle(method, path, func(gc *gin.Context) {
		handler(&ginContext{c: gc})
	})
}

func (s *HttpServer) Start(ctx context.Context) error { return nil }

func (s *HttpServer) Listen(ctx context.Context, port int) error {
	ctx, s.cancel = context.WithCancel(ctx)
	s.done = make(chan struct{})
	defer close(s.done)

	// Listen 参数 port 的优先级高于配置中的 port
	if port <= 0 {
		port = s.cfg.Port
	}
	addr := fmt.Sprintf(":%d", port)

	s.srv.Addr = addr
	s.srv.Handler = s.engine

	// 只有这一处调用 Shutdown，由 ctx 取消触发
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(),
			time.Duration(s.cfg.ShutdownTimeout)*time.Second)
		defer cancel()
		s.srv.Shutdown(shutdownCtx)
	}()

	if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Stop 触发优雅关闭并等待 Listen 完全退出。ctx 控制等待超时。
func (s *HttpServer) Stop(ctx context.Context) error {
	if s.cancel != nil {
		s.cancel()
	}
	select {
	case <-s.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
