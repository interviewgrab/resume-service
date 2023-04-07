package database

import (
	"context"
	"resume-service/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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

func (s *ResumeStore) DeleteResume(ctx context.Context, userId primitive.ObjectID, id string) error {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	filter := bson.M{"_id": objectId, "user_id": userId}
	_, err = s.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	return err
}

func (s *ResumeStore) GetResumesByUserId(ctx context.Context, userId primitive.ObjectID) ([]model.Resume, error) {
	cursor, err := s.collection.Find(ctx, bson.M{"user_id": userId})
	if err != nil {
		return nil, err
	}

	var resumes []model.Resume
	if err = cursor.All(ctx, &resumes); err != nil {
		return nil, err
	}

	return resumes, nil
}

func (s *ResumeStore) UpdateUserResumeIsPublic(ctx context.Context, userId primitive.ObjectID, id string, isPublic bool) error {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	result := s.collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": objectId, "user_id": userId},
		bson.M{"$set": bson.M{"public": isPublic}},
	)
	return result.Err()
}
