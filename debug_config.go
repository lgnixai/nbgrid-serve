package tools

import (
	"fmt"
	"teable-go-backend/internal/config"
)

func PrintDebugConfig() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		return
	}

	fmt.Printf("Database config:\n")
	fmt.Printf("  Host: %s\n", cfg.Database.Host)
	fmt.Printf("  Port: %d\n", cfg.Database.Port)
	fmt.Printf("  User: %s\n", cfg.Database.User)
	fmt.Printf("  Password: %s\n", cfg.Database.Password)
	fmt.Printf("  Name: %s\n", cfg.Database.Name)
	fmt.Printf("  SSLMode: %s\n", cfg.Database.SSLMode)

	fmt.Printf("\nDSN: %s\n", cfg.Database.GetDSN())
}
