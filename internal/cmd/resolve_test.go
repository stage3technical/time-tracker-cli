package cmd

import (
	"testing"
	"time"
)

func TestDefaultWeekStartMonday(t *testing.T) {
	// Tuesday 2026-07-07 → Monday 2026-07-06
	got := defaultWeekStart()
	now := time.Now()
	daysSinceMonday := (int(now.Weekday()) + 6) % 7
	want := now.AddDate(0, 0, -daysSinceMonday).Format("2006-01-02")
	if got != want {
		t.Fatalf("defaultWeekStart() = %q, want %q", got, want)
	}
}

func TestLastWeekStart(t *testing.T) {
	got := lastWeekStart()
	now := time.Now()
	daysSinceMonday := (int(now.Weekday()) + 6) % 7
	want := now.AddDate(0, 0, -daysSinceMonday-7).Format("2006-01-02")
	if got != want {
		t.Fatalf("lastWeekStart() = %q, want %q", got, want)
	}
	thisMon, err := time.Parse("2006-01-02", defaultWeekStart())
	if err != nil {
		t.Fatal(err)
	}
	lastMon, err := time.Parse("2006-01-02", got)
	if err != nil {
		t.Fatal(err)
	}
	if thisMon.Sub(lastMon) != 7*24*time.Hour {
		t.Fatalf("lastWeekStart should be 7 days before this Monday: this=%s last=%s", thisMon, lastMon)
	}
}

func TestWeekStartOrDefault(t *testing.T) {
	if got := weekStartOrDefault("2026-01-05"); got != "2026-01-05" {
		t.Fatalf("explicit week-start = %q", got)
	}
	if got := weekStartOrDefault(""); got != defaultWeekStart() {
		t.Fatalf("empty flag should default")
	}
}

func TestRequireConfirm(t *testing.T) {
	if err := requireConfirm(false, "archive"); err == nil {
		t.Fatal("expected error without --confirm")
	}
	if err := requireConfirm(true, "archive"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
