package cmd

import "testing"

func TestRequireBackportConfirm(t *testing.T) {
	if err := requireBackportConfirm("app-dev-main"); err != nil {
		t.Fatalf("expected ok for app-dev-main, got %v", err)
	}
	if err := requireBackportConfirm(""); err == nil {
		t.Fatal("expected error for empty confirm")
	}
	if err := requireBackportConfirm("app-prod-main"); err == nil {
		t.Fatal("expected error for wrong confirm token")
	}
}
