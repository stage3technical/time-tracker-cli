package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stage3technical/time-tracker-cli/internal/config"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Interactive profile setup (AWS configure style)",
	RunE: func(cmd *cobra.Command, args []string) error {
		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Profile name [default]: ")
		name, _ := reader.ReadString('\n')
		name = strings.TrimSpace(name)
		if name == "" {
			name = "default"
		}

		fmt.Print("API base URL: ")
		baseURL, _ := reader.ReadString('\n')
		baseURL = strings.TrimSpace(baseURL)
		if baseURL == "" {
			return fmt.Errorf("base URL is required")
		}

		fmt.Print("Bearer token (JWT): ")
		token, _ := reader.ReadString('\n')
		token = strings.TrimSpace(token)
		if token == "" {
			return fmt.Errorf("token is required")
		}

		f, err := config.Load()
		if err != nil {
			return err
		}
		if err := config.SetProfile(f, name, baseURL, token); err != nil {
			return err
		}
		if err := config.SetDefaultProfile(f, name); err != nil {
			return err
		}
		if err := config.Save(f); err != nil {
			return err
		}

		fmt.Printf("Saved profile %q to %s\n", name, mustConfigPath())
		return nil
	},
}

var configureListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		defaultName := config.DefaultProfileName(cfgFile)
		names := config.ListProfiles(cfgFile)
		if len(names) == 0 {
			fmt.Println("No profiles configured. Run `tt configure`.")
			return nil
		}
		for _, name := range names {
			baseURL, token, err := config.GetProfile(cfgFile, name)
			if err != nil {
				continue
			}
			marker := " "
			if name == defaultName {
				marker = "*"
			}
			fmt.Printf("%s %s\n  base_url: %s\n  token: %s\n", marker, name, baseURL, config.MaskToken(token))
		}
		return nil
	},
}

var (
	setProfileName string
	setBaseURL     string
	setToken       string
)

var configureSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set profile values non-interactively",
	RunE: func(cmd *cobra.Command, args []string) error {
		if setProfileName == "" {
			return fmt.Errorf("--profile is required")
		}
		if setBaseURL == "" && setToken == "" {
			return fmt.Errorf("at least one of --base-url or --token is required")
		}

		f, err := config.Load()
		if err != nil {
			return err
		}

		existingBase, existingToken, _ := config.GetProfile(f, setProfileName)
		if setBaseURL == "" {
			setBaseURL = existingBase
		}
		if setToken == "" {
			setToken = existingToken
		}

		if err := config.SetProfile(f, setProfileName, setBaseURL, setToken); err != nil {
			return err
		}
		if err := config.SetDefaultProfile(f, setProfileName); err != nil {
			return err
		}
		if err := config.Save(f); err != nil {
			return err
		}

		fmt.Printf("Updated profile %q\n", setProfileName)
		return nil
	},
}

func init() {
	register(configureCmd, configureListCmd, CapLocal)
	register(configureCmd, configureSetCmd, CapLocal)

	configureSetCmd.Flags().StringVar(&setProfileName, "profile", "", "profile name")
	configureSetCmd.Flags().StringVar(&setBaseURL, "base-url", "", "API base URL")
	configureSetCmd.Flags().StringVar(&setToken, "token", "", "JWT bearer token")
	_ = configureSetCmd.MarkFlagRequired("profile")
}

func mustConfigPath() string {
	p, err := config.ConfigPath()
	if err != nil {
		return "~/.tt/config"
	}
	return p
}
