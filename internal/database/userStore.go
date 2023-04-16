package database

import (
	"context"
	"log"
	"resume-service/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserStore struct {
	collection *mongo.Collection
}

const userCollection = "users"

func newUserStore(ctx context.Context, dbClient *mongo.Database) (UserStore, error) {
	collection := dbClient.Collection(userCollection)
	err := createUserIndexes(ctx, collection)
	if err != nil {
		return UserStore{}, err
	}
	return UserStore{collection: collection}, nil
}

func createUserIndexes(ctx context.Context, collection *mongo.Collection) error {
	mod := mongo.IndexModel{
		Keys:    bson.M{"email": 1},
		Options: options.Index().SetUnique(true),
	}

	_, err := collection.Indexes().CreateOne(ctx, mod)
	return err
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
