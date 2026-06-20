package common

import (
	"context"
	"errors"
	"sync"
	"time"
)

// IMessage 消息报文的基础构成。
type IMessage interface {
	Len() int32
	Payload() []byte
}

// IService 定义了公共的服务生命周期流程：Start -> Listen -> Stop。
type IService interface {
	// Start 执行初始化（注册路由、准备资源等），在 Listen 之前调用。
	Start(ctx context.Context) error
	// Listen 启动服务并阻塞监听，ctx 取消时触发优雅关闭。
	Listen(ctx context.Context, port int) error
	// Stop 主动关闭服务。
	Stop(ctx context.Context) error
}

var ErrTopicEmpty = errors.New("topic cannot be empty")

// Msg is the standard structured message used throughout the service layer.
type Msg struct {
	MsgID     string            `json:"msg_id"`
	TraceID   string            `json:"trace_id"`
	SpanID    string            `json:"span_id"`
	Timestamp int64             `json:"timestamp"` // unix nano
	Topic     string            `json:"topic"`
	Retry     int32             `json:"retry"`
	Headers   map[string]string `json:"headers"`
	Payload   []byte            `json:"payload"`
}

// NewMsg creates a new Msg with the current timestamp and the given parameters.
func NewMsg(msgID, traceID, topic string, payload []byte) Msg {
	return Msg{
		MsgID:     msgID,
		TraceID:   traceID,
		Timestamp: time.Now().UnixNano(),
		Topic:     topic,
		Payload:   payload,
		Headers:   make(map[string]string),
	}
}

// MsgHandler is a callback function for processing received messages.
type MsgHandler func(msg Msg)

// IMsgListener defines a service interface for subscribing to messages.
// Subscribe is the only method — publishing is handled by a unified node.
type IMsgListener interface {
	// Subscribe registers a handler for the specified topic.
	// The subscription lifecycle is bound to ctx; when ctx is cancelled the
	// goroutine is stopped and cleanup is performed.
	Subscribe(ctx context.Context, topic string, handler MsgHandler) error
}

// MsgListener is an in-memory implementation of IMsgListener.
// It is intended for prototyping and testing the publish-subscribe pattern.
type MsgListener struct {
	mu       sync.RWMutex
	channels map[string][]chan Msg
}

func NewMsgListener() *MsgListener {
	return &MsgListener{
		channels: make(map[string][]chan Msg),
	}
}

// Publish delivers a message to all subscribers of the given topic.
// This method is NOT part of IMsgListener — publishing is the responsibility
// of a unified node. It is exposed here so the in-memory implementation can
// be used for validation.
func (l *MsgListener) Publish(topic string, msg Msg) error {
	if topic == "" {
		return ErrTopicEmpty
	}
	l.mu.RLock()
	subs, ok := l.channels[topic]
	l.mu.RUnlock()
	if !ok {
		return nil
	}
	for _, ch := range subs {
		select {
		case ch <- msg:
		default:
		}
	}
	return nil
}

// Subscribe registers a handler for the given topic. Messages are delivered
// to the handler in a background goroutine. When ctx is cancelled the goroutine
// exits and the subscriber is cleaned up.
func (l *MsgListener) Subscribe(ctx context.Context, topic string, handler MsgHandler) error {
	if topic == "" {
		return ErrTopicEmpty
	}
	ch := make(chan Msg, 64)

	l.mu.Lock()
	l.channels[topic] = append(l.channels[topic], ch)
	l.mu.Unlock()

	go func() {
		defer func() {
			l.mu.Lock()
			subs := l.channels[topic]
			for i, c := range subs {
				if c == ch {
					l.channels[topic] = append(subs[:i], subs[i+1:]...)
					break
				}
			}
			if len(l.channels[topic]) == 0 {
				delete(l.channels, topic)
			}
			l.mu.Unlock()
			close(ch)
		}()

		for {
			select {
			case <-ctx.Done():
				for msg := range ch {
					handler(msg)
				}
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				handler(msg)
			}
		}
	}()

	return nil
}
