package mysql

// MySQLConfig MySQL配置结构体
type MySQLConfig struct {
	Master   *DBConfig   `json:"master"`   // 主库配置
	Slaves   []*DBConfig `json:"slaves"`   // 从库配置列表
	Database string      `json:"database"` // 数据库名（放这里可以保证主从对应同一个库）
}

// DBConfig 数据库连接配置
type DBConfig struct {
	Host     string `json:"host"`     // 主机地址
	Port     int    `json:"port"`     // 端口
	Username string `json:"username"` // 用户名
	Password string `json:"password"` // 密码
	Charset  string `json:"charset"`  // 字符集
}

// GormConfig GORM配置结构体
type GormConfig struct {
	MaxIdleConns    int  `json:"max_idle_conns"`    // 最大空闲连接数
	MaxOpenConns    int  `json:"max_open_conns"`    // 最大打开连接数
	ConnMaxLifetime int  `json:"conn_max_lifetime"` // 连接最大生命周期（秒）
	LogMode         bool `json:"log_mode"`          // 是否开启日志模式
}
