package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/BlueAlder/secret-santa-allocator/pkg/allocator"
	"github.com/BlueAlder/secret-santa-allocator/pkg/keygenerator"
)

type CreateResponse struct {
	GameURL            string            `json:"gameURL"`
	AllocatedPasswords map[string]string `json:"allocatedPasswords"`
}

func (s *Server) createHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	s.logger.Info("incoming request",
		"method", r.Method,
		"path", r.URL.Path,
		"remote_addr", r.RemoteAddr,
	)

	if r.Method != http.MethodPost {
		s.logger.Warn("method not allowed", "method", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var config allocator.Config
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		s.logger.Error("failed to decode request body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create allocator from config
	s.logger.Debug("creating allocator", "allocation_name", config.Name)
	a, err := allocator.NewFromConfig(&config)
	if err != nil {
		s.logger.Error("failed to create allocator", "error", err)
		http.Error(w, fmt.Sprintf("Failed to create allocator: %v", err), http.StatusBadRequest)
		return
	}

	// Generate allocations
	s.logger.Debug("generating allocations")
	allocation, err := a.Allocate()
	if err != nil {
		s.logger.Error("failed to generate allocations", "error", err)
		http.Error(w, fmt.Sprintf("Failed to generate allocations: %v", err), http.StatusInternalServerError)
		return
	}

	// Generate key
	id := keygenerator.Generate(6)
	filename := fmt.Sprintf("%s.json", id)

	s.logger.Debug("generated allocation ID", "id", id)

	// Get allocation bytes
	data, err := a.OutputToBytes(allocation, "json")
	if err != nil {
		s.logger.Error("failed to marshal allocation", "error", err, "id", id)
		http.Error(w, fmt.Sprintf("Failed to marshal allocation: %v", err), http.StatusInternalServerError)
		return
	}

	// Save allocation
	s.logger.Debug("saving allocation", "id", id, "filename", filename, "size_bytes", len(data))
	if err := s.storage.Save(r.Context(), filename, data); err != nil {
		s.logger.Error("failed to save allocation", "error", err, "id", id, "filename", filename)
		http.Error(w, fmt.Sprintf("Failed to save allocation: %v", err), http.StatusInternalServerError)
		return
	}
	s.logger.Info("saved allocation", "id", id, "filename", filename, "size_bytes", len(data))

	// Return response
	gameUrl := os.Getenv("GAME_URL")
	if gameUrl == "" {
		gameUrl = fmt.Sprintf("http://localhost:%s/game/%s", s.port, id)
	} else {
		gameUrl = fmt.Sprintf("%s/%s", gameUrl, id)
	}
	response := CreateResponse{
		GameURL:            gameUrl,
		AllocatedPasswords: allocation.AllocatedPasswords(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Error("failed to encode response", "error", err, "id", id)
		return
	}

	duration := time.Since(start)
	s.logger.Info("request completed successfully",
		"id", id,
		"game_url", gameUrl,
		"duration_ms", duration.Milliseconds(),
	)
}
