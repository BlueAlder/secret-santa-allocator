package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/BlueAlder/secret-santa-allocator/pkg/storage"
)

type Server struct {
	storage     storage.Storage
	host        string
	port        string
	storageType string
	logger      *slog.Logger
}

func createServer() (*Server, error) {
	ctx := context.Background()
	var store storage.Storage
	var err error

	// Initialize logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	server := &Server{
		logger: logger,
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	server.port = port

	host := os.Getenv("HOST")
	if host == "" {
		host = "localhost"
	}
	server.host = host

	storageType := os.Getenv("STORAGE_TYPE")
	switch storageType {
	case "bucket":
		bucketName := os.Getenv("BUCKET_NAME")
		if bucketName == "" {
			return nil, fmt.Errorf("BUCKET_NAME env var is required for cloud storage")
		}
		logger.Info("initializing GCS storage", "bucket", bucketName)
		store, err = storage.NewGCSStorage(ctx, bucketName)
		server.storageType = "bucket"
	default:
		// Default to disk
		logger.Info("initializing disk storage", "directory", "data")
		store, err = storage.NewDiskStorage("data")
		server.storageType = "disk"
	}
	server.storage = store

	if err != nil {
		logger.Error("failed to initialize storage", "error", err)
		return nil, fmt.Errorf("failed to initialize storage: %v", err)
	}

	logger.Info("storage initialized successfully", "type", server.storageType)

	return server, nil
}

func main() {

	server, err := createServer()
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	http.HandleFunc("/create", server.createHandler)

	server.logger.Info("server starting",
		"host", server.host,
		"port", server.port,
		"storage_type", server.storageType,
	)
	fmt.Printf("Server starting on %s:%s with storage type: %s...\n", server.host, server.port, server.storageType)
	if err := http.ListenAndServe(server.host+":"+server.port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
