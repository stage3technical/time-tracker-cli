package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
)

// Mode is json or pretty.
type Mode string

const (
	ModeJSON   Mode = "json"
	ModePretty Mode = "pretty"
)

// Writer formats API responses for stdout.
type Writer struct {
	Mode   Mode
	Quiet  bool
	Stdout io.Writer
	Stderr io.Writer
}

// NewWriter creates a Writer with sensible defaults.
func NewWriter(mode Mode, quiet bool) *Writer {
	if mode == "" {
		mode = defaultMode()
	}
	return &Writer{
		Mode:   mode,
		Quiet:  quiet,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
}

func defaultMode() Mode {
	if stat, err := os.Stdout.Stat(); err == nil && (stat.Mode()&os.ModeCharDevice) == 0 {
		return ModeJSON
	}
	return ModePretty
}

// PrintJSON writes raw JSON, optionally indented in pretty mode.
func (w *Writer) PrintJSON(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	if w.Mode == ModeJSON {
		_, err := w.Stdout.Write(data)
		if err == nil {
			_, err = io.WriteString(w.Stdout, "\n")
		}
		return err
	}
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		_, err = w.Stdout.Write(data)
		return err
	}
	enc := json.NewEncoder(w.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// PrintPersonsList renders a persons list in pretty mode.
func (w *Writer) PrintPersonsList(data []byte) error {
	if w.Mode == ModeJSON {
		return w.PrintJSON(data)
	}
	var persons []map[string]any
	if err := json.Unmarshal(data, &persons); err != nil {
		return w.PrintJSON(data)
	}
	tw := tabwriter.NewWriter(w.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "ID\tNAME\tEMAIL\tROLE\tTEAM")
	for _, p := range persons {
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			strVal(p, "id"),
			strVal(p, "name"),
			strVal(p, "email"),
			strVal(p, "primaryRole"),
			strVal(p, "team"),
		)
	}
	return tw.Flush()
}

// PrintProjectsList renders a projects list in pretty mode.
func (w *Writer) PrintProjectsList(data []byte) error {
	if w.Mode == ModeJSON {
		return w.PrintJSON(data)
	}
	var projects []map[string]any
	if err := json.Unmarshal(data, &projects); err != nil {
		return w.PrintJSON(data)
	}
	tw := tabwriter.NewWriter(w.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "ID\tNAME\tBILL_TYPE\tSTATUS")
	for _, p := range projects {
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
			strVal(p, "id"),
			strVal(p, "canonicalName"),
			strVal(p, "billType"),
			strVal(p, "status"),
		)
	}
	return tw.Flush()
}

// PrintEntriesList renders time entries in pretty mode.
func (w *Writer) PrintEntriesList(data []byte) error {
	if w.Mode == ModeJSON {
		return w.PrintJSON(data)
	}
	var entries []map[string]any
	if err := json.Unmarshal(data, &entries); err != nil {
		return w.PrintJSON(data)
	}
	tw := tabwriter.NewWriter(w.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "ID\tWORK_DATE\tPROJECT\tROLE\tHOURS\tSTATUS")
	for _, e := range entries {
		project := strVal(e, "projectCanonicalName")
		if project == "" {
			project = strVal(e, "projectId")
		}
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
			strVal(e, "id"),
			strVal(e, "workDate"),
			project,
			strVal(e, "role"),
			strVal(e, "hours"),
			strVal(e, "status"),
		)
	}
	return tw.Flush()
}

// PrintCompanyRolesList renders company roles in pretty mode.
func (w *Writer) PrintCompanyRolesList(data []byte) error {
	if w.Mode == ModeJSON {
		return w.PrintJSON(data)
	}
	var roles []map[string]any
	if err := json.Unmarshal(data, &roles); err != nil {
		return w.PrintJSON(data)
	}
	tw := tabwriter.NewWriter(w.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "ID\tNAME\tDESCRIPTION")
	for _, r := range roles {
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\n",
			strVal(r, "id"),
			strVal(r, "name"),
			strVal(r, "description"),
		)
	}
	return tw.Flush()
}

