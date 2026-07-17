package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stage3technical/time-tracker-cli/internal/client"
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
If the week is globally locked, it is reopened for everyone on that week.
Also records a per-person unlock exception (see timesheets relock).`,
	RunE: runTimesheetUnlock,
}

var timesheetsRelockCmd = &cobra.Command{
	Use:   "relock",
	Short: "Admin re-lock: clear unlock exception and freeze person entries (POST /api/v1/timesheets/{personId}/relock)",
	Long: `Clears the person's unlock exception and sets their entries to locked.
Does not require the person to resubmit. When no unlock exceptions remain for
the week, restores the global week lock.`,
	Example: `  tt timesheets relock --email david.mead@blvdinteractive.com
  tt timesheets relock --email david.mead@blvdinteractive.com --week-start 2026-07-06`,
	RunE: runTimesheetRelock,
}

var timesheetsLockWeekCmd = &cobra.Command{
	Use:   "lock-week",
	Short: "Admin lock: globally lock a Monday week (POST /api/v1/timesheets/weeks/{weekStartDate}/lock)",
	Long: `Locks the given Monday week for everyone (admin recovery when the week is open).
Defaults to this Monday. Distinct from lock-prior (scheduler secret).`,
	Example: `  tt timesheets lock-week
  tt timesheets lock-week --week-start 2026-07-06 --profile prod`,
	RunE: runTimesheetLockWeek,
}

var timesheetsLockPriorCmd = &cobra.Command{
	Use:   "lock-prior",
	Short: "Globally lock the prior Monday week (POST /api/v1/timesheets/weeks/lock-prior)",
	Long: `Ops/scheduler command: locks the prior Monday week at the submission deadline.
Requires TT_SCHEDULER_SECRET (X-Scheduler-Secret header). Does not run manager approve.`,
	RunE: runTimesheetLockPrior,
}

var timesheetsPurgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Admin purge: delete entries and submission for week(s) (POST /api/v1/timesheets/{personId}/purge)",
	Long: `Deletes all time entries and the WeekSubmission for the person/week.
Does not change global WeekLock (other people on that week are unaffected).
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

func runTimesheetApprove(cmd *cobra.Command, args []string) error {
	return timesheetPersonAction("POST", "approve")
}

func runTimesheetReject(cmd *cobra.Command, args []string) error {
	return timesheetPersonAction("POST", "reject")
}

func runTimesheetUnlock(cmd *cobra.Command, args []string) error {
	return timesheetPersonAction("POST", "unlock")
}

func runTimesheetRelock(cmd *cobra.Command, args []string) error {
	return timesheetPersonAction("POST", "relock")
}

func runTimesheetLockWeek(cmd *cobra.Command, args []string) error {
	weekStart := weekStartOrDefault(timesheetWeekStart)
	c, err := resolveClient(true)
	if err != nil {
		return err
	}
	path := "/api/v1/timesheets/weeks/" + weekStart + "/lock"
	resp, err := c.Do("POST", path, nil, nil)
	if err != nil {
		handleAPIError(err)
	}
	return printResponse(resp.StatusCode, resp.Body)
}

func runTimesheetLockPrior(cmd *cobra.Command, args []string) error {
	secret := strings.TrimSpace(os.Getenv("TT_SCHEDULER_SECRET"))
	if secret == "" {
		return fmt.Errorf("TT_SCHEDULER_SECRET is required for lock-prior")
	}
	c, err := resolveClient(false)
	if err != nil {
		return err
	}
	api, ok := c.(*client.Client)
	if !ok {
		return fmt.Errorf("lock-prior requires full write client")
	}
	api.ExtraHeaders = map[string]string{"X-Scheduler-Secret": secret}
	resp, err := api.Do("POST", "/api/v1/timesheets/weeks/lock-prior", nil, nil)
	if err != nil {
		handleAPIError(err)
	}
	return printResponse(resp.StatusCode, resp.Body)
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
	register(timesheetsCmd, timesheetsApproveCmd, CapWrite)
	register(timesheetsCmd, timesheetsRejectCmd, CapWrite)
	register(timesheetsCmd, timesheetsUnlockCmd, CapWrite)
	register(timesheetsCmd, timesheetsRelockCmd, CapWrite)
	register(timesheetsCmd, timesheetsLockWeekCmd, CapWrite)
	register(timesheetsCmd, timesheetsLockPriorCmd, CapWrite)
	register(timesheetsCmd, timesheetsPurgeCmd, CapWrite)

	for _, cmd := range []*cobra.Command{
		timesheetsListCmd,
		timesheetsGetCmd,
		timesheetsSubmitCmd,
		timesheetsApproveCmd,
		timesheetsRejectCmd,
		timesheetsUnlockCmd,
		timesheetsRelockCmd,
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
		timesheetsRelockCmd,
		timesheetsLockWeekCmd,
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
