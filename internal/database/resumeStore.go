package database

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"resume-service/internal/model"
)

type ResumeStore struct {
	collection *mongo.Collection
}

const resumeCollection = "resumeCollection"

func newResumeStore(dbClient *mongo.Database) (ResumeStore, error) {
	return ResumeStore{collection: dbClient.Collection(resumeCollection)}, nil
}

func (s *ResumeStore) StoreResume(ctx context.Context, resume model.Resume) error {
	_, err := s.collection.InsertOne(ctx, resume)
	return err
}

func (s *ResumeStore) GetResume(ctx context.Context, id string) (model.Resume, error) {
	resume := &model.Resume{}
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return model.Resume{}, err
	}
	err = s.collection.FindOne(ctx, bson.M{"_id": objectId}).Decode(resume)
	if err != nil {
		return model.Resume{}, err
	}
	return *resume, err
}
