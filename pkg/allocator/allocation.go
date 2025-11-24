package allocator

import (
	"fmt"
	"time"

	"github.com/BlueAlder/secret-santa-allocator/pkg/utils"
)

// Allocation holds the aliases and allocations
// for a particular derangment.
type Allocation struct {
	Created time.Time
	players []*player `json:"-"`
}

// player represents a node in the directed graph of players
// SantaFor is the player that the player is buying a gift for
// Santa is the player that is buying a gift for the player
type player struct {
	name     string
	alias    string
	santaFor *player
	santa    *player
}

func NewAllocation(playerNames []string, passwords []string) *Allocation {

	a := &Allocation{
		players: make([]*player, 0),
		Created: time.Now().Local(),
	}

	for _, name := range playerNames {
		a.players = append(a.players, &player{name: name})
	}

	// Assign aliases
	a.assignAliases(passwords)
	return a
}

func (a *Allocation) assignAliases(passwords []string) {
	for _, player := range a.players {
		password, randIdx := utils.RandomElementFromSlice(passwords)
		player.alias = password
		passwords, _ = utils.RemoveIndex(passwords, randIdx)
	}
}

func (a *Allocation) GetPlayer(name string) *player {
	for _, player := range a.players {
		if player.name == name {
			return player
		}
	}
	return nil
}

// AllocatedPasswords returns a map[string]string
// mapping each name and their assigned player's alias
func (a *Allocation) AllocatedPasswords() map[string]string {
	var passwordAllocations = make(map[string]string)
	for _, player := range a.players {
		passwordAllocations[player.name] = player.santaFor.alias
	}
	return passwordAllocations
}

// Aliases returns a map[string]string
// mapping each name and their assigned alias
func (a *Allocation) Aliases() map[string]string {
	var aliases = make(map[string]string)
	for _, player := range a.players {
		aliases[player.alias] = player.name
	}
	return aliases
}

// Allocations returns a map[string]string
// mapping each players name and their assigned player's name (not their alias)
func (a *Allocation) Allocations() map[string]string {
	var nameAllocations = make(map[string]string)
	for _, player := range a.players {
		nameAllocations[player.name] = player.santaFor.name
	}
	return nameAllocations
}

func (a *Allocation) PrintNameToPassword() {
	fmt.Println("Printing names to allocated passwords:")
	for name, password := range a.AllocatedPasswords() {
		fmt.Printf("%s -> %s\n", name, password)
	}
}

func (a *Allocation) PrintNameToName() {
	fmt.Println("Printing names to allocated names:")
	for _, player := range a.players {
		fmt.Printf("%s -> %s\n", player.name, player.santaFor.name)
	}
}

func (a *Allocation) PrintAliases() {
	fmt.Println("Printing names aliases:")
	for _, player := range a.players {
		fmt.Printf("%s -> %s\n", player.name, player.alias)
	}
}

func (a *Allocation) String() string {
	res := fmt.Sprintf("Created at %s\n", a.Created.Format("01-02-2006 15:04:05"))
	res += "Aliases:\n"
	for _, player := range a.players {
		res += fmt.Sprintf("%s -> %s\n", player.name, player.alias)
	}
	res += "\n"

	res += "Allocations:\n"
	for _, player := range a.players {
		res += fmt.Sprintf("%s -> %s\n", player.name, player.santaFor.name)
	}
	return res
}
