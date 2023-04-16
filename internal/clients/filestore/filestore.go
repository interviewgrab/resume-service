package filestore

import (
	"bytes"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const bucket = "resume-service-filestore"

type FileStore struct {
	s3 *s3.S3
}

func NewStorageClient() *FileStore {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	return &FileStore{s3: s3.New(sess)}
}

func (s *FileStore) Upload(key string, fileContent []byte) error {
	input := &s3.PutObjectInput{
		Body:   bytes.NewReader(fileContent),
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	_, err := s.s3.PutObject(input)
	return err
}

func (s *FileStore) Download(key string) ([]byte, error) {
	result, err := s.s3.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}

	return io.ReadAll(result.Body)
}
