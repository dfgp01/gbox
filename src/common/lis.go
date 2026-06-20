package common

import (
	"bufio"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// ---------------------------------------------------------------------------
// TcpServer — 简单的 TCP 服务，消息用长度前缀 + JSON 编码的 Msg
// ---------------------------------------------------------------------------

// TcpServer is a simple TCP server. Each message is length-prefixed JSON.
type TcpServer struct {
	hdl      MsgHandler
	listener net.Listener
	conns    map[net.Conn]struct{}
	mu       sync.Mutex
	cancel   context.CancelFunc
	done     chan struct{}
}

func NewTcpServer() *TcpServer {
	return &TcpServer{conns: make(map[net.Conn]struct{})}
}

// HandleMsg sets the handler for received messages.
func (s *TcpServer) HandleMsg(hdl MsgHandler) *TcpServer {
	s.hdl = hdl
	return s
}

func (s *TcpServer) Start(ctx context.Context) error { return nil }

func (s *TcpServer) Listen(ctx context.Context, port int) error {
	ctx, s.cancel = context.WithCancel(ctx)
	s.done = make(chan struct{})
	defer close(s.done)

	addr := fmt.Sprintf(":%d", port)
	var err error
	s.listener, err = net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	// 只有这一处关闭 listener，由 ctx 取消触发
	go func() {
		<-ctx.Done()
		s.listener.Close()
	}()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			break
		}
		s.mu.Lock()
		s.conns[conn] = struct{}{}
		s.mu.Unlock()

		go s.handleConn(conn, ctx)
	}

	// 等待所有连接处理完毕
	s.mu.Lock()
	for conn := range s.conns {
		conn.Close()
	}
	s.mu.Unlock()
	return nil
}

func (s *TcpServer) handleConn(conn net.Conn, ctx context.Context) {
	defer func() {
		conn.Close()
		s.mu.Lock()
		delete(s.conns, conn)
		s.mu.Unlock()
	}()

	reader := bufio.NewReader(conn)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// read 4-byte length prefix
		lenBuf := make([]byte, 4)
		if _, err := io.ReadFull(reader, lenBuf); err != nil {
			return
		}
		msgLen := binary.BigEndian.Uint32(lenBuf)
		if msgLen == 0 {
			continue
		}

		// read message body
		data := make([]byte, msgLen)
		if _, err := io.ReadFull(reader, data); err != nil {
			return
		}

		var msg Msg
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}
		if s.hdl != nil {
			s.hdl(msg)
		}
	}
}

func (s *TcpServer) Stop(ctx context.Context) error {
	if s.cancel != nil {
		s.cancel() // 触发 Listen 内的 listener.Close()
	}
	select {
	case <-s.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// ---------------------------------------------------------------------------
// UdpServer — 简单的 UDP 服务，每个数据报是一个 JSON 编码的 Msg
// ---------------------------------------------------------------------------

// UdpServer is a simple UDP server. Each datagram is a JSON-encoded Msg.
type UdpServer struct {
	hdl    MsgHandler
	pc     net.PacketConn
	cancel context.CancelFunc
	done   chan struct{}
}

func NewUdpServer() *UdpServer {
	return &UdpServer{}
}

// HandleMsg sets the handler for received messages.
func (s *UdpServer) HandleMsg(hdl MsgHandler) *UdpServer {
	s.hdl = hdl
	return s
}

func (s *UdpServer) Start(ctx context.Context) error { return nil }

func (s *UdpServer) Listen(ctx context.Context, port int) error {
	ctx, s.cancel = context.WithCancel(ctx)
	s.done = make(chan struct{})
	defer close(s.done)

	addr := fmt.Sprintf(":%d", port)
	var err error
	s.pc, err = net.ListenPacket("udp", addr)
	if err != nil {
		return err
	}

	// 只有这一处关闭 pc，由 ctx 取消触发
	go func() {
		<-ctx.Done()
		s.pc.Close()
	}()

	buf := make([]byte, 65535)
	for {
		n, _, err := s.pc.ReadFrom(buf)
		if err != nil {
			break
		}
		data := make([]byte, n)
		copy(data, buf[:n])

		var msg Msg
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}
		if s.hdl != nil {
			s.hdl(msg)
		}
	}
	return nil
}

func (s *UdpServer) Stop(ctx context.Context) error {
	if s.cancel != nil {
		s.cancel() // 触发 Listen 内的 pc.Close()
	}
	select {
	case <-s.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// ---------------------------------------------------------------------------
// WsServer — WebSocket 服务 (gorilla/websocket)，消息用 JSON 编码的 Msg
// ---------------------------------------------------------------------------

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// WsServer is a WebSocket server. Messages are JSON-encoded Msg.
type WsServer struct {
	engine *gin.Engine
	srv    *http.Server
	hdl    MsgHandler
	cancel context.CancelFunc
	done   chan struct{}
}

func NewWsServer() *WsServer {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())
	return &WsServer{engine: engine}
}

// Engine returns the internal *gin.Engine for custom route registration.
func (w *WsServer) Engine() *gin.Engine {
	return w.engine
}

// HandleMsg sets the handler and registers the /ws upgrade endpoint.
func (w *WsServer) HandleMsg(hdl MsgHandler) *WsServer {
	w.hdl = hdl
	w.engine.GET("/ws", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				break
			}
			var msg Msg
			if err := json.Unmarshal(data, &msg); err != nil {
				continue
			}
			if w.hdl != nil {
				w.hdl(msg)
			}
		}
	})
	return w
}

func (w *WsServer) Start(ctx context.Context) error { return nil }

func (w *WsServer) Listen(ctx context.Context, port int) error {
	ctx, w.cancel = context.WithCancel(ctx)
	w.done = make(chan struct{})
	defer close(w.done)

	addr := fmt.Sprintf(":%d", port)
	w.srv = &http.Server{Addr: addr, Handler: w.engine}

	// 只有这一处调用 Shutdown，由 ctx 取消触发
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		w.srv.Shutdown(shutdownCtx)
	}()

	if err := w.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Stop 触发优雅关闭并等待 Listen 完全退出。ctx 控制等待超时。
func (w *WsServer) Stop(ctx context.Context) error {
	if w.cancel != nil {
		w.cancel() // 触发 Listen 内的 Shutdown
	}
	select {
	case <-w.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
