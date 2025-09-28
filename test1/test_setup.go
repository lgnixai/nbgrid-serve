package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"teable-go-backend/internal/infrastructure/database/models"
)

// TestDBConfig 测试数据库配置
type TestDBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// GetTestDBConfig 获取测试数据库配置
func GetTestDBConfig() *TestDBConfig {
	return &TestDBConfig{
		Host:     getEnv("TEST_DB_HOST", "localhost"),
		Port:     5432,
		User:     getEnv("TEST_DB_USER", "postgres"),
		Password: getEnv("TEST_DB_PASSWORD", "password"),
		DBName:   getEnv("TEST_DB_NAME", "teable_test"),
		SSLMode:  getEnv("TEST_DB_SSL_MODE", "disable"),
	}
}

// SetupTestDB 设置测试数据库
func SetupTestDB() (*gorm.DB, error) {
	config := GetTestDBConfig()

	// 构建DSN
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode)

	// 连接数据库
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // 测试时静默日志
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to test database: %w", err)
	}

	// 自动迁移数据库结构
	err = db.AutoMigrate(
		&models.User{},
		&models.Account{},
		&models.Space{},
		&models.Base{},
		&models.Table{},
		&models.Field{},
		&models.Record{},
		&models.View{},
		&models.Permission{},
		&models.ShareView{},
		&models.Attachment{},
		&models.Notification{},
		&models.NotificationTemplate{},
		&models.NotificationSubscription{},
		&models.SearchIndex{},
		&models.SearchSuggestion{},
		&models.SearchStats{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate test database: %w", err)
	}

	return db, nil
}

// CleanupTestDB 清理测试数据库
func CleanupTestDB(db *gorm.DB) error {
	// 删除所有测试数据
	tables := []interface{}{
		&models.SearchSuggestion{},
		&models.SearchStats{},
		&models.SearchIndex{},
		&models.NotificationSubscription{},
		&models.NotificationTemplate{},
		&models.Notification{},
		&models.Attachment{},
		&models.ShareView{},
		&models.Permission{},
		&models.View{},
		&models.Record{},
		&models.Field{},
		&models.Table{},
		&models.Base{},
		&models.Space{},
		&models.Account{},
		&models.User{},
	}

	for _, table := range tables {
		if err := db.Unscoped().Where("1 = 1").Delete(table).Error; err != nil {
			log.Printf("Warning: failed to cleanup table %T: %v", table, err)
		}
	}

	return nil
}

// CreateTestDB 创建测试数据库
func CreateTestDB() error {
	config := GetTestDBConfig()

	// 连接到postgres数据库来创建测试数据库
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=postgres sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}
	defer db.Close()

	// 检查数据库是否存在
	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = $1)", config.DBName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check database existence: %w", err)
	}

	if !exists {
		// 创建测试数据库
		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", config.DBName))
		if err != nil {
			return fmt.Errorf("failed to create test database: %w", err)
		}
		log.Printf("Created test database: %s", config.DBName)
	} else {
		log.Printf("Test database already exists: %s", config.DBName)
	}

	return nil
}

// DropTestDB 删除测试数据库
func DropTestDB() error {
	config := GetTestDBConfig()

	// 连接到postgres数据库来删除测试数据库
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=postgres sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}
	defer db.Close()

	// 删除测试数据库
	_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", config.DBName))
	if err != nil {
		return fmt.Errorf("failed to drop test database: %w", err)
	}

	log.Printf("Dropped test database: %s", config.DBName)
	return nil
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
