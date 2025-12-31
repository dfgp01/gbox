package mysql

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

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
	// 基础配置
	SkipDefaultTransaction                   bool `json:"skip_default_transaction"`                      // 禁用默认事务，默认值：false
	PrepareStmt                              bool `json:"prepare_stmt"`                                  // 启用预编译语句，默认值：false
	DisableNestedTransaction                 bool `json:"disable_nested_transaction"`                    // 禁用嵌套事务，默认值：false
	AllowGlobalUpdate                        bool `json:"allow_global_update"`                           // 允许全局更新，默认值：false
	DisableAutomaticPing                     bool `json:"disable_automatic_ping"`                        // 禁用自动连接检测，默认值：false
	DisableForeignKeyConstraintWhenMigrating bool `json:"disable_foreign_key_constraint_when_migrating"` // 迁移时禁用外键约束，默认值：false
	FullSaveAssociations                     bool `json:"full_save_associations"`                        // 保存所有关联数据，默认值：false
	QueryFields                              bool `json:"query_fields"`                                  // 查询时包含所有字段，默认值：false
	CreateBatchSize                          int  `json:"create_batch_size"`                             // 批量创建大小，默认值：0（不限制）

	// 连接池配置
	MaxIdleConns    int `json:"max_idle_conns"`    // 最大空闲连接数，默认值：2
	MaxOpenConns    int `json:"max_open_conns"`    // 最大打开连接数，默认值：0（不限制）
	ConnMaxLifetime int `json:"conn_max_lifetime"` // 连接最大生命周期（秒），默认值：0（不限制）

	// 日志配置
	LogMode       bool            `json:"log_mode"`       // 是否开启日志模式，默认值：false
	SlowThreshold time.Duration   `json:"slow_threshold"` // 慢查询阈值，默认值：200ms
	LogLevel      logger.LogLevel `json:"log_level"`      // 日志级别，默认值：Silent

	// 命名策略
	TablePrefix   string `json:"table_prefix"`   // 表前缀，默认值：""
	SingularTable bool   `json:"singular_table"` // 是否使用单数表名，默认值：false
	NameReplacer  string `json:"name_replacer"`  // 名称替换器，默认值：""
}

// ToGormConfig 转换为GORM配置
func (c *GormConfig) ToGormConfig() *gorm.Config {
	config := &gorm.Config{
		SkipDefaultTransaction:                   c.SkipDefaultTransaction,
		PrepareStmt:                              c.PrepareStmt,
		DisableNestedTransaction:                 c.DisableNestedTransaction,
		AllowGlobalUpdate:                        c.AllowGlobalUpdate,
		DisableAutomaticPing:                     c.DisableAutomaticPing,
		DisableForeignKeyConstraintWhenMigrating: c.DisableForeignKeyConstraintWhenMigrating,
		FullSaveAssociations:                     c.FullSaveAssociations,
		QueryFields:                              c.QueryFields,
		CreateBatchSize:                          c.CreateBatchSize,
	}

	// 配置命名策略
	if c.TablePrefix != "" || c.SingularTable || c.NameReplacer != "" {
		config.NamingStrategy = schema.NamingStrategy{
			TablePrefix:   c.TablePrefix,
			SingularTable: c.SingularTable,
		}
	}

	// 配置日志
	if c.LogMode {
		config.Logger = logger.Default.LogMode(c.LogLevel)
		if c.SlowThreshold > 0 {
			config.Logger = logger.Default.LogMode(c.LogLevel)
		}
	}

	return config
}
