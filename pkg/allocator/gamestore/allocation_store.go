package gamestore

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/BlueAlder/secret-santa-allocator/pkg/allocator"
	"gopkg.in/yaml.v3"
)

// GameStore is a struct which models how marshalled
// allocations should be stored as to hide the specific allocation and aliases
// but show allocated passwords
type GameStore struct {
	Allocations map[string]string `json:"allocations"` //  santa -> santee - should be private
	Aliases     map[string]string `json:"aliases"`     // password -> name - should be public
	Created     time.Time         `json:"created"`
	Name        string            `json:"allocation_name"`
}

func New(a *allocator.Allocation, name string) (*GameStore, error) {
	as := &GameStore{
		Allocations: a.Allocations(),
		Aliases:     a.Aliases(),
		Created:     a.Created,
		Name:        name,
	}

	return as, nil
}

func (a *GameStore) Players() []string {
	var players []string
	for k := range a.Aliases {
		players = append(players, k)
	}
	return players
}

func (a *GameStore) PasswordToPlayer(password string) (string, error) {
	player, ok := a.Aliases[password]
	if !ok {
		return "", fmt.Errorf("password incorrect")
	}
	return player, nil
}

// OutputToBytes returns the allocation as bytes
func (a *GameStore) OutputToBytes(fileType string) ([]byte, error) {
	var data []byte
	var err error
	switch fileType {
	case "yaml":
		data, err = yaml.Marshal(a)
	case "json":
		data, err = json.MarshalIndent(a, "", "\t")
	default:
		return nil, fmt.Errorf("OutputToBytes invalid file type: %s", fileType)
	}

	if err != nil {
		return nil, fmt.Errorf("unable to marshal data got: %w", err)
	}
	return data, nil
}

// OutputToFile will save the allocation to a file in either
// json or yaml depending on the fileType
func (a *GameStore) OutputToFile(fileName string, fileType string) error {
	data, err := a.OutputToBytes(fileType)
	if err != nil {
		return err
	}

	err = os.WriteFile(fileName, data, 0644)
	if err != nil {
		return fmt.Errorf("unable to write to file got: %w", err)
	}

	return nil
}
