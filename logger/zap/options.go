package zp

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

//zap的組成：
//	core,		可用NewTee()來包裝多個core
//		encoder,		*根據自己需求制定，提供幾個示例選擇
//		writesyncer,	*封裝writer，比如lumber，可能要抽到外層，其他logger也要用
//		levelEnable		*AtomicLevel	如果需要動態改變日志level，只能用這個
//	option	*wrapcore		可能是很適合框架封裝的

type option func(z *ZapLogger)

func WithZapCore(cores ...zapcore.Core) option {
	return func(z *ZapLogger) {
		z.cores = append(z.cores, cores...)
	}
}

func WithZapOption(zops ...zap.Option) option {
	return func(z *ZapLogger) {
		z.zops = append(z.zops, zops...)
	}
}

func AddCaller(skip ...int) option {
	ops := []zap.Option{zap.AddCaller()}
	if len(skip) > 0 && skip[0] > 0 {
		ops = append(ops, zap.AddCallerSkip(skip[0]))
	}
	return WithZapOption(ops...)
}

func UseSugar() option {
	return func(z *ZapLogger) {
		z.useSugar = true
	}
}

func WithEncoder(enc zapcore.Encoder) option {
	return func(z *ZapLogger) {
		z.corePart.enc = enc
	}
}

func WithWriter(w zapcore.WriteSyncer) option {
	return func(z *ZapLogger) {
		z.corePart.writer = w
	}
}

func WithLevel(l zapcore.Level) option {
	return func(z *ZapLogger) {
		z.corePart.lv = l
	}
}

func UseAtomLevel() option {
	return WithLevel(zap.NewAtomicLevel().Level())
}
