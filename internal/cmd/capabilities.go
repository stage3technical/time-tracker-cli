package cmd

import "github.com/spf13/cobra"

// Mode selects which commands are registered on the root command tree.
type Mode int

const (
	ModeFull Mode = iota
	ModeReadOnly
)

// Capability classifies a command for read-only filtering.
type Capability int

const (
	CapLocal Capability = iota
	CapRead
	CapWrite
)

var cliMode Mode

type commandSpec struct {
	parent *cobra.Command
	cmd    *cobra.Command
	cap    Capability
}

var commandRegistry []commandSpec

func register(parent *cobra.Command, cmd *cobra.Command, cap Capability) {
	commandRegistry = append(commandRegistry, commandSpec{parent: parent, cmd: cmd, cap: cap})
}

func capAllowed(cap Capability) bool {
	if cliMode == ModeFull {
		return true
	}
	return cap == CapLocal || cap == CapRead
}

func attachCommands(mode Mode) {
	cliMode = mode
	if mode == ModeReadOnly {
		rootCmd.Use = "tt-ro"
		rootCmd.Short = "Time Tracker API CLI (read-only)"
		rootCmd.Long = "tt-ro is a read-only command-line client for the Time Tracker API. " +
			"It can list and fetch data but cannot create, update, delete, or run workflow actions."
	}

	visible := make(map[*cobra.Command]bool)
	for _, spec := range commandRegistry {
		if capAllowed(spec.cap) {
			visible[spec.cmd] = true
		}
	}

	for _, spec := range commandRegistry {
		if !visible[spec.cmd] {
			continue
		}
		if spec.parent == rootCmd || visible[spec.parent] {
			spec.parent.AddCommand(spec.cmd)
		}
	}
}
