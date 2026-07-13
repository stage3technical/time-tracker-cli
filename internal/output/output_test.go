package output

import (
	"bytes"
	"strings"
	"testing"
)

func sampleRosterJSON() []byte {
	return []byte(`{
  "weekStartDate": "2026-07-06",
  "weekLock": { "status": "open" },
  "persons": [
    {
      "personId": "1",
      "name": "Corinna Example",
      "email": "corinna@example.com",
      "entryCount": 5,
      "totalHours": "40.0",
      "submission": { "status": "submitted", "submittedAt": "2026-07-11T18:00:00Z" }
    },
    {
      "personId": "2",
      "name": "Leon Example",
      "email": "leon@example.com",
      "entryCount": 1,
      "totalHours": "8.0",
      "submission": { "status": "draft" }
    }
  ]
}`)
}

func TestPrintWeekRosterPrettyAll(t *testing.T) {
	var buf bytes.Buffer
	w := &Writer{Mode: ModePretty, Stdout: &buf}
	if err := w.PrintWeekRoster(sampleRosterJSON(), "all"); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	for _, want := range []string{
		"Week 2026-07-06  lock=open",
		"Corinna Example",
		"Leon Example",
		"submitted",
		"draft",
		"Submitted: 1 / 2",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("pretty output missing %q\n%s", want, got)
		}
	}
}

func TestPrintWeekRosterPrettySubmittedFilter(t *testing.T) {
	var buf bytes.Buffer
	w := &Writer{Mode: ModePretty, Stdout: &buf}
	if err := w.PrintWeekRoster(sampleRosterJSON(), "submitted"); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, "Corinna Example") {
		t.Fatalf("expected submitted row:\n%s", got)
	}
	if strings.Contains(got, "Leon Example") {
		t.Fatalf("draft row should be filtered out:\n%s", got)
	}
	if !strings.Contains(got, "Submitted: 1 / 1") {
		t.Fatalf("expected filter-aware summary:\n%s", got)
	}
}

func TestPrintWeekRosterJSONFilter(t *testing.T) {
	var buf bytes.Buffer
	w := &Writer{Mode: ModeJSON, Stdout: &buf}
	if err := w.PrintWeekRoster(sampleRosterJSON(), "draft"); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, "Leon Example") {
		t.Fatalf("expected draft person in JSON:\n%s", got)
	}
	if strings.Contains(got, "Corinna Example") {
		t.Fatalf("submitted person should be filtered from JSON:\n%s", got)
	}
}
