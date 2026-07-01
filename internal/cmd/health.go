package cmd

import (
	"github.com/spf13/cobra"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check API liveness (GET /health)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := resolveClient(false)
		if err != nil {
			return err
		}
		resp, err := c.Get("/health", nil)
		if err != nil {
			handleAPIError(err)
		}
		return printResponse(resp.StatusCode, resp.Body)
	},
}

var meCmd = &cobra.Command{
	Use:   "me",
	Short: "Show current authenticated user (GET /me)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := resolveClient(true)
		if err != nil {
			return err
		}
		resp, err := c.Get("/me", nil)
		if err != nil {
			handleAPIError(err)
		}
		return printResponse(resp.StatusCode, resp.Body)
	},
}
