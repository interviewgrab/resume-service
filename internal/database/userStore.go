package database

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"resume-service/internal/model"
)

type UserStore struct {
	collection *mongo.Collection
}

const userCollection = "users"

func newUserStore(dbClient *mongo.Database) (UserStore, error) {
	return UserStore{collection: dbClient.Collection(userCollection)}, nil
}

func (s *UserStore) GetUserByEmail(ctx context.Context, email string) (model.User, error) {
	user := &model.User{}
	filter := bson.M{"email": email}
	err := s.collection.FindOne(ctx, filter).Decode(user)

	if err != nil {
		if !IsNotFound(err) {
			log.Println("Error finding user", err)
		}
		return *user, err
	}

	return *user, nil
}

func (s *UserStore) CreateUser(ctx context.Context, user model.User) (model.User, error) {
	res, err := s.collection.InsertOne(ctx, user)
	if err != nil {
		return model.User{}, err
	}
	user.ID = res.InsertedID.(primitive.ObjectID)
	return user, nil
}

func IsNotFound(err error) bool {
	return err == mongo.ErrNoDocuments
}
