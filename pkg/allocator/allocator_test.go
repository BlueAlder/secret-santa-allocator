package allocator

import (
	"testing"
)

func createAllocator() *Allocator {
	var yaml = `names: 
    data: ["sam", "tom", "jim", "grace", "bill"] # array of names that is unioned with the above file
passwords: 
  data: ["password1", "password2", "password3", "password4", "password5"]`
	c, _ := LoadConfigFromYaml([]byte(yaml))
	a, _ := New(c)
	return a
}

func TestFailOn1Name(t *testing.T) {
	names := map[string]struct{}{"sam": {}}
	passwords := map[string]struct{}{"1": {}, "2": {}, "3": {}, "4": {}, "5": {}}
	a := createAllocator()
	a.Names = names
	a.Passwords = passwords
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
	for name, password := range allocation.Aliases {
		if seenNames[name] {
			t.Errorf("Name has appeared twice in alias allocation: %s", name)
		}
		if seenPasswords[password] {
			t.Errorf("Password has appeared twice in alias allocation: %s", name)
		}
	}
}

func TestAllocationDoesNotAssignSelf(t *testing.T) {
	a := createAllocator()
	allocation, _ := a.Allocate()

	for name, password := range allocation.Allocations {
		assigned := allocation.Aliases[password]
		if name == assigned {
			t.Fatalf("found name with self assigned: %s", name)
		}
	}
}

func TestEveryNameIsInEachList(t *testing.T) {
	a := createAllocator()
	allocation, _ := a.Allocate()
	// check alias map
	for name := range a.Names {
		_, exists := allocation.Aliases[name]
		if !exists {
			t.Logf("Unable to find name in alias: %s", name)
			t.Fatal(allocation)
		}
		_, exists = allocation.Allocations[name]
		if !exists {
			t.Logf("Unable to find name in allocations: %s", name)
			t.Fatal(allocation)
		}
	}

	// Check allocation map

}

// Don't need this anymore
// func TestEveryPasswordIsInEachList(t *testing.T) {
// 	a := createAllocator()
// 	allocation, _ := a.Allocate()

// 	for password := range a.Passwords {
// 		// check alias passwords
// 		exists := false
// 		for _, allocationPassword := range allocation.allocations {
// 			if password == allocationPassword {
// 				exists = true
// 				break
// 			}
// 		}
// 		if !exists {

// 			t.Logf("Unable to find password in allocations: %s", password)
// 			t.Fatal(allocation)
// 		}

// 		for password := range a.Passwords {
// 			// check alias passwords
// 			exists := false
// 			for _, aliasPassword := range allocation.aliases {
// 				if password == aliasPassword {
// 					exists = true
// 					break
// 				}
// 			}
// 			if !exists {

// 				t.Logf("Unable to find password in aliases: %s", password)
// 				t.Fatal(allocation)
// 			}
// 		}
// 	}
// }

func TestAllocateNoDuplicateAllocations(t *testing.T) {
	allocator := createAllocator()
	allocation, err := allocator.Allocate()

	if err != nil {
		t.Errorf("Unable to Allocate got: %v", err)
	}

	seenNames := make(map[string]bool)
	seenPasswords := make(map[string]bool)
	for name, password := range allocation.Allocations {
		if seenNames[name] {
			t.Errorf("Name has appeared twice in allocation: %s", name)
		}
		if seenPasswords[password] {
			t.Errorf("Password has appeared twice in allocation: %s", name)
		}
	}
}

func TestErrorOnLessPasswordsThanNames(t *testing.T) {
	names := map[string]struct{}{"sam": {}, "john": {}, "billiam": {}, "john4": {}, "john3": {}, "extra name": {}}
	passwords := map[string]struct{}{"1": {}, "2": {}, "3": {}, "4": {}, "5": {}}

	// const yaml := ``
	allocator := createAllocator()
	allocator.Names = names
	allocator.Passwords = passwords

	_, err := allocator.Allocate()

	if err == nil {
		t.Error("should have errored on more names than passwords")
	}
}
