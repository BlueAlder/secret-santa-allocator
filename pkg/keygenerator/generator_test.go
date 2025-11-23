package keygenerator

import (
	"testing"
)

func TestGenerateKey(t *testing.T) {
	length := 10
	key := Generate(length)

	if len(key) != length {
		t.Errorf("expected key length %d, got %d", length, len(key))
	}

	for _, char := range key {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z')) {
			t.Errorf("key contains non-alphabetical character: %c", char)
		}
	}
}

func TestGenerateKeyRandomness(t *testing.T) {
	key1 := Generate(10)
	key2 := Generate(10)

	if key1 == key2 {
		t.Errorf("GenerateKey returned identical keys: %s", key1)
	}
}
