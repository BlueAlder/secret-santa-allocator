package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/BlueAlder/secret-santa-allocator/pkg/allocator"
	"github.com/BlueAlder/secret-santa-allocator/pkg/allocator/gamestore"
	"github.com/BlueAlder/secret-santa-allocator/pkg/keygenerator"
	"github.com/BlueAlder/secret-santa-allocator/pkg/storage"
)

type CreateResponse struct {
	GameID             string            `json:"gameID"`
	AllocatedPasswords map[string]string `json:"allocatedPasswords"`
}

// POST /game
// Creates a new secret santa game and saves it to storage
// Returns a link to the game
func (s *Server) createHandler(w http.ResponseWriter, r *http.Request) {

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
		s.logger.Error("failed to generate game", "error", err)
		http.Error(w, fmt.Sprintf("Failed to generate game: %v", err), http.StatusBadRequest)
		return
	}

	// Generate key
	id := keygenerator.Generate(6)
	filename := fmt.Sprintf("%s.json", id)

	s.logger.Debug("generated allocation ID", "id", id)

	// Create new game from allocation
	game, err := gamestore.New(allocation, config.Name)
	if err != nil {
		s.logger.Error("failed to create game store", "error", err, "id", id)
		http.Error(w, fmt.Sprintf("Failed to create game store: %v", err), http.StatusInternalServerError)
		return
	}

	// Marshal game
	data, err := game.OutputToBytes("json")
	if err != nil {
		s.logger.Error("failed to marshal game", "error", err, "id", id)
		http.Error(w, fmt.Sprintf("Failed to marshal game: %v", err), http.StatusInternalServerError)
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
	response := CreateResponse{
		GameID:             id,
		AllocatedPasswords: allocation.AllocatedPasswords(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Error("failed to encode response", "error", err, "id", id)
		return
	}

}

type GetGameResponse struct {
	GameName string    `json:"gameName"`
	Created  time.Time `json:"created"`
	Players  []string  `json:"players"`
}

// GET /game/:id
// Returns the allocation with the given ID in JSON format
func (s *Server) getGameHandler(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("id")

	// Load game
	fileName := fmt.Sprintf("%s.json", id)
	data, err := s.storage.Load(r.Context(), fileName)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			http.Error(w, "Game not found", http.StatusNotFound)
			return
		}
		s.logger.Error("failed to load game", "error", err, "id", id)
		http.Error(w, fmt.Sprintf("Failed to load game: %v", err), http.StatusInternalServerError)
		return
	}

	// Unmarshal game
	var game gamestore.GameStore
	if err := json.Unmarshal(data, &game); err != nil {
		s.logger.Error("failed to unmarshal game", "error", err, "id", id)
		http.Error(w, fmt.Sprintf("Failed to unmarshal game: %v", err), http.StatusInternalServerError)
		return
	}

	// Return response
	response := GetGameResponse{
		GameName: game.Name,
		Created:  game.Created,
		Players:  game.Players(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Error("failed to encode response", "error", err, "id", id)
		return
	}

}

type CheckPasswordRequest struct {
	Password string `json:"password"`
}

type CheckPasswordResponse struct {
	Player string `json:"player"`
}

func (s *Server) checkPasswordHandler(w http.ResponseWriter, r *http.Request) {

	// Load password from body
	var req CheckPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Error("failed to decode request body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	id := r.PathValue("id")
	fmt.Print(id)

	// Load game
	fileName := fmt.Sprintf("%s.json", id)
	data, err := s.storage.Load(r.Context(), fileName)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			http.Error(w, "Game not found", http.StatusNotFound)
			return
		}
		s.logger.Error("failed to load game", "error", err, "id", id)
		http.Error(w, fmt.Sprintf("Failed to load game: %v", err), http.StatusInternalServerError)
		return
	}

	// Unmarshal game
	var game gamestore.GameStore
	if err := json.Unmarshal(data, &game); err != nil {
		s.logger.Error("failed to unmarshal game", "error", err, "id", id)
		http.Error(w, fmt.Sprintf("Failed to unmarshal game: %v", err), http.StatusInternalServerError)
		return
	}

	player, err := game.PasswordToPlayer(req.Password)
	if err != nil {
		s.logger.Error("incorrect password", "error", err, "id", id)
		http.Error(w, fmt.Sprintf("Incorrect password: %v", err), http.StatusBadRequest)
		return
	}

	// Return response
	response := CheckPasswordResponse{
		Player: player,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Error("failed to encode response", "error", err, "id", id)
		return
	}

}
