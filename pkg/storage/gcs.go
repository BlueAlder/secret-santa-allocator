package storage

import (
	"context"
	"fmt"

	"cloud.google.com/go/storage"
)

type GCSStorage struct {
	BucketName string
	Client     *storage.Client
}

func NewGCSStorage(ctx context.Context, bucketName string) (*GCSStorage, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}
	return &GCSStorage{
		BucketName: bucketName,
		Client:     client,
	}, nil
}

func (s *GCSStorage) Save(ctx context.Context, key string, data []byte) error {
	bucket := s.Client.Bucket(s.BucketName)
	obj := bucket.Object(key)
	w := obj.NewWriter(ctx)
	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("failed to write to GCS: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close GCS writer: %w", err)
	}
	return nil
}
