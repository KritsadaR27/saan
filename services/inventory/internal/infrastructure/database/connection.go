package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"inventory/internal/config"
)

// Connection wraps database connection with enhanced functionality
type Connection struct {
	DB     *sql.DB
	logger *logrus.Logger
}

// NewConnection creates a new database connection with configuration
func NewConnection(cfg config.DatabaseConfig, logger *logrus.Logger) (*Connection, error) {
	var dsn string
	if cfg.URL != "" {
		// Use URL if provided (for backward compatibility)
		dsn = cfg.URL
	} else {
		// Build DSN from components
		dsn = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name, cfg.SSLMode)
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Configure connection pool
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)

	return &Connection{
		DB:     db,
		logger: logger,
	}, nil
}

// NewConnectionSimple creates a new database connection with URL (for backward compatibility)
func NewConnectionSimple(databaseURL string) (*Connection, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Connection{
		DB:     db,
		logger: logrus.New(),
	}, nil
}

// Close closes the database connection
func (c *Connection) Close() error {
	return c.DB.Close()
}

// Health checks the database connection health
func (c *Connection) Health() error {
	return c.DB.Ping()
}

// GetDB returns the underlying database connection
func (c *Connection) GetDB() *sql.DB {
	return c.DB
}
