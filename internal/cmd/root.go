package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/stage3technical/time-tracker-cli/internal/client"
	"github.com/stage3technical/time-tracker-cli/internal/config"
	"github.com/stage3technical/time-tracker-cli/internal/output"
	"github.com/stage3technical/time-tracker-cli/internal/version"
	"gopkg.in/ini.v1"
)

var (
	flagProfile string
	flagBaseURL string
	flagToken   string
	flagOutput  string
	flagQuiet   bool

	cfgFile *ini.File
	out     *output.Writer
)

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

var rootCmd = &cobra.Command{
	Use:   "tt",
	Short: "Time Tracker API CLI",
	Long:  "tt is a command-line client for the Time Tracker API.",
	Version: version.Version,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if !commandNeedsConfig(cmd) {
			mode, err := output.ParseOutputFlag(flagOutput)
			if err != nil {
				return err
			}
			out = output.NewWriter(mode, flagQuiet)
			return nil
		}
		mode, err := output.ParseOutputFlag(flagOutput)
		if err != nil {
			return err
		}
		out = output.NewWriter(mode, flagQuiet)

		f, err := config.Load()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		cfgFile = f
		return nil
	},
	SilenceUsage: true,
}

func commandNeedsConfig(cmd *cobra.Command) bool {
	for c := cmd; c != nil; c = c.Parent() {
		if c.Name() == "version" {
			return false
		}
	}
	return true
}

func init() {
	rootCmd.SetVersionTemplate("{{.Version}}\n")
	rootCmd.PersistentFlags().StringVar(&flagProfile, "profile", "", "config profile name")
	rootCmd.PersistentFlags().StringVar(&flagBaseURL, "base-url", "", "API base URL override")
	rootCmd.PersistentFlags().StringVar(&flagToken, "token", "", "JWT bearer token override")
	rootCmd.PersistentFlags().StringVar(&flagOutput, "output", "", "output format: json|pretty")
	rootCmd.PersistentFlags().BoolVar(&flagQuiet, "quiet", false, "suppress non-essential stderr")

	rootCmd.AddCommand(configureCmd)
	rootCmd.AddCommand(healthCmd)
	rootCmd.AddCommand(meCmd)
	rootCmd.AddCommand(apiCmd)
	rootCmd.AddCommand(personsCmd)
}

func flagOverrides() config.FlagOverrides {
	return config.FlagOverrides{
		Profile: flagProfile,
		BaseURL: flagBaseURL,
		Token:   flagToken,
	}
}

func resolveClient(requireAuth bool) (*client.Client, error) {
	if requireAuth {
		resolved, err := config.Resolve(cfgFile, flagOverrides())
		if err != nil {
			return nil, err
		}
		return client.New(resolved.BaseURL, resolved.Token), nil
	}
	resolved := config.ResolveOptional(cfgFile, flagOverrides())
	if resolved.BaseURL == "" {
		return nil, fmt.Errorf("base URL is required (flag, %s, or profile)", config.EnvBaseURL)
	}
	return client.New(resolved.BaseURL, resolved.Token), nil
}

func handleAPIError(err error) {
	if apiErr, ok := err.(*client.APIError); ok {
		detail := client.ParseDetail(apiErr.Body)
		out.PrintError(fmt.Sprintf("error: %s", detail))
		os.Exit(client.ExitCode(apiErr.StatusCode))
	}
	out.PrintError(err.Error())
	os.Exit(1)
}

func printResponse(status int, body []byte) error {
	if status == 204 || len(body) == 0 {
		return nil
	}
	return out.PrintJSON(body)
}
