package core

import (
	"context"
)

//異步運行的組件，Worker

type WorkerComponent interface {
	Component
	Run(context.Context)
}

//	崩潰處理機制

type Handler func(ctx context.Context, r interface{})

func WorkerRun(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			if hasRecHandler() {
				rHandler(r)
			}
		}
	}()
	worker.Action(ctx)
}

//	worker 同步發送事件

func WorkerAction(ctx context.Context) {
	for {
		select {
		case raw := <-redis.brPop():
			//異步
			sendEvent(&RedisRecvEvt{key: "key", data: raw})
			//同步
			worker.handler(&RedisRecvEvt{key: "key", data: raw})
		case ch := <-ctx.Done():
			return
		}
	}
}

func sendEvent(evt Event) {
	go func() {
		evtChan <- evt
	}()
}
