package main

import (
	"context"
	"fmt"
	"gbox/core"
	"time"
)

func main() {
	eventManager := core.GetEventManager()

	//中間件
	eventManager.Use(
		func(ctx context.Context, evt *core.Event, next func(ctx context.Context, evt *core.Event) error) error {
			start := time.Now()
			err := next(ctx, evt)
			fmt.Printf("cost: %d ms\n", time.Since(start).Milliseconds())
			return err
		},
		func(ctx context.Context, evt *core.Event, next func(ctx context.Context, evt *core.Event) error) error {
			defer func() {
				if r := recover(); r != nil {
					fmt.Println("Recovered from panic:", r)
				}
			}()
			return next(ctx, evt)
		},
	)

	handle1 := func(event *core.Event) bool {
		fmt.Printf("type:%v High priority click handler, data:%v\n", event.Type(), event.Data)
		return event.Data != "stop" // 返回 false 表示停止冒泡
	}
	handle2 := func(event *core.Event) bool {
		fmt.Printf("type:%v Low priority click handler, data:%v\n", event.Type(), event.Data)
		return true
	}

	// 注册监听器
	eventManager.RegisterListener("click", handle1, 10)
	eventManager.RegisterListener("click", handle2, 1)
	eventManager.RegisterListener("keydown", handle1, 1)
	eventManager.RegisterListener("keydown", handle2, 10)

	// 创建一个上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动事件管理器
	eventManager.Start(ctx)

	// 发送事件
	eventManager.SendEvent("keydown", "go", ctx)
	eventManager.SendEvent("click", "stop", ctx)

	// 模拟运行一段时间后停止
	<-time.After(2 * time.Second)
	cancel() // 发送停止信号

}
