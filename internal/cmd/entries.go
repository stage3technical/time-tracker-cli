package cmd

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stage3technical/time-tracker-cli/internal/output"
)

var entriesCmd = &cobra.Command{
	Use:   "entries",
	Short: "Time reporting entries",
}

var (
	entryPersonID   string
	entryEmail      string
	entryWeekStart  string
	entryWorkDate   string
	entryProjectID  string
	entryProjectName string
	entryProjectCode string
	entryRole       string
	entryHours      string
	entryNotes      string
	entryFile       string
	entryConfirm    bool
)

var entriesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List entries for a person (GET /api/v1/time-reporting/entries)",
	RunE:  runEntriesList,
}

var entriesGetCmd = &cobra.Command{
	Use:   "get ENTRY_ID",
	Short: "Get an entry by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := resolveClient(true)
		if err != nil {
			return err
		}
		resp, err := c.Get("/api/v1/time-reporting/entries/"+args[0], nil)
		if err != nil {
			handleAPIError(err)
		}
		return printResponse(resp.StatusCode, resp.Body)
	},
}

var entriesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a time entry (POST /api/v1/time-reporting/entries)",
	RunE:  runEntriesCreate,
}

var entriesUpdateCmd = &cobra.Command{
	Use:   "update ENTRY_ID",
	Short: "Update a time entry (PUT /api/v1/time-reporting/entries/{id})",
	Args:  cobra.ExactArgs(1),
	RunE:  runEntriesUpdate,
}

var entriesDeleteCmd = &cobra.Command{
	Use:   "delete ENTRY_ID",
	Short: "Delete a time entry (DELETE /api/v1/time-reporting/entries/{id})",
	Args:  cobra.ExactArgs(1),
	RunE:  runEntriesDelete,
}

func runEntriesList(cmd *cobra.Command, args []string) error {
	personID, err := resolvePersonID(entryPersonID, entryEmail)
	if err != nil {
		return err
	}
	workDate := strings.TrimSpace(entryWorkDate)
	weekStart := strings.TrimSpace(entryWeekStart)
	if workDate != "" && weekStart != "" {
		return fmt.Errorf("use --work-date or --week-start, not both")
	}
	if workDate == "" && weekStart == "" {
		weekStart = defaultWeekStart()
	}

	q := url.Values{"personId": {personID}}
	if workDate != "" {
		q.Set("workDate", workDate)
	} else {
		q.Set("weekStartDate", weekStart)
	}

	c, err := resolveClient(true)
	if err != nil {
		return err
	}
	resp, err := c.Get("/api/v1/time-reporting/entries", q)
	if err != nil {
		handleAPIError(err)
	}
	if out.Mode == output.ModePretty {
		return out.PrintEntriesList(resp.Body)
	}
	return printResponse(resp.StatusCode, resp.Body)
}

func runEntriesCreate(cmd *cobra.Command, args []string) error {
	personID, err := resolvePersonID(entryPersonID, entryEmail)
	if err != nil {
		return err
	}
	projectID, err := resolveProjectID(entryProjectID, entryProjectName, entryProjectCode)
	if err != nil {
		return err
	}

	fields := map[string]any{
		"personId":  personID,
		"projectId": projectID,
		"workDate":  strings.TrimSpace(entryWorkDate),
		"role":      strings.TrimSpace(entryRole),
		"notes":     strings.TrimSpace(entryNotes),
	}
	if entryWorkDate == "" && entryFile == "" {
		return fmt.Errorf("--work-date is required (or include workDate in --file)")
	}
	if h := strings.TrimSpace(entryHours); h != "" {
		hours, err := strconv.ParseFloat(h, 64)
		if err != nil {
			return fmt.Errorf("invalid --hours: %w", err)
		}
		fields["hours"] = hours
	} else if entryFile == "" {
		return fmt.Errorf("--hours is required (or include hours in --file)")
	}
	body, err := mergePayload(entryFile, fields)
	if err != nil {
		return err
	}

	c, err := resolveClient(true)
	if err != nil {
		return err
	}
	resp, err := c.Do("POST", "/api/v1/time-reporting/entries", nil, body)
	if err != nil {
		handleAPIError(err)
	}
	return printResponse(resp.StatusCode, resp.Body)
}

