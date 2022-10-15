package allocator

import (
	"fmt"
	"time"
)

type Allocation struct {
	aliases     map[string]string // name -> password
	allocations map[string]string // name -> name
	created     time.Time
}

func newAllocation() *Allocation {
	return &Allocation{
		aliases:     make(map[string]string),
		allocations: make(map[string]string),
		created:     time.Now().Local(),
	}
}

func (a *Allocation) PrintNameToPassword() {
	fmt.Println("Printing names to allocated passwords:")
	for name, allocated := range a.allocations {
		fmt.Printf("%s -> %s\n", name, a.aliases[allocated])
	}
}

func (a *Allocation) PrintNameToName() {
	fmt.Println("Printing names to allocated names:")
	for name, allocated := range a.allocations {
		fmt.Printf("%s -> %s\n", name, allocated)
	}
}

func (a *Allocation) PrintAliases() {
	fmt.Println("Printing names aliases:")
	for name, password := range a.aliases {
		fmt.Printf("%s -> %s\n", name, password)
	}
}

func (a *Allocation) String() string {

	res := fmt.Sprintf("Created at %s\n", a.created.Format("01-02-2006 15:04:05"))
	res += "Aliases:\n"
	for name, password := range a.aliases {
		res += fmt.Sprintf("%s -> %s\n", name, password)
	}
	res += "\n"

	res += "Allocations:\n"
	for password, name := range a.allocations {
		res += fmt.Sprintf("%s -> %s\n", password, name)
	}
	return res
}
