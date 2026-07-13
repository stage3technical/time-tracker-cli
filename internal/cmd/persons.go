package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stage3technical/time-tracker-cli/internal/output"
)

var personsCmd = &cobra.Command{
	Use:   "persons",
	Short: "Person (employee) operations",
}

var (
	personsListStatus string
	personsListType   string
)

var personsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List persons (GET /api/v1/persons)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := resolveClient(true)
		if err != nil {
			return err
		}
		q := url.Values{}
		if personsListStatus != "" {
			q.Set("status", personsListStatus)
		}
		if personsListType != "" {
			q.Set("type", personsListType)
		}
		resp, err := c.Get("/api/v1/persons", q)
		if err != nil {
			handleAPIError(err)
		}
		if out.Mode == output.ModePretty {
			return out.PrintPersonsList(resp.Body)
		}
		return printResponse(resp.StatusCode, resp.Body)
	},
}

var personsGetCmd = &cobra.Command{
	Use:   "get ID",
	Short: "Get a person by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := resolveClient(true)
		if err != nil {
			return err
		}
		resp, err := c.Get("/api/v1/persons/"+args[0], nil)
		if err != nil {
			handleAPIError(err)
		}
		return printResponse(resp.StatusCode, resp.Body)
	},
}

var (
	updateName           string
	updateEmail          string
	updatePrimaryRole    string
	updateEmploymentType string
	updateTeam           string
)

var personsUpdateCmd = &cobra.Command{
	Use:   "update ID",
	Short: "Update person fields (PUT /api/v1/persons/{id})",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		payload := map[string]string{}
		if updateName != "" {
			payload["name"] = updateName
		}
		if updateEmail != "" {
			payload["email"] = updateEmail
		}
		if updatePrimaryRole != "" {
			payload["primaryRole"] = updatePrimaryRole
		}
		if updateEmploymentType != "" {
			payload["employmentType"] = normalizeEmploymentType(updateEmploymentType)
		}
		if updateTeam != "" {
			payload["team"] = updateTeam
		}
		if len(payload) == 0 {
			return fmt.Errorf("at least one field flag is required")
		}

		body, err := json.Marshal(payload)
		if err != nil {
			return err
		}

		c, err := resolveClient(true)
		if err != nil {
			return err
		}
		resp, err := c.Do("PUT", "/api/v1/persons/"+args[0], nil, body)
		if err != nil {
			handleAPIError(err)
		}
		return printResponse(resp.StatusCode, resp.Body)
	},
}

var (
	importFile        string
	importOnDuplicate string
)

var personsImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import a person (POST /api/v1/persons/import)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if importFile == "" {
			return fmt.Errorf("--file is required")
		}
		data, err := os.ReadFile(importFile)
		if err != nil {
			return fmt.Errorf("read file: %w", err)
		}
		if !json.Valid(data) {
			return fmt.Errorf("file must contain valid JSON")
		}

		q := url.Values{}
		if importOnDuplicate != "" {
			q.Set("onDuplicate", importOnDuplicate)
		}

		c, err := resolveClient(true)
		if err != nil {
			return err
		}
		resp, err := c.Do("POST", "/api/v1/persons/import", q, data)
		if err != nil {
			handleAPIError(err)
		}
		return printResponse(resp.StatusCode, resp.Body)
	},
}

var personsManagerCmd = &cobra.Command{
	Use:   "manager",
	Short: "Manager relationship operations",
}

var personsManagerGetCmd = &cobra.Command{
	Use:   "get PERSON_ID",
	Short: "Get a person's manager",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := resolveClient(true)
		if err != nil {
			return err
		}
		resp, err := c.Get("/api/v1/persons/"+args[0]+"/manager", nil)
		if err != nil {
			handleAPIError(err)
		}
		return printResponse(resp.StatusCode, resp.Body)
	},
}

var managerSetID string

var personsManagerSetCmd = &cobra.Command{
	Use:   "set PERSON_ID",
	Short: "Set a person's manager (PUT /api/v1/persons/{id}/manager)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if managerSetID == "" {
			return fmt.Errorf("--manager-id is required")
		}
		body, err := json.Marshal(map[string]string{"managerId": managerSetID})
		if err != nil {
			return err
		}
		c, err := resolveClient(true)
		if err != nil {
			return err
		}
		resp, err := c.Do("PUT", "/api/v1/persons/"+args[0]+"/manager", nil, body)
		if err != nil {
			handleAPIError(err)
		}
		return printResponse(resp.StatusCode, resp.Body)
	},
}

var personsSubordinatesCmd = &cobra.Command{
	Use:   "subordinates",
	Short: "Subordinate relationship operations",
}

var personsSubordinatesListCmd = &cobra.Command{
	Use:   "list MANAGER_ID",
	Short: "List subordinates for a manager",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := resolveClient(true)
		if err != nil {
			return err
		}
		resp, err := c.Get("/api/v1/persons/"+args[0]+"/subordinates", nil)
		if err != nil {
			handleAPIError(err)
		}
		return printResponse(resp.StatusCode, resp.Body)
	},
}

func init() {
	register(personsCmd, personsListCmd, CapRead)
	register(personsCmd, personsGetCmd, CapRead)
	register(personsCmd, personsUpdateCmd, CapWrite)
	register(personsCmd, personsImportCmd, CapWrite)
	register(personsCmd, personsManagerCmd, CapRead)
	register(personsCmd, personsSubordinatesCmd, CapRead)

	personsListCmd.Flags().StringVar(&personsListStatus, "status", "", "filter by status (e.g. active)")
	personsListCmd.Flags().StringVar(&personsListType, "type", "", "filter by employment type (W2, 1099)")

	personsUpdateCmd.Flags().StringVar(&updateName, "name", "", "person name")
	personsUpdateCmd.Flags().StringVar(&updateEmail, "email", "", "email address")
	personsUpdateCmd.Flags().StringVar(&updatePrimaryRole, "primary-role", "", "primary role")
	personsUpdateCmd.Flags().StringVar(&updateEmploymentType, "employment-type", "", "employment type (W2, 1099)")
	personsUpdateCmd.Flags().StringVar(&updateTeam, "team", "", "team name")

	personsImportCmd.Flags().StringVar(&importFile, "file", "", "JSON payload file")
	personsImportCmd.Flags().StringVar(&importOnDuplicate, "on-duplicate", "update", "on duplicate: update|skip|fail")
	_ = personsImportCmd.MarkFlagRequired("file")

	register(personsManagerCmd, personsManagerGetCmd, CapRead)
	register(personsManagerCmd, personsManagerSetCmd, CapWrite)
	personsManagerSetCmd.Flags().StringVar(&managerSetID, "manager-id", "", "manager person ID")
	_ = personsManagerSetCmd.MarkFlagRequired("manager-id")

	register(personsSubordinatesCmd, personsSubordinatesListCmd, CapRead)
}

// normalizeEmploymentType maps common inputs to API values.
func normalizeEmploymentType(s string) string {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "W2", "W-2":
		return "W2"
	case "1099", "CONTRACTOR":
		return "1099"
	default:
		return s
	}
}
