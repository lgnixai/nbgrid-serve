package main

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"teable-go-backend/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		return
	}

	dsn := cfg.Database.GetDSN()
	fmt.Printf("Connecting with DSN: %s\n", dsn)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		return
	}

	sqlDB, err := db.DB()
	if err != nil {
		fmt.Printf("Failed to get sql.DB: %v\n", err)
		return
	}

	if err := sqlDB.Ping(); err != nil {
		fmt.Printf("Failed to ping: %v\n", err)
		return
	}

	fmt.Println("Database connection successful!")
}
