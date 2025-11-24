package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

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
		port = "3000"
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

	server.registerHandlers()

	return server, nil
}

func (s *Server) registerHandlers() {

	http.HandleFunc("POST /api/game", s.createHandler)

	http.HandleFunc("GET /api/game/{id}", s.getGameHandler)
	http.HandleFunc("POST /api/game/{id}", s.checkPasswordHandler)

}

func (s *Server) Run() error {
	return http.ListenAndServe(fmt.Sprintf("%s:%s", s.host, s.port), s.loggingMiddleware(http.DefaultServeMux))
}

type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		s.logger.Info("incoming request",
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
		)

		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)

		duration := time.Since(start)

		s.logger.Debug("request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
			"status", rw.status,
			"size_bytes", rw.size,
			"duration_ms", duration.Milliseconds(),
		)
	})
}
