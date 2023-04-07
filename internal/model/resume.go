package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Resume struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID     primitive.ObjectID `bson:"user_id,required" json:"user_id"`
	FileName   string             `bson:"file_name,required" json:"file_name"`
	Key        string             `bson:"key,required" json:"key"`
	UploadDate time.Time          `bson:"upload_date,required" json:"upload_date"`
	Metadata   map[string]string  `bson:"metadata,omitempty" json:"metadata"`
	Public     bool               `bson:"public,required" json:"public"`
}
