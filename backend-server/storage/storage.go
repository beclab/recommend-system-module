package storage

import (
	"database/sql"
)

// Storage handles all operations related to the database.
type Storage struct {
	//mongodb *mongo.Client
	db *sql.DB
}

// NewStorage returns a new Storage.
func NewStorage(db *sql.DB) *Storage {
	return &Storage{db}
}
