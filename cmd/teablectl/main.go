package main

import (
	"fmt"
	"os"

	"teable-go-backend/internal/config"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	root := &cobra.Command{
		Use:   "teablectl",
		Short: "Teable backend utility CLI",
	}

	// generate-password command
	var password string
	genPass := &cobra.Command{
		Use:   "generate-password",
		Short: "Generate bcrypt hash for a password",
		RunE: func(cmd *cobra.Command, args []string) error {
			if password == "" {
				return fmt.Errorf("--password is required")
			}
			hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				return err
			}
			fmt.Printf("Password: %s\nHash: %s\n", password, string(hash))
			return nil
		},
	}
	genPass.Flags().StringVar(&password, "password", "", "password to hash")
	root.AddCommand(genPass)

	// debug-config command
	debugCfg := &cobra.Command{
		Use:   "debug-config",
		Short: "Print loaded configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			fmt.Printf("Database config:\n")
			fmt.Printf("  Host: %s\n", cfg.Database.Host)
			fmt.Printf("  Port: %d\n", cfg.Database.Port)
			fmt.Printf("  User: %s\n", cfg.Database.User)
			fmt.Printf("  Password: %s\n", cfg.Database.Password)
			fmt.Printf("  Name: %s\n", cfg.Database.Name)
			fmt.Printf("  SSLMode: %s\n", cfg.Database.SSLMode)
			fmt.Printf("\nDSN: %s\n", cfg.Database.GetDSN())
			return nil
		},
	}
	root.AddCommand(debugCfg)

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
