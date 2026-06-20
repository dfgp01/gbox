package common

// HTTP 模式常量
const (
	ModeRelease = "release"
	ModeDebug   = "debug"
	ModeTest    = "test"
)

// HttpContext 是 HTTP 请求上下文接口，由 internal 层实现，pkg 层使用。
// 定义在 common 中使得 pkg 和 internal 共享同一类型，避免桥接时的类型不匹配。
type HttpContext interface {
	JSON(code int, obj interface{})
	ShouldBindJSON(obj interface{}) error
	String(code int, str string)
	GetHeader(key string) string
}

// HttpServerConfig HTTP 服务配置项。
type HttpServerConfig struct {
	Port            int    `json:"port"`             // 监听端口，默认 8080
	Mode            string `json:"mode"`             // release / debug / test
	ReadTimeout     int    `json:"read_timeout"`     // 读超时（秒），0 表示不限制
	WriteTimeout    int    `json:"write_timeout"`    // 写超时（秒），0 表示不限制
	MaxHeaderBytes  int    `json:"max_header_bytes"` // 最大请求头字节数，0 使用默认值
	ShutdownTimeout int    `json:"shutdown_timeout"` // 优雅关闭等待时间（秒），默认 5
}

// FillDefaults 用合理的默认值填充空字段。
func (c *HttpServerConfig) FillDefaults() {
	if c.Port == 0 {
		c.Port = 8080
	}
	if c.Mode == "" {
		c.Mode = ModeRelease
	}
	if c.ShutdownTimeout == 0 {
		c.ShutdownTimeout = 5
	}
}