// PrintTimesheetWeeksList renders timesheet week summaries in pretty mode.
func (w *Writer) PrintTimesheetWeeksList(data []byte) error {
	if w.Mode == ModeJSON {
		return w.PrintJSON(data)
	}
	var weeks []map[string]any
	if err := json.Unmarshal(data, &weeks); err != nil {
		return w.PrintJSON(data)
	}
	tw := tabwriter.NewWriter(w.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "WEEK_START\tENTRIES\tHOURS\tSUBMISSION\tWEEK_LOCK")
	for _, week := range weeks {
		submission := nestedStr(week, "submission", "status")
		lock := nestedStr(week, "weekLock", "status")
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			strVal(week, "weekStartDate"),
			strVal(week, "entryCount"),
			strVal(week, "totalHours"),
			submission,
			lock,
		)
	}
	return tw.Flush()
}

// PrintWeekRoster renders GET /timesheets/weeks/{date} with optional status filter.
// statusFilter is "all", "submitted", or "draft".
func (w *Writer) PrintWeekRoster(data []byte, statusFilter string) error {
	var roster map[string]any
	if err := json.Unmarshal(data, &roster); err != nil {
		return w.PrintJSON(data)
	}
	persons, _ := roster["persons"].([]any)
	filtered := make([]any, 0, len(persons))
	submittedShown := 0
	for _, raw := range persons {
		p, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		status := nestedStr(p, "submission", "status")
		if status == "" {
			status = "draft"
		}
		if statusFilter == "submitted" || statusFilter == "draft" {
			if !strings.EqualFold(status, statusFilter) {
				continue
			}
		}
		if strings.EqualFold(status, "submitted") {
			submittedShown++
		}
		filtered = append(filtered, p)
	}
	roster["persons"] = filtered

	if w.Mode == ModeJSON {
		enc, err := json.Marshal(roster)
		if err != nil {
			return err
		}
		return w.PrintJSON(enc)
	}

	weekStart := strVal(roster, "weekStartDate")
	lock := nestedStr(roster, "weekLock", "status")
	if lock == "" {
		lock = "open"
	}
	_, _ = fmt.Fprintf(w.Stdout, "Week %s  lock=%s\n\n", weekStart, lock)

	tw := tabwriter.NewWriter(w.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "NAME\tEMAIL\tSTATUS\tUNLOCKED\tHOURS\tENTRIES")
	for _, raw := range filtered {
		p := raw.(map[string]any)
		status := nestedStr(p, "submission", "status")
		if status == "" {
			status = "draft"
		}
		unlocked := "-"
		if boolVal(p, "unlockException") {
			unlocked = "yes"
		}
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
			strVal(p, "name"),
			strVal(p, "email"),
			status,
			unlocked,
			strVal(p, "totalHours"),
			strVal(p, "entryCount"),
		)
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(w.Stdout, "\nSubmitted: %d / %d\n", submittedShown, len(filtered))
	return nil
}

func nestedStr(m map[string]any, outer, inner string) string {
	obj, ok := m[outer].(map[string]any)
	if !ok {
		return ""
	}
	return strVal(obj, inner)
}

// PrintError writes an error message to stderr unless quiet.
func (w *Writer) PrintError(msg string) {
	if w.Quiet {
		return
	}
	_, _ = io.WriteString(w.Stderr, msg+"\n")
}

func strVal(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	return fmt.Sprint(v)
}

func boolVal(m map[string]any, key string) bool {
	v, ok := m[key]
	if !ok || v == nil {
		return false
	}
	b, ok := v.(bool)
	return ok && b
}

// ParseOutputFlag validates --output value.
func ParseOutputFlag(s string) (Mode, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "pretty":
		return ModePretty, nil
	case "json":
		return ModeJSON, nil
	default:
		return "", fmt.Errorf("invalid output %q (use json or pretty)", s)
	}
}
