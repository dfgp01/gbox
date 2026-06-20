package main

import (
	"context"
	"fmt"
	"gbox/common"
	"gbox/pkg/server"
	"time"
)

func main() {
	// 1. 加载配置（pkg 层，无 gin 依赖）
	server.LoadDefault(&server.HttpServerConfig{
		Port: 8080,
		Mode: common.ModeDebug,
	})

	// 2. 通过便捷路由链注册
	server.Default.Router("default",
		server.GET("/ping", func(ctx server.Context) {
			ctx.JSON(200, map[string]interface{}{
				"msg":  "pong",
				"time": time.Now().Unix(),
			})
		}),

		server.POST("/msg", func(ctx server.Context) {
			var msg common.Msg
			if err := ctx.ShouldBindJSON(&msg); err != nil {
				ctx.JSON(400, map[string]string{"error": err.Error()})
				return
			}
			msg.Timestamp = time.Now().UnixNano()
			fmt.Printf("recv msg: %+v\n", msg)
			ctx.JSON(200, msg)
		}),
	)

	// 3. 直接路由注册
	server.Default.GET("default", "/hello", func(ctx server.Context) {
		name := ctx.GetHeader("X-Name")
		if name == "" {
			name = "world"
		}
		ctx.String(200, fmt.Sprintf("Hello, %s!", name))
	})

	// 4. 启动
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := server.Default.Start(ctx, "default"); err != nil {
		panic(err)
	}
	fmt.Println("server listening on :8080")

	if err := server.Default.Listen(ctx, "default"); err != nil {
		panic(err)
	}
}
