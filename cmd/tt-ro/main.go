package main

import (
	"os"

	"github.com/stage3technical/time-tracker-cli/internal/cmd"
)

func main() {
	if err := cmd.Execute(cmd.ModeReadOnly); err != nil {
		os.Exit(1)
	}
}
