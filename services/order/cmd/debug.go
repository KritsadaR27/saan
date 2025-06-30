package main

import (
	"fmt"
	"os"

	"github.com/saan/order-service/internal/infrastructure/config"
)

func main() {
	fmt.Println("Debug: Loading configuration...")
	
	// Print environment variables
	fmt.Printf("DB_HOST: %s\n", os.Getenv("DB_HOST"))
	fmt.Printf("DB_PORT: %s\n", os.Getenv("DB_PORT"))
	fmt.Printf("DB_USER: %s\n", os.Getenv("DB_USER"))
	fmt.Printf("DB_PASSWORD: %s\n", os.Getenv("DB_PASSWORD"))
	fmt.Printf("DB_NAME: %s\n", os.Getenv("DB_NAME"))
	fmt.Printf("DB_SSLMODE: %s\n", os.Getenv("DB_SSLMODE"))
	
	// Load configuration
	cfg := config.LoadConfig()
	
	fmt.Printf("Config loaded:\n")
	fmt.Printf("  Host: %s\n", cfg.Database.Host)
	fmt.Printf("  Port: %d\n", cfg.Database.Port)
	fmt.Printf("  User: %s\n", cfg.Database.User)
	fmt.Printf("  Password: %s\n", cfg.Database.Password)
	fmt.Printf("  DBName: %s\n", cfg.Database.DBName)
	fmt.Printf("  SSLMode: %s\n", cfg.Database.SSLMode)
	
	// Construct DSN like the db package does
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.DBName, cfg.Database.SSLMode,
	)
	fmt.Printf("DSN: %s\n", dsn)
}
