package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func resetForTest() {
	rootCmd.ResetCommands()
	for _, cmd := range []*cobra.Command{
		configureCmd, timesheetsCmd, projectsCmd, personsCmd,
		personsManagerCmd, personsSubordinatesCmd, entriesCmd, companyRolesCmd,
	} {
		cmd.ResetCommands()
	}
	commandRegistry = nil
}

func TestReadOnlyMode_HidesWriteCommands(t *testing.T) {
	resetForTest()

	// Re-register a minimal subset mirroring production commands.
	register(rootCmd, configureCmd, CapLocal)
	register(configureCmd, configureListCmd, CapLocal)
	register(rootCmd, apiCmd, CapWrite)
	register(rootCmd, timesheetsCmd, CapRead)
	register(timesheetsCmd, timesheetsListCmd, CapRead)
	register(timesheetsCmd, timesheetsPurgeCmd, CapWrite)
	register(rootCmd, projectsCmd, CapRead)
	register(projectsCmd, projectsArchiveCmd, CapWrite)

	attachCommands(ModeReadOnly)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"--help"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("help: %v", err)
	}
	help := buf.String()

	for _, hidden := range []string{"purge", "archive", "api"} {
		if strings.Contains(help, hidden) {
			t.Errorf("help should not mention %q", hidden)
		}
	}
	for _, visible := range []string{"timesheets", "list", "configure"} {
		if !strings.Contains(help, visible) {
			t.Errorf("help should mention %q", visible)
		}
	}
}

func TestReadOnlyMode_ExcludesWriteSubcommands(t *testing.T) {
	resetForTest()

	register(rootCmd, timesheetsCmd, CapRead)
	register(timesheetsCmd, timesheetsListCmd, CapRead)
	register(timesheetsCmd, timesheetsPurgeCmd, CapWrite)

	attachCommands(ModeReadOnly)

	if hasSubcommand(timesheetsCmd, "purge") {
		t.Fatal("purge should not be registered in read-only mode")
	}
	if !hasSubcommand(timesheetsCmd, "list") {
		t.Fatal("list should be registered in read-only mode")
	}
}

func hasSubcommand(parent *cobra.Command, name string) bool {
	for _, c := range parent.Commands() {
		if c.Name() == name {
			return true
		}
	}
	return false
}

func TestFullMode_IncludesWriteCommands(t *testing.T) {
	resetForTest()

	register(rootCmd, apiCmd, CapWrite)
	register(rootCmd, timesheetsCmd, CapRead)
	register(timesheetsCmd, timesheetsPurgeCmd, CapWrite)

	attachCommands(ModeFull)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"timesheets", "--help"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("help: %v", err)
	}
	help := buf.String()

	if !strings.Contains(help, "purge") {
		t.Errorf("full mode timesheets help should mention purge")
	}
}
