package database

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"order/internal/infrastructure/config"
)

// Connection wraps sqlx.DB with additional functionality
type Connection struct {
	DB     *sqlx.DB
	logger *logrus.Logger
}

// NewConnection creates a new database connection
func NewConnection(cfg config.DatabaseConfig, logger *logrus.Logger) (*Connection, error) {
	var dsn string
	if cfg.URL != "" {
		dsn = cfg.URL
	} else {
		dsn = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode,
		)
	}

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connection established successfully")

	return &Connection{
		DB:     db,
		logger: logger,
	}, nil
}

// Connect establishes a connection to PostgreSQL database (backward compatibility)
func Connect(cfg *config.DatabaseConfig) (*sqlx.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode,
	)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// Close closes the database connection
func (c *Connection) Close() error {
	if c.DB != nil {
		c.logger.Info("Closing database connection")
		return c.DB.Close()
	}
	return nil
}

// Health checks the database connection health
func (c *Connection) Health() error {
	if c.DB == nil {
		return fmt.Errorf("database connection is nil")
	}
	return c.DB.Ping()
}

// GetDB returns the underlying sqlx.DB for backward compatibility
func (c *Connection) GetDB() *sqlx.DB {
	return c.DB
}
