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

func newAllocationStore(a *Allocation, name string) (*AllocationStore, error) {
	data, err := a.toJson()
	if err != nil {
		return nil, err
	}
	enc := base64.StdEncoding.EncodeToString(data)

	as := &AllocationStore{
		Allocation:         enc,
		AllocatedPasswords: a.allocatedPasswords(),
		Created:            a.Created,
		Name:               name,
	}

	return as, nil
}

// outputToFile will save the allocation to a file in either
// json or yaml depending on the fileType
func (a *AllocationStore) ouputToFile(fileName string, fileType string) error {
	var data []byte
	var err error
	switch fileType {
	case "yaml":
		data, err = yaml.Marshal(a)
	case "json":
		data, err = json.MarshalIndent(a, "", "\t")
	default:
		return fmt.Errorf("OutputToFile invalid file type: %s", fileType)
	}

	if err != nil {
		return fmt.Errorf("unable to marshal data got: %w", err)
	}

	err = os.WriteFile(fileName, data, 0644)
	if err != nil {
		return fmt.Errorf("unable to write to file got: %w", err)
	}

	return nil
}
