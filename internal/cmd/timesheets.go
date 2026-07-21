package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stage3technical/time-tracker-cli/internal/output"
)

var timesheetsCmd = &cobra.Command{
	Use:   "timesheets",
	Short: "Timesheet workflow (Advanced Workflow API)",
}

var (
	timesheetWeekStart    string
	timesheetBefore       string
	timesheetAfter        string
	timesheetEmail        string
	timesheetPersonID     string
	timesheetConfirm      bool
	timesheetRosterStatus string
	timesheetRosterWeek   string
)

var timesheetsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all weeks for a person (GET /api/v1/timesheets/{personId}/weeks)",
	RunE:  runTimesheetList,
}

var timesheetsWeekCmd = &cobra.Command{
	Use:   "week",
	Short: "Week roster for all persons (GET /api/v1/timesheets/weeks/{weekStartDate})",
	Long: `Show submission status for every active person for one week.

Defaults to this week's Monday. Use --status to filter rows (submitted, draft, or all).`,
	Example: `  tt timesheets week
  tt timesheets week --week-start 2026-07-06
  tt timesheets week --status submitted
  tt --output json timesheets week --status submitted`,
	RunE: runTimesheetWeek,
}

var timesheetsLastWeekCmd = &cobra.Command{
	Use:   "lastweek",
	Short: "Week roster for the previous Monday",
	Long: `Show submission status for every active person for last week (previous Monday).

Same output as timesheets week. Use --status to filter rows.`,
	Example: `  tt timesheets lastweek
  tt timesheets lastweek --status submitted
  tt --output json timesheets lastweek`,
	RunE: runTimesheetLastWeek,
}

var timesheetsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a person's week timesheet (GET /api/v1/timesheets/{personId})",
	RunE:  runTimesheetGet,
}

var timesheetsSubmitCmd = &cobra.Command{
	Use:   "submit",
	Short: "Submit a week (POST /api/v1/timesheets/submit)",
	RunE:  runTimesheetSubmit,
}


var timesheetsRejectCmd = &cobra.Command{
	Use:   "reject",
	Short: "Reject a submitted week (POST /api/v1/timesheets/{personId}/reject)",
	RunE:  runTimesheetReject,
}

var timesheetsUnlockCmd = &cobra.Command{
	Use:   "unlock",
	Short: "Admin unlock one person's timesheet for one week (POST /api/v1/timesheets/{personId}/unlock)",
	Long: `Unlock sets that person+week submission and entries back to draft.
They must submit again. Does not use global week lock (retired).
See docs/SUBMISSION_UNLOCK_MODEL.md.`,
	Example: `  tt timesheets unlock --email david.mead@blvdinteractive.com
  tt timesheets unlock --email david.mead@blvdinteractive.com --week-start 2026-07-06`,
	RunE: runTimesheetUnlock,
}




var timesheetsPurgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Admin purge: delete entries and submission for week(s) (POST /api/v1/timesheets/{personId}/purge)",
	Long: `Deletes all time entries and the WeekSubmission for the person/week.
Requires --confirm. Use --week-start for one week or --before for all prior weeks.`,
	RunE: runTimesheetPurge,
}

func runTimesheetWeek(cmd *cobra.Command, args []string) error {
	weekStart := weekStartOrDefault(timesheetRosterWeek)
	return runWeekRoster(weekStart)
}

func runTimesheetLastWeek(cmd *cobra.Command, args []string) error {
	return runWeekRoster(lastWeekStart())
}

func runWeekRoster(weekStart string) error {
	statusFilter := strings.ToLower(strings.TrimSpace(timesheetRosterStatus))
	if statusFilter == "" {
		statusFilter = "all"
	}
	switch statusFilter {
	case "all", "submitted", "draft":
	default:
		return fmt.Errorf("--status must be submitted, draft, or all")
	}

	c, err := resolveClient(true)
	if err != nil {
		return err
	}
	path := "/api/v1/timesheets/weeks/" + weekStart
	resp, err := c.Get(path, nil)
	if err != nil {
		handleAPIError(err)
	}
	if out.Mode == output.ModePretty {
		return out.PrintWeekRoster(resp.Body, statusFilter)
	}
	if statusFilter != "all" {
		return out.PrintWeekRoster(resp.Body, statusFilter)
	}
	return printResponse(resp.StatusCode, resp.Body)
}

func runTimesheetList(cmd *cobra.Command, args []string) error {
	personID, err := resolvePersonID(timesheetPersonID, timesheetEmail)
	if err != nil {
		return err
	}
	q := url.Values{}
	if strings.TrimSpace(timesheetBefore) != "" {
		q.Set("before", strings.TrimSpace(timesheetBefore))
	}
	if strings.TrimSpace(timesheetAfter) != "" {
		q.Set("after", strings.TrimSpace(timesheetAfter))
	}
	c, err := resolveClient(true)
	if err != nil {
		return err
	}
	resp, err := c.Get("/api/v1/timesheets/"+personID+"/weeks", q)
	if err != nil {
		handleAPIError(err)
	}
	if out.Mode == output.ModePretty {
		return out.PrintTimesheetWeeksList(resp.Body)
	}
	return printResponse(resp.StatusCode, resp.Body)
}

