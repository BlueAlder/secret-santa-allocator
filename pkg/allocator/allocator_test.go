package allocator

import (
	"reflect"
	"slices"
	"testing"
)

func createAllocator() *Allocator {
	var yaml = `names: 
    data: ["sam", "tom", "jim", "grace", "bill"] # array of names that is unioned with the above file
passwords: 
  data: ["password1", "password2", "password3", "password4", "password5"]
rules:
  - name: sam
    cannotGet: ["tom", "bill"]
  - name: tom
    cannotGet: ["grace"]
    inverse: true`
	c, err := LoadConfigFromYaml([]byte(yaml))

	if err != nil {
		panic(err)
	}

	a, _ := NewFromConfig(c)
	return a
}

func TestFailOn1Name(t *testing.T) {
	names := []string{"sam"}
	passwords := []string{"1", "2", "3", "4", "5"}
	a := createAllocator()
	a.names = names
	a.passwords = passwords
	if _, err := a.Allocate(); err == nil {
		t.Fatalf("should fail on 1 name")
	}
}

func TestAllocateNoDuplicateAliases(t *testing.T) {
	allocator := createAllocator()
	allocation, err := allocator.Allocate()

	if err != nil {
		t.Errorf("Unable to Allocate got: %v", err)
	}

	seenNames := make(map[string]bool)
	seenPasswords := make(map[string]bool)
	for _, player := range allocation.players {
		if seenNames[player.name] {
			t.Errorf("Name has appeared twice in alias allocation: %s", player.name)
		}
		if seenPasswords[player.alias] {
			t.Errorf("Password has appeared twice in alias allocation: %s", player.alias)
		}
	}
}

func TestAllocationDoesNotAssignSelf(t *testing.T) {
	a := createAllocator()
	allocation, _ := a.Allocate()

	for _, player := range allocation.players {
		assigned := player.santaFor
		if player == assigned {
			t.Fatalf("found name with self assigned: %s", player.name)
		}
	}
}

func TestEveryNameIsInThePlayerList(t *testing.T) {
	a := createAllocator()
	allocation, _ := a.Allocate()
	// check alias map
	for _, name := range a.names {
		idx := slices.IndexFunc(allocation.players, func(p *player) bool { return p.name == name })
		if idx == -1 {
			t.Logf("Unable to find name in list of player: %s", name)
			t.Fatal(allocation)
		}
	}

	// Check allocation map

}

func TestAllocateNoDuplicateAllocations(t *testing.T) {
	allocator := createAllocator()
	allocation, err := allocator.Allocate()

	if err != nil {
		t.Errorf("Unable to Allocate got: %v", err)
	}

	seenNames := make(map[string]bool)
	seenPasswords := make(map[string]bool)
	for _, player := range allocation.players {
		if seenNames[player.name] {
			t.Errorf("Name has appeared twice in allocation: %s", player.name)
		}
		seenNames[player.name] = true
		if seenPasswords[player.alias] {
			t.Errorf("Password has appeared twice in allocation: %s", player.alias)
		}
		seenPasswords[player.alias] = true
	}
}

func TestErrorOnLessPasswordsThanNames(t *testing.T) {
	names := []string{"sam", "john", "billiam", "john4", "john3", "extra name"}
	passwords := []string{"1", "2", "3", "4", "5"}

	allocator := createAllocator()
	allocator.names = names
	allocator.passwords = passwords

	_, err := allocator.Allocate()

	if err == nil {
		t.Error("should have errored on more names than passwords")
	}
}

func TestLoadingRules(t *testing.T) {
	allocator := createAllocator()

	want := make(map[string][]string)
	want["sam"] = []string{"tom", "bill"}
	want["tom"] = []string{"grace"}
	want["grace"] = []string{"tom"}

	if len(allocator.exclusionRules) != 3 {
		t.Fatalf("expected 3 rules, got %d", len(allocator.exclusionRules))
	}

	testNames := []string{"sam", "tom", "grace"}

	for _, name := range testNames {
		if !reflect.DeepEqual(allocator.exclusionRules[name], want[name]) {
			t.Fatalf("expected %v, got %v", want[name], allocator.exclusionRules[name])
		}
	}

}

// tests setting multiple mustGet rules for a single name
func TestMultipleMustGetRules(t *testing.T) {
	a := New()
	a.names = []string{"sam", "tom", "jim", "grace", "bill"}
	a.passwords = []string{"password1", "password2", "password3", "password4", "password5"}
	a.mustGetRules["sam"] = "jim"
	a.mustGetRules["grace"] = "jim"

	_, err := a.Allocate()
	if err == nil {
		t.Fatalf("should fail on multiple must get rules")
	}
}

// tests setting a mustGet rule should always assign the same santa
func TestMustGetRule(t *testing.T) {
	a := New()
	a.names = []string{"sam", "tom", "jim", "grace", "bill"}
	a.passwords = []string{"password1", "password2", "password3", "password4", "password5"}
	a.mustGetRules["sam"] = "jim"

	for i := 0; i < 100; i++ {
		allocation, err := a.Allocate()
		if err != nil {
			t.Fatalf("should not fail on must get rule")
		}

		if allocation.GetPlayer("sam").santaFor != allocation.GetPlayer("jim") || allocation.GetPlayer("jim").santa != allocation.GetPlayer("sam") {
			t.Fatalf("must get rule not followed")
		}
	}
}

// test must never get

func TestMustNeverGetRule(t *testing.T) {
	a := New()
	a.names = []string{"sam", "tom", "jim", "grace", "bill"}
	a.passwords = []string{"password1", "password2", "password3", "password4", "password5"}
	a.exclusionRules["sam"] = []string{"tom"}

	for i := 0; i < 100; i++ {
		allocation, err := a.Allocate()
		if err != nil {
			t.Fatalf("should not fail on must never get rule")
		}

		if allocation.GetPlayer("sam").santaFor == allocation.GetPlayer("tom") || allocation.GetPlayer("tom").santa == allocation.GetPlayer("sam") {
			t.Fatalf("must never get rule not followed")
		}
	}
}
