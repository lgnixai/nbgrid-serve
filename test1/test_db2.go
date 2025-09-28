package main

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 尝试不同的DSN格式
	dsns := []string{
		"host=localhost user=leven password= dbname=teable port=5432 sslmode=disable",
		"postgres://leven@localhost:5432/teable?sslmode=disable",
		"user=leven host=localhost port=5432 dbname=teable sslmode=disable",
	}

	for i, dsn := range dsns {
		fmt.Printf("Trying DSN %d: %s\n", i+1, dsn)
		
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			fmt.Printf("Failed to connect: %v\n", err)
		} else {
			fmt.Println("Connection successful!")
			
			sqlDB, err := db.DB()
			if err != nil {
				fmt.Printf("Failed to get sql.DB: %v\n", err)
				continue
			}

			if err := sqlDB.Ping(); err != nil {
				fmt.Printf("Failed to ping: %v\n", err)
				continue
			}

			fmt.Println("Ping successful!")
			sqlDB.Close()
			break
		}
		fmt.Println()
	}
}
