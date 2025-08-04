package database

import (
	"fmt"
	"log"
	"time"

	"chain/internal/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database 数据库管理器
type Database struct {
	db *gorm.DB
}

// New 创建新的数据库实例
func New(cfg *config.DatabaseConfig) (*Database, error) {
	// 构建 MySQL DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
	)

	// 配置 GORM 日志
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 获取底层的 sql.DB 对象进行连接池配置
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(10)                  // 最大空闲连接数
	sqlDB.SetMaxOpenConns(100)                 // 最大打开连接数
	sqlDB.SetConnMaxLifetime(time.Hour)        // 连接最大生存时间

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connected successfully")
	return &Database{db: db}, nil
}

// Close 关闭数据库连接
func (d *Database) Close() error {
	if d.db != nil {
		sqlDB, err := d.db.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// GetDB 获取数据库实例
func (d *Database) GetDB() *gorm.DB {
	return d.db
}

// AutoMigrate 自动迁移数据库表结构
func (d *Database) AutoMigrate(models ...interface{}) error {
	if d.db == nil {
		return fmt.Errorf("database not initialized")
	}
	return d.db.AutoMigrate(models...)
}