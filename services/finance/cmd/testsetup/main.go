package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	
	_ "github.com/lib/pq" // PostgreSQL driver
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run testsetup.go [setup|teardown]")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "setup":
		setupTestDatabase()
	case "teardown":
		teardownTestDatabase()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func setupTestDatabase() {
	fmt.Println("Setting up test database...")
	
	// Set test environment variables
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USER", "postgres")
	os.Setenv("DB_PASSWORD", "postgres")
	os.Setenv("DB_NAME", "finance_test")
	os.Setenv("DB_SSLMODE", "disable")
	
	// Connect to database
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSLMODE"),
	)
	
	database, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}
	defer database.Close()
	
	// Test the connection
	if err := database.Ping(); err != nil {
		log.Fatalf("Failed to ping test database: %v", err)
	}
	
	// Run migrations would go here
	fmt.Println("Test database setup complete!")
}

func teardownTestDatabase() {
	fmt.Println("Tearing down test database...")
	
	// Connect and clean up test data
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "postgres"),
		getEnv("DB_NAME", "finance_test"),
		getEnv("DB_SSLMODE", "disable"),
	)
	
	database, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("Failed to connect to test database for cleanup: %v", err)
		return
	}
	defer database.Close()
	
	// Drop test tables or truncate data
	tables := []string{
		"cash_flow_records",
		"expense_entries", 
		"cash_transfers",
		"profit_allocation_rules",
		"daily_cash_summaries",
	}
	
	for _, table := range tables {
		_, err := database.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			log.Printf("Warning: Failed to truncate table %s: %v", table, err)
		}
	}
	
	fmt.Println("Test database cleanup complete!")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
