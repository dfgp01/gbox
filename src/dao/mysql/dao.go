package mysql

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

// DAO 基础DAO结构体
type DAO struct {
	db *gorm.DB
}

// NewDAO 创建新的DAO实例
func NewDAO(mysqlConfig *MySQLConfig, gormConfig *GormConfig) (*DAO, error) {

	//check mysqlConfig
	if mysqlConfig == nil || mysqlConfig.Master == nil {
		return nil, fmt.Errorf("mysqlConfig or mysqlConfig.Master is nil")
	}
	if gormConfig == nil {
		return nil, fmt.Errorf("gormConfig is nil")
	}

	// 创建主库DSN
	masterDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		mysqlConfig.Master.Username,
		mysqlConfig.Master.Password,
		mysqlConfig.Master.Host,
		mysqlConfig.Master.Port,
		mysqlConfig.Database,
		mysqlConfig.Master.Charset,
	)

	// 创建从库DSN列表
	var slaveDSNs []string
	for _, slave := range mysqlConfig.Slaves {
		slaveDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
			slave.Username,
			slave.Password,
			slave.Host,
			slave.Port,
			mysqlConfig.Database,
			slave.Charset,
		)
		slaveDSNs = append(slaveDSNs, slaveDSN)
	}

	// 连接主库
	db, err := gorm.Open(mysql.Open(masterDSN), gormConfig.ToGormConfig())
	if err != nil {
		return nil, fmt.Errorf("connect master db error: %v", err)
	}

	// 配置主从
	if len(slaveDSNs) > 0 {
		err = db.Use(dbresolver.Register(dbresolver.Config{
			Sources: []gorm.Dialector{mysql.Open(masterDSN)},
			Replicas: func() []gorm.Dialector {
				var dialectors []gorm.Dialector
				for _, dsn := range slaveDSNs {
					dialectors = append(dialectors, mysql.Open(dsn))
				}
				return dialectors
			}(),
			Policy: dbresolver.RandomPolicy{}, // 使用随机策略
		}))
		if err != nil {
			return nil, fmt.Errorf("register dbresolver error: %v", err)
		}
	}

	// 设置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(gormConfig.MaxIdleConns)
	sqlDB.SetMaxOpenConns(gormConfig.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(gormConfig.ConnMaxLifetime) * time.Second)

	return &DAO{db: db}, nil
}

// DB 获取数据库连接
func (d *DAO) DB() *gorm.DB {
	return d.db
}

// Close 关闭数据库连接
func (d *DAO) Close() error {
	if d.db != nil {
		sqlDB, err := d.db.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
