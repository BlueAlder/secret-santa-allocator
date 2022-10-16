package allocator

import (
	"encoding/json"
	"fmt"
	"time"
)

// Allocation holds the aliases and allocations
// for a particular derangment.
type Allocation struct {
	Aliases     map[string]string `json:"aliases"`     // name -> password
	Allocations map[string]string `json:"allocations"` // name -> name
	Created     time.Time
}

func newAllocation() *Allocation {
	return &Allocation{
		Aliases:     make(map[string]string),
		Allocations: make(map[string]string),
		Created:     time.Now().Local(),
	}
}

// allocatedPasswords returns a map[string]string
// mapping each name and their assigned name (not their alias)
func (a *Allocation) allocatedPasswords() map[string]string {
	var passwordAllocations = make(map[string]string)
	for name, allocated := range a.Allocations {
		passwordAllocations[name] = a.Aliases[allocated]
	}
	return passwordAllocations
}

func (a *Allocation) PrintNameToPassword() {
	fmt.Println("Printing names to allocated passwords:")
	for name, password := range a.allocatedPasswords() {
		fmt.Printf("%s -> %s\n", name, password)
	}
}

func (a *Allocation) PrintNameToName() {
	fmt.Println("Printing names to allocated names:")
	for name, allocated := range a.Allocations {
		fmt.Printf("%s -> %s\n", name, allocated)
	}
}

func (a *Allocation) PrintAliases() {
	fmt.Println("Printing names aliases:")
	for name, password := range a.Aliases {
		fmt.Printf("%s -> %s\n", name, password)
	}
}

func (a *Allocation) String() string {
	res := fmt.Sprintf("Created at %s\n", a.Created.Format("01-02-2006 15:04:05"))
	res += "Aliases:\n"
	for name, password := range a.Aliases {
		res += fmt.Sprintf("%s -> %s\n", name, password)
	}
	res += "\n"

	res += "Allocations:\n"
	for password, name := range a.Allocations {
		res += fmt.Sprintf("%s -> %s\n", password, name)
	}
	return res
}

// OutputToFile writes an instance of Allocation to fileName
// with either "json" or "yaml" as the fileType
func (a *Allocation) OutputToFile(fileName string, fileType string) error {
	as, err := newAllocationStore(a)
	if err != nil {
		return err
	}
	return as.ouputToFile(fileName, fileType)
}

func (a *Allocation) toJson() ([]byte, error) {
	jsonData, err := json.Marshal(a)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal json data got: %w", err)
	}
	return jsonData, nil
}
