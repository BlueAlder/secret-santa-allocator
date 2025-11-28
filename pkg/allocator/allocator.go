// Package allocator implements santa allocations
package allocator

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/BlueAlder/secret-santa-allocator/pkg/utils"
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
func NewFromConfig(conf *Config) (*Allocator, error) {
	a := &Allocator{
		names:           []string{},
		passwords:       []string{},
		CanAllocateSelf: conf.CanAllocateSelf,
		timeout:         conf.Timeout,
		exclusionRules:  make(map[string][]string),
		mustGetRules:    make(map[string]string),
		Name:            conf.Name,
	}

	if err := a.loadNames(conf); err != nil {
		return nil, err
	}

	if err := a.loadPasswords(conf); err != nil {
		return nil, err
	}

	if err := a.loadRules(conf); err != nil {
		return nil, err
	}

	return a, nil
}

func (a *Allocator) loadNames(conf *Config) error {
	var undupedNames []string
	if conf.Names.File != "" {
		err := utils.ReadFileIntoSlice(conf.Names.File, &undupedNames)
		if err != nil {
			return fmt.Errorf("error while loading names from file: %w", err)
		}
	}
	undupedNames = append(undupedNames, conf.Names.Data...)
	var formattedUndupedNames []string

	for _, name := range undupedNames {
		formattedName := strings.ToLower(strings.TrimSpace(name))
		formattedUndupedNames = append(formattedUndupedNames, formattedName)
	}

	a.names = utils.RemoveDuplicatesFromSlice(formattedUndupedNames)
	return nil
}

func (a *Allocator) loadPasswords(conf *Config) error {
	var undupedPasswords []string
	if conf.Passwords.File != "" {
		err := utils.ReadFileIntoSlice(conf.Passwords.File, &undupedPasswords)
		if err != nil {
			return fmt.Errorf("error while loading passwords from file: %v", err)
		}
	}
	undupedPasswords = append(undupedPasswords, conf.Passwords.Data...)
	a.passwords = utils.RemoveDuplicatesFromSlice(undupedPasswords)
	return nil
}

func (a *Allocator) loadRules(conf *Config) error {
	// Load exclusionRules
	for _, rule := range conf.Rules {
		for _, bannedName := range rule.CannotGet {
			// check if the name is in the list of names
			if !slices.Contains(a.names, strings.ToLower(bannedName)) {
				return fmt.Errorf("name [%s] in exclusion rule is not in the list of names", bannedName)
			}
			if !slices.Contains(a.names, strings.ToLower(rule.Name)) {
				return fmt.Errorf("name [%s] in exclusion rule is not in the list of names", rule.Name)
			}

			a.exclusionRules[strings.ToLower(rule.Name)] = append(a.exclusionRules[strings.ToLower(rule.Name)], strings.ToLower(bannedName))
			if rule.Inverse {
				a.exclusionRules[strings.ToLower(bannedName)] = append(a.exclusionRules[strings.ToLower(bannedName)], strings.ToLower(rule.Name))
			}
		}
	}

	// Load mustGet Rules
	for _, rule := range conf.Rules {
		if rule.MustGet != "" {
			if !slices.Contains(a.names, rule.MustGet) {
				return fmt.Errorf("name [%s] in must get rule is not in the list of names", rule.MustGet)
			}
			a.mustGetRules[strings.ToLower(rule.Name)] = strings.ToLower(rule.MustGet)
		}
	}
	return nil
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

	alloc := NewAllocation(a.names, a.passwords)

	// 1. Handle MustGet rules
	// Keep track of who is available to be a Santa
	availableSantas := make([]*player, len(alloc.players))
	copy(availableSantas, alloc.players)

	for santaName, mustGet := range a.mustGetRules {
		santa := alloc.GetPlayer(santaName)
		santee := alloc.GetPlayer(mustGet)

		if santa == nil || santee == nil {
			return nil, fmt.Errorf("must get rule contains unknown player: %s -> %s", santaName, mustGet)
		}

		if santa.santaFor != nil {
			return nil, fmt.Errorf("player %s is assigned to multiple people", santa.name)
		}
		if santee.santa != nil {
			return nil, fmt.Errorf("player %s is assigned multiple santas", santee.name)
		}

		// Assign
		santa.santaFor = santee
		santee.santa = santa

		// Remove from available santas
		idx := slices.Index(availableSantas, santa)
		if idx == -1 {
			// Should not happen if logic is correct
			return nil, fmt.Errorf("player %s already used as santa", santa.name)
		}
		availableSantas = slices.Delete(availableSantas, idx, idx+1)
	}

	// 2. Solve for the rest using backtracking
	if a.solve(0, alloc.players, availableSantas) {
		a.lastAllocation = *alloc
		return alloc, nil
	}

	return nil, fmt.Errorf("impossible to create allocation, check rules")
}

func (a *Allocator) solve(santeeIndex int, players []*player, availableSantas []*player) bool {
	// Base case: all players have been processed as santees
	if santeeIndex >= len(players) {
		return true
	}

	santee := players[santeeIndex]

	// If this player already has a santa (from MustGet), skip to next
	if santee.santa != nil {
		return a.solve(santeeIndex+1, players, availableSantas)
	}

	// Try to find a santa for this santee
	// Shuffle available santas to ensure randomness
	candidates := make([]*player, len(availableSantas))
	copy(candidates, availableSantas)
	utils.ShuffleSlice(candidates)

	for _, santa := range candidates {
		if a.canAssign(santa, santee) {
			// Assign
			santa.santaFor = santee
			santee.santa = santa

			// Prepare next available santas
			nextAvailable := make([]*player, 0, len(availableSantas)-1)
			for _, p := range availableSantas {
				if p != santa {
					nextAvailable = append(nextAvailable, p)
				}
			}

			// Recurse
			if a.solve(santeeIndex+1, players, nextAvailable) {
				return true
			}

			// Backtrack
			santa.santaFor = nil
			santee.santa = nil
		}
	}

	return false
}

func (a *Allocator) canAssign(santa, santee *player) bool {
	if santa == santee && !a.CanAllocateSelf {
		return false
	}

	// Check exclusion rules
	// exclusionRules maps Santa Name -> List of names they cannot get
	if excluded, ok := a.exclusionRules[strings.ToLower(santa.name)]; ok {
		for _, name := range excluded {
			if strings.EqualFold(name, santee.name) {
				return false
			}
		}
	}

	return true
}

// validateSetup ensures that a give Allocator
// has enough names and passwords to create an
// allocation. Does not check rules.
func (a *Allocator) validateSetup() error {
	minNames := 2
	if a.CanAllocateSelf {
		minNames = 1
	}

	if len(a.names) < minNames {
		return fmt.Errorf("need at least %d names", minNames)
	}

	if len(a.passwords) < minNames {
		return fmt.Errorf("need at least %d passwords", minNames)
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
