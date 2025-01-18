package logger

import (
	"fmt"
	"gbox/core"
	"io"
	"log"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"gopkg.in/natefinch/lumberjack.v2"
)

type (
	Level uint32
)

const (
	Panic Level = iota //意外崩潰錯誤
	Fatal              //程序終止錯誤
	Error              //一般錯誤
	Warn               //警告信息
	Info               //正常信息
	Debug              //調試信息
	Trace              //追蹤棧
)

type (
	Logger interface {
		Info(msg string, kv ...interface{})
		Infof(template string, args ...interface{})
	}

	LoggerComponent interface {
		core.Component
		Info(msg string, kv ...interface{})
		Infof(template string, args ...interface{})
	}
)

type LogConfig struct {
	Level Level             `json:"level"`
	Rot   *RotateLogConfig  `json:"rotate"`
	Lum   *LumberjackConfig `json:"lum"`
}

// Validate 验证配置的正确性
func (c *LogConfig) Validate() error {

	if c.Level < Panic || c.Level > Trace {
		return fmt.Errorf("配置错误: 无效的日志级别")
	}

	if c.Rot != nil {
		if c.Rot.Link == "" || c.Rot.Formatter == "" {
			return fmt.Errorf("配置错误: Link 和 Formatter 不能为空")
		}
		if c.Rot.KeepDays <= 0 {
			return fmt.Errorf("配置错误: KeepDays 必須大於0")
		}
	}

	if c.Lum != nil {
		if c.Lum.Filename == "" {
			return fmt.Errorf("配置错误: Filename 不能为空")
		}
	}

	return nil
}

// DefaultLogConfig 返回默认的日志配置
func DefaultLogConfig() *LogConfig {
	return &LogConfig{
		Level: Info,
		Rot: &RotateLogConfig{
			Link:      "./logs/app.log",
			Formatter: "./logs/app.%Y%m%d.log",
			KeepDays:  7,
			SplitFlag: "day",
		},
		Lum: &LumberjackConfig{
			Filename:   "./logs/app.log",
			MaxSize:    200,  // 200 MB
			MaxBackups: 3,    // 保留3个旧文件
			MaxAge:     7,    // 保留7天
			Compress:   true, // 压缩旧文件
		},
	}
}

// RotateLogConfig 日志旋转配置
type RotateLogConfig struct {
	Link      string `json:"link"`       // 日志文件软链，如：./logs/app.log
	Formatter string `json:"formatter"`  // 日志切割格式，如：./logs/app.%Y%m%d%H%M%S.log
	KeepDays  int    `json:"keep_days"`  // 日志文件保留天数
	SplitFlag string `json:"split_flag"` // 分割時間標志，缺省->""=每天，"hour"=每小時
}

// rotateLogWriter 根据配置生成日志旋转写入器
func rotateLogWriter(config *RotateLogConfig) io.Writer {

	// 示例配置
	// config := &RotateConfig{
	// 	Link:      "./logs/app.log",
	// 	Formatter: "./logs/app.%Y%m%d%H%M%S.log",
	// 	KeepDays:  7,
	//	SplitFlag: "day",
	// }

	// 解析日志文件路径和格式
	filePath := config.Formatter
	linkName := config.Link

	keepDays := config.KeepDays
	if keepDays <= 0 {
		keepDays = 1
	}
	dayTime := 24 * time.Hour
	maxAge := time.Duration(keepDays) * dayTime
	if config.SplitFlag == "hour" {
		dayTime = time.Hour
	}

	hook, err := rotatelogs.New(
		filePath, //logFile+"_%Y-%m-%d %H:%M:%S.log",
		rotatelogs.WithLinkName(linkName),
		rotatelogs.WithMaxAge(maxAge),
		rotatelogs.WithRotationTime(dayTime), // 这里可以根据需要调整旋转时间
	)
	if err != nil {
		panic(fmt.Sprintf("rotatelogs.New  failed: %v", err))
	}
	return hook
}

// LumberjackConfig 日志旋转配置
type LumberjackConfig struct {
	Filename   string `json:"filename"`    // 日志文件路径，如：./logs/app.log
	MaxSize    int    `json:"max_size"`    // 单个日志文件的最大大小（MB）
	MaxBackups int    `json:"max_backups"` // 保留的旧文件数量
	MaxAge     int    `json:"max_age"`     // 保留的旧文件的最大天数
	Compress   bool   `json:"compress"`    // 是否压缩旧文件
}

// lumberjackLogWriter 根据配置生成日志旋转写入器
func lumberjackLogWriter(config *LumberjackConfig) io.Writer {
	// 验证配置
	if config.Filename == "" {
		log.Fatalf("配置错误: Filename 不能为空")
	}
	if config.MaxSize <= 0 {
		config.MaxSize = 200 // 默认200MB
	}
	if config.MaxBackups <= 0 {
		config.MaxBackups = 3 // 默认保留3个旧文件
	}
	if config.MaxAge <= 0 {
		config.MaxAge = 7 // 默认保留7天
	}

	// 创建 lumberjack 日志旋转写入器
	lj := &lumberjack.Logger{
		Filename:   config.Filename,
		MaxSize:    config.MaxSize,
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge,
		Compress:   config.Compress,
	}

	//log.SetOutput(lj)
	//log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	return lj
}