func runEntriesUpdate(cmd *cobra.Command, args []string) error {
	fields := map[string]any{}
	if cmd.Flags().Changed("role") {
		fields["role"] = strings.TrimSpace(entryRole)
	}
	if cmd.Flags().Changed("hours") {
		h := strings.TrimSpace(entryHours)
		if h == "" {
			return fmt.Errorf("--hours value is required when flag is set")
		}
		hours, err := strconv.ParseFloat(h, 64)
		if err != nil {
			return fmt.Errorf("invalid --hours: %w", err)
		}
		fields["hours"] = hours
	}
	if cmd.Flags().Changed("notes") {
		fields["notes"] = strings.TrimSpace(entryNotes)
	}
	body, err := mergePayload(entryFile, fields)
	if err != nil {
		return err
	}

	c, err := resolveClient(true)
	if err != nil {
		return err
	}
	resp, err := c.Do("PUT", "/api/v1/time-reporting/entries/"+args[0], nil, body)
	if err != nil {
		handleAPIError(err)
	}
	return printResponse(resp.StatusCode, resp.Body)
}

func runEntriesDelete(cmd *cobra.Command, args []string) error {
	if err := requireConfirm(entryConfirm, "delete entry"); err != nil {
		return err
	}
	c, err := resolveClient(true)
	if err != nil {
		return err
	}
	resp, err := c.Do("DELETE", "/api/v1/time-reporting/entries/"+args[0], nil, nil)
	if err != nil {
		handleAPIError(err)
	}
	return printResponse(resp.StatusCode, resp.Body)
}

func init() {
	register(rootCmd, entriesCmd, CapRead)
	register(entriesCmd, entriesListCmd, CapRead)
	register(entriesCmd, entriesGetCmd, CapRead)
	register(entriesCmd, entriesCreateCmd, CapWrite)
	register(entriesCmd, entriesUpdateCmd, CapWrite)
	register(entriesCmd, entriesDeleteCmd, CapWrite)

	for _, cmd := range []*cobra.Command{entriesListCmd, entriesCreateCmd} {
		cmd.Flags().StringVar(&entryPersonID, "person-id", "", "person UUID")
		cmd.Flags().StringVar(&entryEmail, "email", "", "person email (looks up ID)")
	}
	entriesListCmd.Flags().StringVar(&entryWeekStart, "week-start", "", "Monday week start (default: this Monday)")
	entriesListCmd.Flags().StringVar(&entryWorkDate, "work-date", "", "single day (YYYY-MM-DD) instead of week")

	entriesCreateCmd.Flags().StringVar(&entryProjectID, "project-id", "", "project UUID")
	entriesCreateCmd.Flags().StringVar(&entryProjectName, "project-name", "", "project canonical name lookup")
	entriesCreateCmd.Flags().StringVar(&entryProjectCode, "project-code", "", "exact project name lookup (canonicalName)")
	entriesCreateCmd.Flags().StringVar(&entryWorkDate, "work-date", "", "work date (YYYY-MM-DD)")
	entriesCreateCmd.Flags().StringVar(&entryRole, "role", "", "role on the entry")
	entriesCreateCmd.Flags().StringVar(&entryHours, "hours", "", "hours (decimal)")
	entriesCreateCmd.Flags().StringVar(&entryNotes, "notes", "", "optional notes")
	entriesCreateCmd.Flags().StringVar(&entryFile, "file", "", "JSON payload file (merged with flags)")

	entriesUpdateCmd.Flags().StringVar(&entryRole, "role", "", "role on the entry")
	entriesUpdateCmd.Flags().StringVar(&entryHours, "hours", "", "hours (decimal)")
	entriesUpdateCmd.Flags().StringVar(&entryNotes, "notes", "", "optional notes")
	entriesUpdateCmd.Flags().StringVar(&entryFile, "file", "", "JSON payload file (merged with flags)")

	entriesDeleteCmd.Flags().BoolVar(&entryConfirm, "confirm", false, "confirm destructive delete")
}
