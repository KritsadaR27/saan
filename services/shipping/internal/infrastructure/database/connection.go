package database

import (
	"fmt"
	
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// NewConnection creates a new database connection
func NewConnection(databaseURL string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	
	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	
	return db, nil
}