func runTimesheetPurge(cmd *cobra.Command, args []string) error {
	if err := requireConfirm(timesheetConfirm, "purge timesheet"); err != nil {
		return err
	}
	personID, err := resolvePersonID(timesheetPersonID, timesheetEmail)
	if err != nil {
		return err
	}
	weekStart := strings.TrimSpace(timesheetWeekStart)
	before := strings.TrimSpace(timesheetBefore)
	if weekStart != "" && before != "" {
		return fmt.Errorf("use --week-start or --before, not both")
	}
	if weekStart == "" && before == "" {
		return fmt.Errorf("one of --week-start or --before is required")
	}
	q := url.Values{}
	if weekStart != "" {
		q.Set("weekStartDate", weekStart)
	} else {
		q.Set("before", before)
	}
	c, err := resolveClient(true)
	if err != nil {
		return err
	}
	resp, err := c.Do("POST", "/api/v1/timesheets/"+personID+"/purge", q, nil)
	if err != nil {
		handleAPIError(err)
	}
	return printResponse(resp.StatusCode, resp.Body)
}

func runTimesheetGet(cmd *cobra.Command, args []string) error {
	personID, err := resolvePersonID(timesheetPersonID, timesheetEmail)
	if err != nil {
		return err
	}
	weekStart := weekStartOrDefault(timesheetWeekStart)
	c, err := resolveClient(true)
	if err != nil {
		return err
	}
	q := url.Values{"weekStartDate": {weekStart}}
	resp, err := c.Get("/api/v1/timesheets/"+personID, q)
	if err != nil {
		handleAPIError(err)
	}
	return printResponse(resp.StatusCode, resp.Body)
}

func runTimesheetSubmit(cmd *cobra.Command, args []string) error {
	personID, err := resolvePersonID(timesheetPersonID, timesheetEmail)
	if err != nil {
		return err
	}
	weekStart := weekStartOrDefault(timesheetWeekStart)
	body, err := json.Marshal(map[string]string{
		"personId":      personID,
		"weekStartDate": weekStart,
	})
	if err != nil {
		return err
	}
	c, err := resolveClient(true)
	if err != nil {
		return err
	}
	resp, err := c.Do("POST", "/api/v1/timesheets/submit", nil, body)
	if err != nil {
		handleAPIError(err)
	}
	return printResponse(resp.StatusCode, resp.Body)
}


func runTimesheetReject(cmd *cobra.Command, args []string) error {
	return timesheetPersonAction("POST", "reject")
}

func runTimesheetUnlock(cmd *cobra.Command, args []string) error {
	return timesheetPersonAction("POST", "unlock")
}




func timesheetPersonAction(method, action string) error {
	personID, err := resolvePersonID(timesheetPersonID, timesheetEmail)
	if err != nil {
		return err
	}
	weekStart := weekStartOrDefault(timesheetWeekStart)
	c, err := resolveClient(true)
	if err != nil {
		return err
	}
	q := url.Values{"weekStartDate": {weekStart}}
	path := fmt.Sprintf("/api/v1/timesheets/%s/%s", personID, action)
	resp, err := c.Do(method, path, q, nil)
	if err != nil {
		handleAPIError(err)
	}
	return printResponse(resp.StatusCode, resp.Body)
}

func init() {
	register(rootCmd, timesheetsCmd, CapRead)
	register(timesheetsCmd, timesheetsListCmd, CapRead)
	register(timesheetsCmd, timesheetsWeekCmd, CapRead)
	register(timesheetsCmd, timesheetsLastWeekCmd, CapRead)
	register(timesheetsCmd, timesheetsGetCmd, CapRead)
	register(timesheetsCmd, timesheetsSubmitCmd, CapWrite)
	register(timesheetsCmd, timesheetsRejectCmd, CapWrite)
	register(timesheetsCmd, timesheetsUnlockCmd, CapWrite)
	register(timesheetsCmd, timesheetsPurgeCmd, CapWrite)

	for _, cmd := range []*cobra.Command{
		timesheetsListCmd,
		timesheetsGetCmd,
		timesheetsSubmitCmd,
		timesheetsRejectCmd,
		timesheetsUnlockCmd,
		timesheetsPurgeCmd,
	} {
		cmd.Flags().StringVar(&timesheetPersonID, "person-id", "", "person UUID")
		cmd.Flags().StringVar(&timesheetEmail, "email", "", "person email (looks up ID via GET /persons)")
	}
	timesheetsListCmd.Flags().StringVar(&timesheetBefore, "before", "", "exclude weeks starting on/after this Monday (YYYY-MM-DD)")
	timesheetsListCmd.Flags().StringVar(&timesheetAfter, "after", "", "exclude weeks starting on/before this Monday (YYYY-MM-DD)")

	for _, cmd := range []*cobra.Command{
		timesheetsGetCmd,
		timesheetsSubmitCmd,
		timesheetsRejectCmd,
		timesheetsUnlockCmd,
	} {
		cmd.Flags().StringVar(&timesheetWeekStart, "week-start", "", "Monday week start (default: this Monday)")
	}

	timesheetsWeekCmd.Flags().StringVar(&timesheetRosterWeek, "week-start", "", "Monday week start (default: this Monday)")
	for _, cmd := range []*cobra.Command{timesheetsWeekCmd, timesheetsLastWeekCmd} {
		cmd.Flags().StringVar(&timesheetRosterStatus, "status", "all", "filter rows: submitted, draft, or all")
	}

	timesheetsPurgeCmd.Flags().StringVar(&timesheetWeekStart, "week-start", "", "purge one week (Monday YYYY-MM-DD)")
	timesheetsPurgeCmd.Flags().StringVar(&timesheetBefore, "before", "", "purge all weeks before this Monday (exclusive)")
	timesheetsPurgeCmd.Flags().BoolVar(&timesheetConfirm, "confirm", false, "confirm destructive purge")
}
