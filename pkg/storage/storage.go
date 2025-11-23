package storage

import "context"

// Storage is the interface for storing allocations
type Storage interface {
	Save(ctx context.Context, key string, data []byte) error
}
