package storage

import (
	"go.mongodb.org/mongo-driver/mongo"
)

// Storage handles all operations related to the database.
type Storage struct {
	mongodb *mongo.Client
}

// NewStorage returns a new Storage.
func NewStorage(mongodb *mongo.Client) *Storage {
	return &Storage{mongodb}
}
