// Package allocator implements santa allocations
package allocator

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/BlueAlder/secret-santa-allocator/pkg/utils"
	"golang.org/x/exp/slices"
)

// Set is a makeshift set using a map for deduping purposes
type Set map[string]struct{}

// Allocator creates a new allocation given a particular config
type Allocator struct {
	names          []string
	passwords      []string
	lastAllocation Allocation
	// maps names to names they cannot be assigned (rules)
	exclusionRules map[string][]string
	// maps names to names they must be assigned (rules)
	mustGetRules map[string]string
	// Timeout for allocation to complete before failing
	timeout         time.Duration
	CanAllocateSelf bool
	// Name of the allocation e.g Friendmas 2024
	Name string
}

// Creates a new allocator with default values with a config
func New() *Allocator {
	return &Allocator{
		names:           []string{},
		passwords:       []string{},
		CanAllocateSelf: false,
		timeout:         5 * time.Second,
		exclusionRules:  make(map[string][]string),
		mustGetRules:    make(map[string]string),
		Name:            "",
	}
}

// creates a new instance of an Allocator
// takes a config which is used to setup the allocator
func NewFromConfig(config *Config) (*Allocator, error) {
	a := &Allocator{
		names:           []string{},
		passwords:       []string{},
		CanAllocateSelf: config.CanAllocateSelf,
		timeout:         config.Timeout,
		exclusionRules:  make(map[string][]string),
		mustGetRules:    make(map[string]string),
		Name:            config.Name,
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
	var formattedUndupedNames []string

	for _, name := range undupedNames {
		formattedName := strings.ToLower(strings.TrimSpace(name))
		formattedUndupedNames = append(formattedUndupedNames, formattedName)
	}

	a.names = utils.RemoveDuplicatesFromSlice(formattedUndupedNames)

	// Load passwords
	var undupedPasswords []string
	if config.Passwords.File != "" {
		err := utils.ReadFileIntoSlice(config.Passwords.File, &undupedPasswords)
		if err != nil {
			return nil, fmt.Errorf("error while loading passwords from file: %v", err)
		}
	}
	undupedPasswords = append(undupedPasswords, config.Passwords.Data...)
	a.passwords = utils.RemoveDuplicatesFromSlice(undupedPasswords)

	// Load exclusionRules
	for _, rule := range config.Rules {
		for _, bannedName := range rule.CannotGet {
			// check if the name is in the list of names
			if !slices.Contains(a.names, bannedName) {
				return nil, fmt.Errorf("name [%s] in exclusion rule is not in the list of names", bannedName)
			}
			if !slices.Contains(a.names, rule.Name) {
				return nil, fmt.Errorf("name [%s] in exclusion rule is not in the list of names", rule.Name)
			}

			a.exclusionRules[strings.ToLower(rule.Name)] = append(a.exclusionRules[strings.ToLower(rule.Name)], strings.ToLower(bannedName))
			if rule.Inverse {
				a.exclusionRules[strings.ToLower(bannedName)] = append(a.exclusionRules[strings.ToLower(bannedName)], strings.ToLower(rule.Name))
			}
		}
	}

	// Load mustGet Rules
	for _, rule := range config.Rules {
		if rule.MustGet != "" {
			if !slices.Contains(a.names, rule.MustGet) {
				return nil, fmt.Errorf("name [%s] in must get rule is not in the list of names", rule.MustGet)
			}
			a.mustGetRules[strings.ToLower(rule.Name)] = strings.ToLower(rule.MustGet)
		}
	}

	return a, nil
}

// Allocate will allocate the names to a password and then the
// password to a name to create anonymity
func (a *Allocator) Allocate() (*Allocation, error) {
	if err := a.validateSetup(); err != nil {
		return nil, fmt.Errorf("invalid allocator setup: %w", err)
	}

	if err := a.validateRules(); err != nil {
		return nil, fmt.Errorf("invalid allocator rules: %w", err)
	}

	// a.createAliases(allocation)
	allocation, err := a.createAllocations()

	if err != nil {
		return nil, err
	}

	a.lastAllocation = *allocation
	return allocation, nil
}

// createAliases will use the names and passwords in the allocator
// and map each name to a random alias password
// func (a *Allocator) createAliases(alloc *Allocation) {

// 	remainingPasswords := utils.MapKeysToSlice(a.Passwords)
// 	for name := range a.Names {
// 		password, randIdx := utils.RandomElementFromSlice(remainingPasswords)
// 		alloc.Aliases[name] = password
// 		remainingPasswords, _ = utils.RemoveIndex(remainingPasswords, randIdx)
// 	}
// }

// createAllocations will spin up 5 goroutines to find
// an allocation that meets all the requirements of the config
// returns and error if it cannot find a valid one within the configured
// timeout value
func (a *Allocator) createAllocations() (*Allocation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), a.timeout)
	allocationsChan := make(chan *Allocation)
	errorChan := make(chan error)
	for i := 0; i < 15; i++ {
		go func() {
			alloc := NewAllocation(a.names, a.passwords)

			remainingSantas := make([]*Player, len(alloc.Players))
			copy(remainingSantas, alloc.Players)

			// Handle must get rules
			for santa, mustGet := range a.mustGetRules {
				santaIdx := slices.IndexFunc(alloc.Players, func(p *Player) bool { return p.Name == santa })
				santeeIdx := slices.IndexFunc(alloc.Players, func(p *Player) bool { return p.Name == mustGet })
				if santaIdx == -1 {
					errorChan <- fmt.Errorf("cannot allocate [%s] to [%s] as [%s] is not in the list of names", mustGet, santa, santa)
					return
				}

				santa := alloc.Players[santaIdx]
				santee := alloc.Players[santeeIdx]

				if santa.SantaFor != nil || santee.Santa != nil {
					errorChan <- fmt.Errorf("cannot allocate [%s] to [%s] as they are already allocated", santa.Name, santee.Name)
					return
				}

				santa.SantaFor = santee
				santee.Santa = santa
				remainingSantas, _ = utils.RemoveIndex(remainingSantas, santaIdx)
			}

			// Allocate the rest of the config

			for _, santee := range alloc.Players {

			infinite:
				for {
					select {
					case <-ctx.Done():
						return
					default:
						// Player has already been allocated to someone
						if santee.Santa != nil {
							break infinite
						}
						santa, santaIdx := utils.RandomElementFromSlice(remainingSantas)
						// Double check santa doesn't already have someone
						if a.checkAllocationValidRuleset(santa, santee) {
							santa.SantaFor = santee
							santee.Santa = santa
							remainingSantas, _ = utils.RemoveIndex(remainingSantas, santaIdx)
							break infinite
						}
					}

				}
			}
			allocationsChan <- alloc
		}()
	}

	select {
	case alloc := <-allocationsChan:
		// Allocation succesfull, close channels.
		cancel()
		return alloc, nil
	case e := <-errorChan:
		cancel()
		return nil, fmt.Errorf("error while allocating: %v", e)
	case <-ctx.Done():
		cancel()
		return nil, fmt.Errorf("unable to find a suitable allocation with the given rules within %s. May be impossible ❌", a.timeout.String())
	}

}

