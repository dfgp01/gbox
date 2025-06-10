package main

import (
	"gbox/logger"
	"time"

	"go.uber.org/zap/zapcore"
)

var lg *logger.ZapLogger

func encoderDemo() zapcore.Encoder {
	// 一個encoder示例
	return zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
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

func useRotateLog() {
	cfg := &logger.LogConfig{
		Level: logger.Info,
		Rot: &logger.RotateLogConfig{
			Link:      "./logs/app.log",
			Formatter: "./logs/app.%Y%m%d.log",
		},
	}
	lg = logger.NewZapLogger(
		logger.UseRotateWriter(cfg.Rot),
		logger.UseSugar(),
		logger.WithEncoder(encoderDemo()))
}

func useLumberjack() {
	cfg := &logger.LogConfig{
		Level: logger.Info,
		Lum: &logger.LumberjackConfig{
			Filename: "testlumber.log",
			Compress: true,
		},
	}
	lg = logger.NewZapLogger(
		logger.UseLumberjackWriter(cfg.Lum),
		logger.UseSugar(),
		logger.WithEncoder(encoderDemo()))
}

func main() {
	//useLumberjack()
	useRotateLog()
	lg.Info("this is info msg", "key1", 123, "key2", "abc")
	lg.Infof("this is infof msg key1=%v, key2=%v", 123, "abc")
}
