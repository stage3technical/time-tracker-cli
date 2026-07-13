package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var adminCmd = &cobra.Command{
	Use:   "admin",
	Short: "Admin operations (require Cognito admins group)",
}

var adminBackportCmd = &cobra.Command{
	Use:   "backport",
	Short: "Environment data backport (dev only)",
}

var adminBackportFromProdCmd = &cobra.Command{
	Use:   "from-prod",
	Short: "Copy all prod DynamoDB data into dev (wipes dev first)",
	Long: `Starts an async backport job on the dev API.
Requires a dev admin JWT (--profile dev) and --confirm app-dev-main.`,
	RunE: runAdminBackportFromProd,
}

var (
	backportConfirm string
	backportDryRun  bool
	backportWait    bool
)

func init() {
	adminBackportFromProdCmd.Flags().StringVar(&backportConfirm, "confirm", "", "must be app-dev-main")
	adminBackportFromProdCmd.Flags().BoolVar(&backportDryRun, "dry-run", false, "report counts only; no writes")
	adminBackportFromProdCmd.Flags().BoolVar(&backportWait, "wait", false, "poll until the job completes")

	adminBackportCmd.AddCommand(adminBackportFromProdCmd)
	adminCmd.AddCommand(adminBackportCmd)
	register(rootCmd, adminCmd, CapWrite)
}

func runAdminBackportFromProd(cmd *cobra.Command, args []string) error {
	if !backportDryRun {
		if err := requireConfirm(backportConfirm == "app-dev-main", "backport from prod"); err != nil {
			return err
		}
	}

	c, err := resolveClient(true)
	if err != nil {
		return err
	}

	path := "/api/v1/admin/backport/jobs"
	if backportDryRun {
		path += "?dryRun=true"
	}

	body, err := json.Marshal(map[string]any{
		"confirm": backportConfirm,
	})
	if err != nil {
		return err
	}

	resp, err := c.Do("POST", path, nil, body)
	if err != nil {
		handleAPIError(err)
	}

	var result map[string]any
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return fmt.Errorf("parse response: %w", err)
	}

	if err := printResponse(resp.StatusCode, resp.Body); err != nil {
		return err
	}

	if backportDryRun || !backportWait {
		return nil
	}

	jobID, _ := result["id"].(string)
	if strings.TrimSpace(jobID) == "" {
		return nil
	}

	for {
		time.Sleep(2 * time.Second)
		statusResp, err := c.Get("/api/v1/admin/backport/jobs/"+jobID, nil)
		if err != nil {
			handleAPIError(err)
		}
		var job map[string]any
		if err := json.Unmarshal(statusResp.Body, &job); err != nil {
			return fmt.Errorf("parse job status: %w", err)
		}
		status, _ := job["status"].(string)
		out.PrintError(fmt.Sprintf("backport job %s: %s (copied=%v deleted=%v)", jobID, status, job["itemsCopied"], job["itemsDeleted"]))
		switch status {
		case "completed":
			if err := printResponse(statusResp.StatusCode, statusResp.Body); err != nil {
				return err
			}
			return nil
		case "failed", "cancelled":
			return fmt.Errorf("backport job %s ended with status %s", jobID, status)
		}
	}
}
