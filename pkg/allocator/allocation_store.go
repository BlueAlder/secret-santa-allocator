package allocator

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// AllocationStore is a struct which models how marshalled
// allocations should be stored as to hide the specific allocation and aliases
// but show allocated passwords
type AllocationStore struct {
	Allocation         string            `json:"allocation"` // base64 encoded version of Allocation
	AllocatedPasswords map[string]string `json:"allocated_passwords"`
	Created            time.Time         `json:"created"`
	Name               string            `json:"allocation_name"`
}

type FlatAllocation struct {
	Aliases     map[string]string `json:"aliases"`     // name -> password
	Allocations map[string]string `json:"allocations"` // name -> name
}

func getFlatAllocationAndAliases(a *allocation) FlatAllocation {

	aliases := make(map[string]string)
	allocations := make(map[string]string)

	for _, player := range a.players {
		aliases[player.name] = player.alias
		allocations[player.name] = player.santaFor.name
	}

	return FlatAllocation{Aliases: aliases, Allocations: allocations}
}

func newAllocationStore(a *allocation, name string) (*AllocationStore, error) {
	fa := getFlatAllocationAndAliases(a)
	data, err := json.Marshal(fa)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal json data got: %w", err)
	}
	enc := base64.StdEncoding.EncodeToString(data)

	as := &AllocationStore{
		Allocation:         enc,
		AllocatedPasswords: a.AllocatedPasswords(),
		Created:            a.Created,
		Name:               name,
	}

	return as, nil
}

// outputToBytes returns the allocation as bytes
func (a *AllocationStore) outputToBytes(fileType string) ([]byte, error) {
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

// outputToFile will save the allocation to a file in either
// json or yaml depending on the fileType
func (a *AllocationStore) ouputToFile(fileName string, fileType string) error {
	data, err := a.outputToBytes(fileType)
	if err != nil {
		return err
	}

	err = os.WriteFile(fileName, data, 0644)
	if err != nil {
		return fmt.Errorf("unable to write to file got: %w", err)
	}

	return nil
}
