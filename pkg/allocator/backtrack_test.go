package allocator

import (
	"testing"
)

func TestBacktrackingConstraints(t *testing.T) {
	// A cannot get B
	// B cannot get C
	// C cannot get A
	// Should solve to A->C, B->A, C->B
	a := New()
	a.names = []string{"A", "B", "C"}
	a.passwords = []string{"1", "2", "3"}
	a.exclusionRules["a"] = []string{"b"}
	a.exclusionRules["b"] = []string{"c"}
	a.exclusionRules["c"] = []string{"a"}

	alloc, err := a.Allocate()
	if err != nil {
		t.Fatalf("failed to allocate: %v", err)
	}

	check := func(santa, santee string) {
		s := alloc.GetPlayer(santa)
		if s.SantaFor.Name != santee {
			t.Errorf("expected %s -> %s, got %s -> %s", santa, santee, santa, s.SantaFor.Name)
		}
	}

	check("A", "C")
	check("B", "A")
	check("C", "B")
}

func TestImpossibleConstraint(t *testing.T) {
	// A cannot get B, C
	// A needs to get someone.
	// If A cannot get anyone, it should fail.
	a := New()
	a.names = []string{"A", "B", "C"}
	a.passwords = []string{"1", "2", "3"}
	a.exclusionRules["a"] = []string{"b", "c"}
	// A cannot get itself (default)
	// So A has no options.

	_, err := a.Allocate()
	if err == nil {
		t.Fatal("should have failed")
	}
}

func TestCanAllocateSelf(t *testing.T) {
	a := New()
	a.names = []string{"A"}
	a.passwords = []string{"1"}
	a.CanAllocateSelf = true

	alloc, err := a.Allocate()
	if err != nil {
		t.Fatalf("failed to allocate self: %v", err)
	}

	if alloc.GetPlayer("A").SantaFor.Name != "A" {
		t.Errorf("expected A -> A, got A -> %s", alloc.GetPlayer("A").SantaFor.Name)
	}
}
