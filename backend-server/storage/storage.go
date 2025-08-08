package storage

import (
	"database/sql"

	"github.com/go-redis/redis"
)

// Storage handles all operations related to the database.
type Storage struct {
	db      *sql.DB
	redisdb *redis.Client
}

// NewStorage returns a new Storage.
func NewStorage(db *sql.DB, redisdb *redis.Client) *Storage {
	return &Storage{db, redisdb}
}
