package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name          string             `bson:"name,required" json:"name"`
	Email         string             `bson:"email,required" json:"email"`
	Password      string             `bson:"password,required" json:"password"`
	EmailVerified bool               `bson:"email_verified,required" json:"email_verified"`
	EmailToken    string             `bson:"email_otp" json:"email_otp"`
}