// checkAllocationVaildRuleset will return whether a given santa
// is able to be allocated to a given santee, given the allocators
// ruleset.
func (a *Allocator) checkAllocationValidRuleset(santa *Player, santee *Player) bool {
	santaName := strings.ToLower(santa.Name)
	santeeName := strings.ToLower(santee.Name)

	if santa == santee {
		return a.CanAllocateSelf
	}

	excludedNames := a.exclusionRules[santaName]
	if excludedNames == nil {
		// No rule for this santa so is valid
		return true
	}

	for _, name := range excludedNames {
		if strings.ToLower(name) == santeeName {
			return false
		}
	}

	return true
}

// validateSetup ensures that a give Allocator
// has enough names and passwords to create an
// allocation. Does not check rules.
func (a *Allocator) validateSetup() error {
	if len(a.names) < 2 {
		return errors.New("need at least 2 names")
	}

	if len(a.passwords) < 2 {
		return errors.New("need at least 2 passwords")
	}

	if len(a.names) > len(a.passwords) {
		return errors.New("there must be the same or more passwords than names")
	}

	return nil
}

// validateRules ensures that the rules in the config are valid and an allocation
// is possible. However this is not perfect.
func (a *Allocator) validateRules() error {
	// Check must get rules are all unique
	allocatedNames := make(Set)
	for _, mustGet := range a.mustGetRules {
		if _, ok := allocatedNames[mustGet]; ok {
			return fmt.Errorf("name [%s] is in multiple must get rules", mustGet)
		}
		allocatedNames[mustGet] = struct{}{}
	}

	// Check mustGet is not in exclusion

	// Check exclusion list is not longer than the list of names
	for name, excludedNames := range a.exclusionRules {
		if len(excludedNames) > len(a.names) {
			return fmt.Errorf("name [%s] has more exclusion rules than names", name)
		}
	}
	return nil
}

// OutputToFile writes an instance of Allocation to fileName
// with either "json" or "yaml" as the fileType
func (a *Allocator) OutputToFile(allocation *Allocation, fileName string, fileType string) error {
	as, err := newAllocationStore(allocation, a.Name)
	if err != nil {
		return err
	}
	return as.ouputToFile(fileName, fileType)
}
