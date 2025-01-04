package logger

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

type ILogger interface {
	Info(msg string, kv ...interface{})
	Infof(template string, args ...interface{})
}

type Config struct {
	Level Level `json:"level"`
	File  *File `json:"file"`
}

type File struct {
	Path       string `json:"path"`        // 日志文件路径，如：./logs/app.log
	Formatter  string `json:"formatter"`   // 日志切割格式，如：./logs/app.%Y%m%d%H%M%S.log
	MaxSize    int    `json:"max_size"`    // 日志文件大小限制，单位为MB
	MaxBackups int    `json:"max_backups"` // 保留旧文件的最大数量
	MaxAge     int    `json:"max_age"`     // 旧文件保留天数
	Compress   bool   `json:"compress"`    // 是否压缩旧文件
}

var Logger ILogger

// 暂时放这里
func UseLogrus(cfg *Config) ILogger {
	Logger = NewLogrus(cfg)
	return Logger
}

func UseZap(cfg *Config) ILogger {
	Logger = NewZap(cfg)
	return Logger
}
