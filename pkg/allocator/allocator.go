package allocator

import (
	"errors"
	"math/rand"
	"time"

	"github.com/BlueAlder/secret-santa-allocator/pkg/utils"
)

type Allocator struct {
	Names     []string
	Passwords []string
	// Rulesets       []Ruleset
	lastAllocation Allocation
}

// creates a new instance of an Allocator
// takes a slice of names and passwords to distribute to eachother
func New(names []string, passwords []string) *Allocator {
	return &Allocator{
		Names:     names,
		Passwords: passwords,
	}
}

// Will allocate the names to a password and then the
// password to a name to create anonymity
func (a *Allocator) Allocate() (*Allocation, error) {
	if err := a.validateSetup(); err != nil {
		return nil, err
	}

	allocation := newAllocation()
	a.createAliases(allocation)
	a.createAllocations(allocation)

	// Map each name to a password

	a.lastAllocation = *allocation

	return allocation, nil
}

// using the names and passwords in the allocator
// map each name to an alias password
func (a *Allocator) createAliases(alloc *Allocation) {
	rand.Seed(time.Now().Unix())
	remainingPasswords := a.Passwords
	for _, name := range a.Names {
		password, randIdx := utils.RandomElement(remainingPasswords)
		alloc.aliases[name] = password
		remainingPasswords = utils.RemoveIndex(remainingPasswords, randIdx)
	}
}

func (a *Allocator) createAllocations(alloc *Allocation) {
	remainingSantas := a.Names
	for _, name := range a.Names {
		for {
			santa, santaIdx := utils.RandomElement(remainingSantas)
			if a.checkAllocationValidRuleset(santa, name) {
				// get santa alias
				alias := alloc.aliases[santa]
				alloc.allocations[name] = alias
				remainingSantas = utils.RemoveIndex(remainingSantas, santaIdx)
				break
			}
		}
	}
}

func (a *Allocator) checkAllocationValidRuleset(santa string, santee string) bool {
	// TODO: Implemenet this to check for rules being violated
	return true
}

func (a *Allocator) validateSetup() error {
	if len(a.Names) < 2 {
		return errors.New("need at least 2 names")
	}

	if len(a.Passwords) < 2 {
		return errors.New("need at least 2 passwords")
	}

	if len(a.Names) > len(a.Passwords) {
		return errors.New("there must be the same or more passwords than names")
	}

	return nil
}
