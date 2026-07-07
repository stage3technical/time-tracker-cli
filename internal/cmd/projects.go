package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stage3technical/time-tracker-cli/internal/output"
)

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Project operations",
}

var (
	projectsListStatus string
	projectName        string
	projectCode        string
	projectBillType    string
	projectSystem      bool
	projectAccountID   string
	projectStartDate   string
	projectEndDate     string
	projectAllowedRoles string
	projectFile        string
	projectConfirm     bool
)

var projectsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List projects (GET /api/v1/projects)",
	RunE:  runProjectsList,
}

var projectsGetCmd = &cobra.Command{
	Use:   "get [PROJECT_ID]",
	Short: "Get a project by ID or lookup",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runProjectsGet,
}

var projectsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a project (POST /api/v1/projects)",
	RunE:  runProjectsCreate,
}

var projectsUpdateCmd = &cobra.Command{
	Use:   "update PROJECT_ID",
	Short: "Update a project (PUT /api/v1/projects/{id})",
	Args:  cobra.ExactArgs(1),
	RunE:  runProjectsUpdate,
}

var projectsArchiveCmd = &cobra.Command{
	Use:   "archive [PROJECT_ID]",
	Short: "Archive a project (DELETE /api/v1/projects/{id})",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runProjectsArchive,
}

func runProjectsList(cmd *cobra.Command, args []string) error {
	c, err := resolveClient(true)
	if err != nil {
		return err
	}
	q := url.Values{}
	if projectsListStatus != "" {
		q.Set("status", projectsListStatus)
	}
	resp, err := c.Get("/api/v1/projects", q)
	if err != nil {
		handleAPIError(err)
	}
	if out.Mode == output.ModePretty {
		return out.PrintProjectsList(resp.Body)
	}
	return printResponse(resp.StatusCode, resp.Body)
}

func runProjectsGet(cmd *cobra.Command, args []string) error {
	projectID := ""
	if len(args) == 1 {
		projectID = args[0]
	}
	id, err := resolveProjectID(projectID, projectName, projectCode)
	if err != nil {
		return err
	}
	c, err := resolveClient(true)
	if err != nil {
		return err
	}
	resp, err := c.Get("/api/v1/projects/"+id, nil)
	if err != nil {
		handleAPIError(err)
	}
	return printResponse(resp.StatusCode, resp.Body)
}

func runProjectsCreate(cmd *cobra.Command, args []string) error {
	fields := map[string]any{
		"canonicalName":   strings.TrimSpace(projectName),
		"billType":        strings.TrimSpace(projectBillType),
		"isSystemProject": projectSystem,
		"accountId":       strings.TrimSpace(projectAccountID),
		"startDate":       strings.TrimSpace(projectStartDate),
		"endDate":         strings.TrimSpace(projectEndDate),
		"allowedRoles":    splitCSV(projectAllowedRoles),
	}
	if projectName == "" && projectFile == "" {
		return fmt.Errorf("--name is required (or canonicalName in --file)")
	}
	if projectBillType == "" && projectFile == "" {
		return fmt.Errorf("--bill-type is required (or billType in --file)")
	}
	body, err := mergePayload(projectFile, fields)
	if err != nil {
		return err
	}

	c, err := resolveClient(true)
	if err != nil {
		return err
	}
	resp, err := c.Do("POST", "/api/v1/projects", nil, body)
	if err != nil {
		handleAPIError(err)
	}
	return printResponse(resp.StatusCode, resp.Body)
}

func runProjectsUpdate(cmd *cobra.Command, args []string) error {
	fields := map[string]any{}
	if cmd.Flags().Changed("name") {
		fields["canonicalName"] = strings.TrimSpace(projectName)
	}
	if cmd.Flags().Changed("bill-type") {
		fields["billType"] = strings.TrimSpace(projectBillType)
	}
	if cmd.Flags().Changed("system-project") {
		fields["isSystemProject"] = projectSystem
	}
	if cmd.Flags().Changed("account-id") {
		fields["accountId"] = strings.TrimSpace(projectAccountID)
	}
	if cmd.Flags().Changed("start-date") {
		fields["startDate"] = strings.TrimSpace(projectStartDate)
	}
	if cmd.Flags().Changed("end-date") {
		fields["endDate"] = strings.TrimSpace(projectEndDate)
	}
	if cmd.Flags().Changed("allowed-roles") {
		fields["allowedRoles"] = splitCSV(projectAllowedRoles)
	}
	body, err := mergePayload(projectFile, fields)
	if err != nil {
		return err
	}

	c, err := resolveClient(true)
	if err != nil {
		return err
	}
	resp, err := c.Do("PUT", "/api/v1/projects/"+args[0], nil, body)
	if err != nil {
		handleAPIError(err)
	}
	return printResponse(resp.StatusCode, resp.Body)
}

