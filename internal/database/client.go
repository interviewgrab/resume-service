package database

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
)

type DB struct {
	client *mongo.Client
	User   UserStore
	Resume ResumeStore
}

const (
	dbName = "resume_service"
)

func NewClient(ctx context.Context) (*DB, error) {
	mongoUri := os.Getenv("MONGO_URI")
	if mongoUri == "" {
		mongoUri = "mongodb://0.0.0.0:27017"
	}
	connection, err := createConnection(ctx, mongoUri)
	if err != nil {
		return nil, err
	}
	database := connection.Database(dbName)
	userStore, err := newUserStore(database)
	resumeStore, err := newResumeStore(database)
	return &DB{
		client: connection,
		User:   userStore,
		Resume: resumeStore,
	}, nil
}

func createConnection(ctx context.Context, uri string) (*mongo.Client, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	return client, err
}

func (db *DB) Disconnect(ctx context.Context) error {
	return db.client.Disconnect(ctx)
}
