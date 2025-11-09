package bint

import (
	"database/sql"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type PostgresConfig struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	Username        string        `yaml:"username"`
	Password        string        `yaml:"password"`
	Database        string        `yaml:"database"`
	MaxIdleConns    int           `yaml:"maxIdleConns"`
	MaxOpenConns    int           `yaml:"maxOpenConns"`
	ConnMaxLifetime time.Duration `yaml:"connMaxLifetime"`
	ConnName        string        `yaml:"connName"`
	Debug           bool          `yaml:"debug"`
}

func initPostgres(cfg PostgresConfig, log logger.Interface) (db *gorm.DB, err error) {
	connName := cfg.ConnName
	if connName == "" {
		connName = cfg.Database
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=disable TimeZone=Asia/Shanghai application_name=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Database, cfg.Password, cfg.ConnName)
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "",
			SingularTable: true,
		},
		Logger: log,
	})
	if err != nil {
		return
	}

	if cfg.Debug {
		db = db.Debug()
	}

	// 设置连接池大小
	// 获取通用数据库对象 sql.DB ，然后使用其提供的功能
	var sqlDB *sql.DB
	sqlDB, err = db.DB()
	if err != nil {
		return
	}

	// SetMaxIdleConns 用于设置连接池中空闲连接的最大数量。
	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}

	// SetMaxOpenConns 设置打开数据库连接的最大数量。
	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}

	// SetConnMaxLifetime 设置了连接可复用的最大时间。
	if cfg.ConnMaxLifetime > time.Second {
		sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}

	return db, db.Select("select 1;").Error
}
