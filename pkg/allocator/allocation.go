package allocator

import (
	"fmt"
	"time"
)

type Allocation struct {
	aliases     map[string]string // name -> password
	allocations map[string]string // name -> password
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
	for name, password := range a.allocations {
		fmt.Printf("Name: %s   Password: %s\n", name, password)
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
