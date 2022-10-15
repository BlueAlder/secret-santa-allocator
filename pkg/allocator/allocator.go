package allocator

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/BlueAlder/secret-santa-allocator/pkg/utils"
)

type Set map[string]struct{}

type Allocator struct {
	Names          Set // using maps here to dedupe the list
	Passwords      Set
	Config         Config
	lastAllocation Allocation
}

// creates a new instance of an Allocator
// takes a slice of names and passwords to distribute to eachother
func New(config *Config) (*Allocator, error) {
	a := &Allocator{
		Names:     make(Set),
		Passwords: make(Set),
		Config:    *config,
	}

	// Load names
	var undupedNames []string
	if config.Names.File != "" {
		err := utils.ReadFileIntoSlice(config.Names.File, &undupedNames)
		if err != nil {
			return nil, fmt.Errorf("error while loading names from file: %v", err)
		}
	}
	undupedNames = append(undupedNames, config.Names.Data...)
	for _, name := range undupedNames {
		a.Names[strings.TrimSpace(name)] = struct{}{}
	}

	// Load passwords
	var undupedPasswords []string
	if config.Passwords.File != "" {
		err := utils.ReadFileIntoSlice(config.Passwords.File, &undupedPasswords)
		if err != nil {
			return nil, fmt.Errorf("error while loading passwords from file: %v", err)
		}
	}
	undupedPasswords = append(undupedPasswords, config.Passwords.Data...)
	for _, password := range undupedPasswords {
		a.Passwords[strings.TrimSpace(password)] = struct{}{}
	}
	return a, nil
}

// Will allocate the names to a password and then the
// password to a name to create anonymity
func (a *Allocator) Allocate() (*Allocation, error) {
	if err := a.validateSetup(); err != nil {
		return nil, err
	}

	allocation := newAllocation()
	a.createAliases(allocation)
	err := a.createAllocations(allocation)

	if err != nil {
		return nil, err
	}

	// Map each name to a password
	a.lastAllocation = *allocation
	return allocation, nil
}

// using the names and passwords in the allocator
// map each name to an alias password
func (a *Allocator) createAliases(alloc *Allocation) {
	remainingPasswords := utils.MapKeysToSlice(a.Passwords)
	for name := range a.Names {
		password, randIdx := utils.RandomElementFromSlice(remainingPasswords)
		alloc.aliases[name] = password
		remainingPasswords = utils.RemoveIndex(remainingPasswords, randIdx)
	}
}

func (a *Allocator) createAllocations(alloc *Allocation) error {
	ctx, cancel := context.WithTimeout(context.Background(), a.Config.Timeout)
	complete := make(chan map[string]string)
	for i := 0; i < 5; i++ {
		go func() {
			remainingSantas := utils.MapKeysToSlice(a.Names)
			allocations := make(map[string]string)
			for name := range a.Names {

			infinite:
				for {
					select {
					case <-ctx.Done():
						return
					default:
						santa, santaIdx := utils.RandomElementFromSlice(remainingSantas)
						if a.checkAllocationValidRuleset(santa, name) {
							allocations[santa] = name
							remainingSantas = utils.RemoveIndex(remainingSantas, santaIdx)
							break infinite
						}
					}

				}
			}
			complete <- allocations
		}()
	}

	select {
	case a := <-complete:
		alloc.allocations = a
		cancel()
		break
	case <-ctx.Done():
		cancel()
		return fmt.Errorf("unable to find a suitable allocation with the given rules within %s. May be impossible", a.Config.Timeout.String())
	}

	return nil
}

func (a *Allocator) checkAllocationValidRuleset(santa string, santee string) bool {
	// TODO: Implemenet this to check for rules being violated

	if santa == santee {
		return a.Config.CanAllocateSelf
	}

	rule, exists := a.Config.Rules[santa]
	// No rule for this santa so return true
	if !exists {
		return true
	}

	// Check for exclusion in rule
	for _, name := range rule.CannotGet {
		if name == santee {
			return false
		}
	}

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
