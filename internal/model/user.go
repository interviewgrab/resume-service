package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID       primitive.ObjectID `bson:"_id,required" json:"id"`
	Name     string             `bson:"name,required" json:"name"`
	Email    string             `bson:"email,required" json:"email"`
	Password string             `bson:"password,required" json:"password"`
}
