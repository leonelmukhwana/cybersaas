package terminal

import (
	"regexp"
	"testing"
)

func TestGenerateLicenceKey(t *testing.T) {
	pattern := regexp.MustCompile(`^[A-HJ-NP-Z2-9]{3}-[A-HJ-NP-Z2-9]{3}-[A-HJ-NP-Z2-9]{3}$`)

	for i := 0; i < 100; i++ {
		key, err := generateLicenceKey()
		if err != nil {
			t.Fatalf("generateLicenceKey() returned error: %v", err)
		}

		if !pattern.MatchString(key) {
			t.Fatalf("generated licence key %q does not match expected format", key)
		}

		if len(key) != 11 {
			t.Fatalf("generated licence key %q has length %d, want 11", key, len(key))
		}
	}
}

func TestGenerateLicenceKeyProducesDifferentKeys(t *testing.T) {
	first, err := generateLicenceKey()
	if err != nil {
		t.Fatalf("first generateLicenceKey() returned error: %v", err)
	}

	second, err := generateLicenceKey()
	if err != nil {
		t.Fatalf("second generateLicenceKey() returned error: %v", err)
	}

	if first == second {
		t.Fatalf("generated the same licence key twice: %q", first)
	}
}
