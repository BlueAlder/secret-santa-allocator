package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

type DiskStorage struct {
	BaseDir string
}

func NewDiskStorage(baseDir string) (*DiskStorage, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}
	return &DiskStorage{BaseDir: baseDir}, nil
}

func (s *DiskStorage) Save(ctx context.Context, key string, data []byte) error {
	filename := filepath.Join(s.BaseDir, key)
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}
