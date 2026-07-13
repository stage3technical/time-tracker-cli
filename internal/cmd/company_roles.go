package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stage3technical/time-tracker-cli/internal/output"
)

var companyRolesCmd = &cobra.Command{
	Use:   "company-roles",
	Short: "Company role registry",
}

var (
	companyRoleName        string
	companyRoleDescription string
	companyRoleFile        string
	companyRoleConfirm     bool
)

var companyRolesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List company roles (GET /api/v1/company-roles)",
	RunE:  runCompanyRolesList,
}

var companyRolesGetCmd = &cobra.Command{
	Use:   "get ROLE_ID",
	Short: "Get a company role by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := resolveClient(true)
		if err != nil {
			return err
		}
		resp, err := c.Get("/api/v1/company-roles/"+args[0], nil)
		if err != nil {
			handleAPIError(err)
		}
		return printResponse(resp.StatusCode, resp.Body)
	},
}

var companyRolesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a company role (POST /api/v1/company-roles)",
	RunE:  runCompanyRolesCreate,
}

var companyRolesUpdateCmd = &cobra.Command{
	Use:   "update ROLE_ID",
	Short: "Update a company role (PUT /api/v1/company-roles/{id})",
	Args:  cobra.ExactArgs(1),
	RunE:  runCompanyRolesUpdate,
}

var companyRolesDeleteCmd = &cobra.Command{
	Use:   "delete ROLE_ID",
	Short: "Delete a company role (DELETE /api/v1/company-roles/{id})",
	Args:  cobra.ExactArgs(1),
	RunE:  runCompanyRolesDelete,
}

func runCompanyRolesList(cmd *cobra.Command, args []string) error {
	c, err := resolveClient(true)
	if err != nil {
		return err
	}
	resp, err := c.Get("/api/v1/company-roles", nil)
	if err != nil {
		handleAPIError(err)
	}
	if out.Mode == output.ModePretty {
		return out.PrintCompanyRolesList(resp.Body)
	}
	return printResponse(resp.StatusCode, resp.Body)
}

func runCompanyRolesCreate(cmd *cobra.Command, args []string) error {
	fields := map[string]any{
		"name":        strings.TrimSpace(companyRoleName),
		"description": strings.TrimSpace(companyRoleDescription),
	}
	if companyRoleName == "" && companyRoleFile == "" {
		return fmt.Errorf("--name is required (or name in --file)")
	}
	body, err := mergePayload(companyRoleFile, fields)
	if err != nil {
		return err
	}
	c, err := resolveClient(true)
	if err != nil {
		return err
	}
	resp, err := c.Do("POST", "/api/v1/company-roles", nil, body)
	if err != nil {
		handleAPIError(err)
	}
	return printResponse(resp.StatusCode, resp.Body)
}

func runCompanyRolesUpdate(cmd *cobra.Command, args []string) error {
	fields := map[string]any{}
	if cmd.Flags().Changed("name") {
		fields["name"] = strings.TrimSpace(companyRoleName)
	}
	if cmd.Flags().Changed("description") {
		fields["description"] = strings.TrimSpace(companyRoleDescription)
	}
	body, err := mergePayload(companyRoleFile, fields)
	if err != nil {
		return err
	}
	c, err := resolveClient(true)
	if err != nil {
		return err
	}
	resp, err := c.Do("PUT", "/api/v1/company-roles/"+args[0], nil, body)
	if err != nil {
		handleAPIError(err)
	}
	return printResponse(resp.StatusCode, resp.Body)
}

func runCompanyRolesDelete(cmd *cobra.Command, args []string) error {
	if err := requireConfirm(companyRoleConfirm, "delete company role"); err != nil {
		return err
	}
	c, err := resolveClient(true)
	if err != nil {
		return err
	}
	resp, err := c.Do("DELETE", "/api/v1/company-roles/"+args[0], nil, nil)
	if err != nil {
		handleAPIError(err)
	}
	return printResponse(resp.StatusCode, resp.Body)
}

func init() {
	register(rootCmd, companyRolesCmd, CapRead)
	register(companyRolesCmd, companyRolesListCmd, CapRead)
	register(companyRolesCmd, companyRolesGetCmd, CapRead)
	register(companyRolesCmd, companyRolesCreateCmd, CapWrite)
	register(companyRolesCmd, companyRolesUpdateCmd, CapWrite)
	register(companyRolesCmd, companyRolesDeleteCmd, CapWrite)

	companyRolesCreateCmd.Flags().StringVar(&companyRoleName, "name", "", "role name")
	companyRolesCreateCmd.Flags().StringVar(&companyRoleDescription, "description", "", "optional description")
	companyRolesCreateCmd.Flags().StringVar(&companyRoleFile, "file", "", "JSON payload file (merged with flags)")

	companyRolesUpdateCmd.Flags().StringVar(&companyRoleName, "name", "", "role name")
	companyRolesUpdateCmd.Flags().StringVar(&companyRoleDescription, "description", "", "description")
	companyRolesUpdateCmd.Flags().StringVar(&companyRoleFile, "file", "", "JSON payload file (merged with flags)")

	companyRolesDeleteCmd.Flags().BoolVar(&companyRoleConfirm, "confirm", false, "confirm destructive delete")
}
