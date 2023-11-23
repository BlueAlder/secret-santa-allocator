package allocator

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/BlueAlder/secret-santa-allocator/pkg/utils"
)

// Allocation holds the aliases and allocations
// for a particular derangment.
type Allocation struct {
	Aliases     map[string]string `json:"aliases"`     // name -> password
	Allocations map[string]string `json:"allocations"` // name -> name
	Created     time.Time
	Players     []*Player `json:"-"`
}

type Player struct {
	Name     string
	Alias    string
	SantaFor *Player
	Santa    *Player
}

func newAllocation(playerNames []string) *Allocation {

	a := &Allocation{
		Aliases:     make(map[string]string),
		Allocations: make(map[string]string),
		Players:     make([]*Player, 0),
		Created:     time.Now().Local(),
	}

	for _, name := range playerNames {
		a.Players = append(a.Players, &Player{Name: name})
	}
	return a
}

func (a *Allocation) AssignAliases(passwords []string) {
	for _, player := range a.Players {
		password, randIdx := utils.RandomElementFromSlice(passwords)
		player.Alias = password
		passwords, _ = utils.RemoveIndex(passwords, randIdx)
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

type FlatAllocation struct {
	Aliases     map[string]string `json:"aliases"`     // name -> password
	Allocations map[string]string `json:"allocations"` // name -> name
}

func (a *Allocation) getFlatAllocationAndAliases() FlatAllocation {

	aliases := make(map[string]string)
	allocations := make(map[string]string)

	for _, player := range a.Players {
		aliases[player.Name] = player.Alias
		allocations[player.Name] = player.SantaFor.Name
	}

	return FlatAllocation{Aliases: aliases, Allocations: allocations}
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

func (a *Allocation) toJson() ([]byte, error) {
	jsonData, err := json.Marshal(a)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal json data got: %w", err)
	}
	return jsonData, nil
}