func runProjectsArchive(cmd *cobra.Command, args []string) error {
	if err := requireConfirm(projectConfirm, "archive project"); err != nil {
		return err
	}
	projectID := ""
	if len(args) == 1 {
		projectID = args[0]
	}
	id, err := resolveProjectID(projectID, projectName, projectCode)
	if err != nil {
		return err
	}
	c, err := resolveClient(true)
	if err != nil {
		return err
	}
	resp, err := c.Do("DELETE", "/api/v1/projects/"+id, nil, nil)
	if err != nil {
		handleAPIError(err)
	}
	return printResponse(resp.StatusCode, resp.Body)
}

func init() {
	rootCmd.AddCommand(projectsCmd)
	projectsCmd.AddCommand(projectsListCmd)
	projectsCmd.AddCommand(projectsGetCmd)
	projectsCmd.AddCommand(projectsCreateCmd)
	projectsCmd.AddCommand(projectsUpdateCmd)
	projectsCmd.AddCommand(projectsArchiveCmd)

	projectsListCmd.Flags().StringVar(&projectsListStatus, "status", "", "filter by status (active, archived)")

	projectsGetCmd.Flags().StringVar(&projectName, "name", "", "lookup by canonical name")
	projectsGetCmd.Flags().StringVar(&projectCode, "code", "", "exact canonical name lookup")

	projectsCreateCmd.Flags().StringVar(&projectName, "name", "", "canonical project name")
	projectsCreateCmd.Flags().StringVar(&projectBillType, "bill-type", "", "bill type (BIL-S, BIL-PO, N-BIL-C, N-BIL-I)")
	projectsCreateCmd.Flags().BoolVar(&projectSystem, "system-project", false, "mark as system project")
	projectsCreateCmd.Flags().StringVar(&projectAccountID, "account-id", "", "linked account UUID")
	projectsCreateCmd.Flags().StringVar(&projectStartDate, "start-date", "", "start date (YYYY-MM-DD)")
	projectsCreateCmd.Flags().StringVar(&projectEndDate, "end-date", "", "end date (YYYY-MM-DD)")
	projectsCreateCmd.Flags().StringVar(&projectAllowedRoles, "allowed-roles", "", "comma-separated roles")
	projectsCreateCmd.Flags().StringVar(&projectFile, "file", "", "JSON payload file (merged with flags)")

	projectsUpdateCmd.Flags().StringVar(&projectName, "name", "", "canonical project name")
	projectsUpdateCmd.Flags().StringVar(&projectBillType, "bill-type", "", "bill type")
	projectsUpdateCmd.Flags().BoolVar(&projectSystem, "system-project", false, "mark as system project")
	projectsUpdateCmd.Flags().StringVar(&projectAccountID, "account-id", "", "linked account UUID")
	projectsUpdateCmd.Flags().StringVar(&projectStartDate, "start-date", "", "start date (YYYY-MM-DD)")
	projectsUpdateCmd.Flags().StringVar(&projectEndDate, "end-date", "", "end date (YYYY-MM-DD)")
	projectsUpdateCmd.Flags().StringVar(&projectAllowedRoles, "allowed-roles", "", "comma-separated roles")
	projectsUpdateCmd.Flags().StringVar(&projectFile, "file", "", "JSON payload file (merged with flags)")

	projectsArchiveCmd.Flags().StringVar(&projectName, "name", "", "lookup by canonical name")
	projectsArchiveCmd.Flags().StringVar(&projectCode, "code", "", "exact canonical name lookup")
	projectsArchiveCmd.Flags().BoolVar(&projectConfirm, "confirm", false, "confirm destructive archive")
}
