package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapLogger struct {
	ops      []option
	zops     []zap.Option
	cores    []zapcore.Core      //自定義的core
	enc      zapcore.Encoder     //指定的encoder，如果有cores，則忽略此項
	writer   zapcore.WriteSyncer //指定的writer，如果有cores，則忽略此項
	lv       zapcore.Level       //指定的level，如果有cores，則忽略此項
	lg       *zap.Logger         //如果自行指定此項，則忽略cores, enc, writer, lv
	su       *zap.SugaredLogger
	useSugar bool
}

func (z *ZapLogger) build() {
	defer func() {
		z.su = z.lg.Sugar()
	}()

	//已指定了logger
	if z.lg != nil {
		return
	}

	for _, op := range z.ops {
		op(z)
	}

	// 修正 level
	if z.lv < zap.DebugLevel {
		z.lv = zap.DebugLevel
	} else if z.lv > zap.FatalLevel {
		z.lv = zap.FatalLevel
	}

	//enc, writer 兩項都指定了才創建core
	if z.enc != nil && z.writer != nil {
		z.cores = append(z.cores, zapcore.NewCore(z.enc, z.writer, z.lv))
	}

	// 沒有任何指定core
	if len(z.cores) == 0 {
		//z.lg = zap.NewNop()
		z.lg, _ = zap.NewDevelopment()
		return
	}

	// 有一個或多個core
	if len(z.cores) == 1 {
		z.lg = zap.New(z.cores[0], z.zops...)
	} else {
		z.lg = zap.New(zapcore.NewTee(z.cores...), z.zops...)
	}
}

func NewZapLogger(ops ...option) *ZapLogger {
	z := &ZapLogger{ops: ops}
	z.build()
	return z
}
