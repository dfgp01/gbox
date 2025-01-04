package zp

import (
	"fmt"
	"time"

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

type ZapLogger struct {
	ops      []option
	zops     []zap.Option
	cores    []zapcore.Core
	lg       *zap.Logger
	su       *zap.SugaredLogger
	useSugar bool
}

func (z *ZapLogger) build() {
	for _, op := range z.ops {
		op(z)
	}

	if len(z.cores) == 0 {
		z.lg = zap.NewNop()
		return
	}
	if len(z.cores) == 1 {
		z.lg = zap.New(z.cores[0])
	} else {
		z.lg = zap.New(zapcore.NewTee(z.cores...))
	}
	z.lg.WithOptions(z.zops...)
	z.su = z.lg.Sugar()
}

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

func NewZapLogger(ops ...option) *ZapLogger {
	z := &ZapLogger{ops: ops}
	z.build()
	return z
}

func encoderDemo() {
	// 一個encoder示例
	zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		MessageKey:  "msg",
		LevelKey:    "level",
		EncodeLevel: zapcore.CapitalLevelEncoder,
		TimeKey:     "ts",
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05"))
		},
		CallerKey:    "file",
		EncodeCaller: zapcore.ShortCallerEncoder,
		EncodeDuration: func(d time.Duration, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendInt64(int64(d) / 1000000)
		},
	})
}

func fields(kv ...interface{}) (fs []zap.Field) {
	if len(kv) > 0 && len(kv)%2 == 0 {
		for i := 0; i < len(kv); i += 2 {
			k, ok := kv[i].(string)
			if !ok {
				return
			}
			v := kv[i+1]
			fs = append(fs, zap.Any(k, v))
		}
	}
	return
}

func (z *ZapLogger) Info(msg string, kv ...interface{}) {
	if z.useSugar {
		z.su.Infow(msg, kv...)
	} else {
		z.lg.Info(msg, fields(kv...)...)
	}
}

func (z *ZapLogger) Infof(template string, args ...interface{}) {
	if z.useSugar {
		z.su.Infof(template, args...)
	} else {
		z.lg.Info(fmt.Sprintf(template, args...))
	}
}
