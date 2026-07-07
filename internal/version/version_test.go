package version

import "testing"

func TestStringDevDefaults(t *testing.T) {
	got := String()
	if got == "" {
		t.Fatal("expected non-empty version string")
	}
}
