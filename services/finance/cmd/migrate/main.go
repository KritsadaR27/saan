package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Build database connection string
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "saan")
	dbPassword := getEnv("DB_PASSWORD", "saan_password")
	dbName := getEnv("DB_NAME", "saan_db")
	sslMode := getEnv("DB_SSLMODE", "disable")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, sslMode)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	// Run migrations
	if len(os.Args) > 1 && os.Args[1] == "down" {
		log.Println("Running down migrations...")
		if err := runDownMigrations(db); err != nil {
			log.Fatal("Migration failed:", err)
		}
	} else {
		log.Println("Running up migrations...")
		if err := runUpMigrations(db); err != nil {
			log.Fatal("Migration failed:", err)
		}
	}

	log.Println("✅ Finance Service migrations completed successfully!")
}

func runUpMigrations(db *sql.DB) error {
	migrations, err := getMigrationFiles("up")
	if err != nil {
		return err
	}

	for _, migration := range migrations {
		log.Printf("Running migration: %s", migration)
		content, err := ioutil.ReadFile(migration)
		if err != nil {
			return err
		}

		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("failed to run migration %s: %w", migration, err)
		}
	}

	return nil
}

func runDownMigrations(db *sql.DB) error {
	migrations, err := getMigrationFiles("down")
	if err != nil {
		return err
	}

	// Reverse order for down migrations
	for i := len(migrations) - 1; i >= 0; i-- {
		migration := migrations[i]
		log.Printf("Running down migration: %s", migration)
		content, err := ioutil.ReadFile(migration)
		if err != nil {
			return err
		}

		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("failed to run down migration %s: %w", migration, err)
		}
	}

	return nil
}

func getMigrationFiles(direction string) ([]string, error) {
	migrationsDir := "migrations"
	files, err := ioutil.ReadDir(migrationsDir)
	if err != nil {
		return nil, err
	}

	var migrations []string
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".sql" && 
		   filepath.Ext(filepath.Base(file.Name()[:len(file.Name())-4])) == "."+direction {
			migrations = append(migrations, filepath.Join(migrationsDir, file.Name()))
		}
	}

	sort.Strings(migrations)
	return migrations, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
