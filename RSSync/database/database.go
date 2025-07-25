package database

import (
	"database/sql"
	"time"

	// Postgresql driver import
	_ "github.com/lib/pq"
)

// NewConnectionPool configures the database connection pool.
func NewConnectionPool(dsn string, minConnections, maxConnections int, connectionLifetime time.Duration) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(maxConnections)
	db.SetMaxIdleConns(minConnections)
	db.SetConnMaxLifetime(connectionLifetime)

	return db, nil
}
