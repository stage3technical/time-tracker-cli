package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
)

var timesheetsCmd = &cobra.Command{
	Use:   "timesheets",
	Short: "Timesheet workflow (Advanced Workflow API)",
}

var (
	timesheetWeekStart string
	timesheetEmail     string
	timesheetPersonID  string
)

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
	timesheetsCmd.AddCommand(timesheetsGetCmd)
	timesheetsCmd.AddCommand(timesheetsSubmitCmd)
	timesheetsCmd.AddCommand(timesheetsApproveCmd)
	timesheetsCmd.AddCommand(timesheetsRejectCmd)
	timesheetsCmd.AddCommand(timesheetsUnlockCmd)

	for _, cmd := range []*cobra.Command{
		timesheetsGetCmd,
		timesheetsSubmitCmd,
		timesheetsApproveCmd,
		timesheetsRejectCmd,
		timesheetsUnlockCmd,
	} {
		cmd.Flags().StringVar(&timesheetWeekStart, "week-start", "", "Monday week start (default: this Monday)")
		cmd.Flags().StringVar(&timesheetPersonID, "person-id", "", "person UUID")
		cmd.Flags().StringVar(&timesheetEmail, "email", "", "person email (looks up ID via GET /persons)")
	}
}
