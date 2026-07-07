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
	timesheetWeekStart string
	timesheetBefore    string
	timesheetAfter     string
	timesheetEmail     string
	timesheetPersonID  string
	timesheetConfirm   bool
)

var timesheetsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all weeks for a person (GET /api/v1/timesheets/{personId}/weeks)",
	RunE:  runTimesheetList,
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

var timesheetsApproveCmd = &cobra.Command{
	Use:   "approve",
	Short: "Approve and lock a week (POST /api/v1/timesheets/{personId}/approve)",
	RunE:  runTimesheetApprove,
}

var timesheetsRejectCmd = &cobra.Command{
	Use:   "reject",
	Short: "Reject a submitted week (POST /api/v1/timesheets/{personId}/reject)",
	RunE:  runTimesheetReject,
}

var timesheetsUnlockCmd = &cobra.Command{
	Use:   "unlock",
	Short: "Admin unlock: reopen a person's week for editing (POST /api/v1/timesheets/{personId}/unlock)",
	Long: `Unlock reverts the person's entries and submission to draft for the week.
If the week is globally locked, it is reopened for everyone on that week.`,
	RunE: runTimesheetUnlock,
}

var timesheetsPurgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Admin purge: delete entries and submission for week(s) (POST /api/v1/timesheets/{personId}/purge)",
	Long: `Deletes all time entries and the WeekSubmission for the person/week.
Does not change global WeekLock (other people on that week are unaffected).
Requires --confirm. Use --week-start for one week or --before for all prior weeks.`,
	RunE: runTimesheetPurge,
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

func runTimesheetApprove(cmd *cobra.Command, args []string) error {
	return timesheetPersonAction("POST", "approve")
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
	rootCmd.AddCommand(timesheetsCmd)
	timesheetsCmd.AddCommand(timesheetsListCmd)
	timesheetsCmd.AddCommand(timesheetsGetCmd)
	timesheetsCmd.AddCommand(timesheetsSubmitCmd)
	timesheetsCmd.AddCommand(timesheetsApproveCmd)
	timesheetsCmd.AddCommand(timesheetsRejectCmd)
	timesheetsCmd.AddCommand(timesheetsUnlockCmd)
	timesheetsCmd.AddCommand(timesheetsPurgeCmd)

	for _, cmd := range []*cobra.Command{
		timesheetsListCmd,
		timesheetsGetCmd,
		timesheetsSubmitCmd,
		timesheetsApproveCmd,
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
		timesheetsApproveCmd,
		timesheetsRejectCmd,
		timesheetsUnlockCmd,
	} {
		cmd.Flags().StringVar(&timesheetWeekStart, "week-start", "", "Monday week start (default: this Monday)")
	}

	timesheetsPurgeCmd.Flags().StringVar(&timesheetWeekStart, "week-start", "", "purge one week (Monday YYYY-MM-DD)")
	timesheetsPurgeCmd.Flags().StringVar(&timesheetBefore, "before", "", "purge all weeks before this Monday (exclusive)")
	timesheetsPurgeCmd.Flags().BoolVar(&timesheetConfirm, "confirm", false, "confirm destructive purge")
}
