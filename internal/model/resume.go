package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Resume struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	UserID     primitive.ObjectID `bson:"user_id"`
	FileName   string             `bson:"file_name"`
	Key        string             `bson:"key"`
	UploadDate time.Time          `bson:"upload_date"`
	Metadata   map[string]string  `bson:"metadata"`
	Public     bool               `bson:"public"`
}
