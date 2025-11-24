package storage

import (
	"context"
	"fmt"
	"io"

	"cloud.google.com/go/storage"
)

type GCSStorage struct {
	BucketName string
	Client     *storage.Client
}

// NewGCSStorage creates a new GCSStorage instance provided a given bucketname
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

// Save saves an allocation to the configured GCS Bucket
func (s *GCSStorage) Save(ctx context.Context, key string, data []byte) error {
	bucket := s.Client.Bucket(s.BucketName)
	obj := bucket.Object(key)
	w := obj.NewWriter(ctx)
	w.ContentType = "application/json"
	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("failed to write to GCS: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close GCS writer: %w", err)
	}
	return nil
}

// Load loads an allocation from the configured GCS Bucket
func (s *GCSStorage) Load(ctx context.Context, key string) ([]byte, error) {
	bucket := s.Client.Bucket(s.BucketName)
	obj := bucket.Object(key)
	reader, err := obj.NewReader(ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to create GCS reader: %w", err)
	}
	defer reader.Close()
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read from GCS: %w", err)
	}
	return data, nil
}
