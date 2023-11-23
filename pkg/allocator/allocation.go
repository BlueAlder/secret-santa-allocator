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
	Players []*Player `json:"-"`
}

type Player struct {
	Name     string
	Alias    string
	SantaFor *Player
	Santa    *Player
}

func newAllocation(playerNames []string) *Allocation {

	a := &Allocation{
		Players: make([]*Player, 0),
		Created: time.Now().Local(),
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
	for _, player := range a.Players {
		passwordAllocations[player.Name] = player.SantaFor.Alias
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
	for _, player := range a.Players {
		fmt.Printf("%s -> %s\n", player.Name, player.SantaFor.Name)
	}
}

func (a *Allocation) PrintAliases() {
	fmt.Println("Printing names aliases:")
	for _, player := range a.Players {
		fmt.Printf("%s -> %s\n", player.Name, player.Alias)
	}
}

func (a *Allocation) String() string {
	res := fmt.Sprintf("Created at %s\n", a.Created.Format("01-02-2006 15:04:05"))
	res += "Aliases:\n"
	for _, player := range a.Players {
		res += fmt.Sprintf("%s -> %s\n", player.Name, player.Alias)
	}
	res += "\n"

	res += "Allocations:\n"
	for _, player := range a.Players {
		res += fmt.Sprintf("%s -> %s\n", player.Name, player.SantaFor.Name)
	}
	return res
}
