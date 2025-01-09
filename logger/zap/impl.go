package zp

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapLogger struct {
	ops      []option
	zops     []zap.Option
	cores    []zapcore.Core
	corePart struct {
		enc    zapcore.Encoder
		writer zapcore.WriteSyncer
		lv     zapcore.Level
	}
	lg       *zap.Logger
	su       *zap.SugaredLogger
	useSugar bool
}

func (z *ZapLogger) build() {
	for _, op := range z.ops {
		op(z)
	}

	if z.corePart.enc != nil && z.corePart.writer != nil {
		z.cores = append(z.cores, zapcore.NewCore(z.corePart.enc, z.corePart.writer, z.corePart.lv))
	}

	if len(z.cores) == 0 {
		z.lg = zap.NewNop()
		return
	}
	if len(z.cores) == 1 {
		z.lg = zap.New(z.cores[0], z.zops...)
	} else {
		z.lg = zap.New(zapcore.NewTee(z.cores...), z.zops...)
	}
	z.su = z.lg.Sugar()
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
