package shipinternal

import "testing"

func TestNewRejectsUnknownProvider(t *testing.T) {
	provider, err := New("nope")
	if err == nil {
		t.Fatal("New returned nil error for unknown provider")
	}
	if provider != nil {
		t.Fatal("New returned provider for unknown provider")
	}
}

func TestFirstNonEmpty(t *testing.T) {
	if got := firstNonEmpty("", "fallback"); got != "fallback" {
		t.Fatalf("firstNonEmpty returned %q, want fallback", got)
	}
	if got := firstNonEmpty("value", "fallback"); got != "value" {
		t.Fatalf("firstNonEmpty returned %q, want value", got)
	}
}
