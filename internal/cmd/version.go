package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stage3technical/time-tracker-cli/internal/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print tt version",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := fmt.Fprintln(cmd.OutOrStdout(), version.String())
		return err
	},
}

func init() {
	register(rootCmd, versionCmd, CapRead)
}
