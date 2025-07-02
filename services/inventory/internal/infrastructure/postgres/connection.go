package postgres

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type Connection struct {
	DB *sql.DB
}

func NewConnection(databaseURL string) *Connection {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to database: %v", err))
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		panic(fmt.Sprintf("Failed to ping database: %v", err))
	}

	return &Connection{
		DB: db,
	}
}

func (c *Connection) Close() error {
	return c.DB.Close()
}
