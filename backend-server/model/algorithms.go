package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Algorithms struct {
	ID        primitive.ObjectID `bson:"_id"`
	Entry     string             `bson:"entry"`
	Source    string             `bson:"source"`
	CreatedAt time.Time          `bson:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt"`
}
