package storage

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("not found")

// Storage is the interface for storing allocations
type Storage interface {
	Save(ctx context.Context, key string, data []byte) error
	Load(ctx context.Context, key string) ([]byte, error)
}
