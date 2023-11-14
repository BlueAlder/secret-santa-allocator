// Package allocator implements santa allocations
package allocator

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/BlueAlder/secret-santa-allocator/pkg/utils"
	"golang.org/x/exp/slices"
)

// Set is a makeshift set using a map for deduping purposes
type Set map[string]struct{}

// Allocator creates a new allocation given a particular config
type Allocator struct {
	Names          Set // using maps here to dedupe the list
	Passwords      Set
	Config         Config
	lastAllocation Allocation
}

// creates a new instance of an Allocator
// takes a config which is used to setup the allocator
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
			return nil, fmt.Errorf("error while loading names from file: %w", err)
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

// Allocate will allocate the names to a password and then the
// password to a name to create anonymity
func (a *Allocator) Allocate() (*Allocation, error) {
	if err := a.validateSetup(); err != nil {
		return nil, fmt.Errorf("invalid allocator setup: %w", err)
	}

	allocation := newAllocation()
	a.createAliases(allocation)
	err := a.createAllocations(allocation)

	if err != nil {
		return nil, err
	}

	a.lastAllocation = *allocation
	return allocation, nil
}

// createAliases will use the names and passwords in the allocator
// and map each name to a random alias password
func (a *Allocator) createAliases(alloc *Allocation) {
	remainingPasswords := utils.MapKeysToSlice(a.Passwords)
	for name := range a.Names {
		password, randIdx := utils.RandomElementFromSlice(remainingPasswords)
		alloc.Aliases[name] = password
		remainingPasswords = utils.RemoveIndex(remainingPasswords, randIdx)
	}
}

// createAllocations will spin up 5 goroutines to find
// an allocation that meets all the requirements of the config
// returns and error if it cannot find a valid one within the configured
// timeout value
func (a *Allocator) createAllocations(alloc *Allocation) error {
	ctx, cancel := context.WithTimeout(context.Background(), a.Config.Timeout)
	allocationsChan := make(chan map[string]string)
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
			allocationsChan <- allocations
		}()
	}

	select {
	case a := <-allocationsChan:
		fmt.Println("found a suitable allocation! ✅")
		alloc.Allocations = a
		cancel()
		break
	case <-ctx.Done():
		cancel()
		return fmt.Errorf("unable to find a suitable allocation with the given rules within %s. May be impossible ❌", a.Config.Timeout.String())
	}

	return nil
}

// checkAllocationVaildRuleset will return whether a given santa
// is able to be allocated to a given santee, given the allocators
// ruleset.
func (a *Allocator) checkAllocationValidRuleset(santa string, santee string) bool {
	santa = strings.ToLower(santa)
	santee = strings.ToLower(santee)

	if santa == santee {
		return a.Config.CanAllocateSelf
	}

	idx := slices.IndexFunc(a.Config.Rules, func(r Rule) bool { return strings.ToLower(r.Name) == santa })

	// No rule for this santa so is valid
	if idx < 0 {
		return true
	}
	rule := a.Config.Rules[idx]
	// Check for exclusion in rule
	for _, name := range rule.CannotGet {
		if strings.ToLower(name) == santee {
			return false
		}
	}

	return true
}

// validateSetup ensures that a give Allocator
// has enough names and passwords to create an
// allocation. Does not check rules.
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

// OutputToFile writes an instance of Allocation to fileName
// with either "json" or "yaml" as the fileType
func (a *Allocator) OutputToFile(allocation *Allocation, fileName string, fileType string) error {
	as, err := newAllocationStore(allocation, a.Config.Name)
	if err != nil {
		return err
	}
	return as.ouputToFile(fileName, fileType)
}
