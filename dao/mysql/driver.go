package mysql

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	Database string
}

type GormDriver struct {
	db *gorm.DB
}

func Dsn(config *Config) string {
	//check config
	if config == nil || config.Host == "" || config.Database == "" {
		return ""
	}
	if config.Port == "" {
		config.Port = "3306"
	}
	if config.Username == "" {
		config.Username = "root"
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", config.Username, config.Password, config.Host, config.Port, config.Database)
}

func NewGormDriver(config *Config) (*GormDriver, error) {

	// use gorm
	db, err := gorm.Open(mysql.Open(Dsn(config)), &gorm.Config{
		SkipDefaultTransaction: true,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold: 300 * time.Millisecond, // 慢查询阈值
				LogLevel:      logger.Info,            // 日志级别
			},
		),
	})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	return &GormDriver{db: db}, nil
}
