package version

import "fmt"

// Set at link time via -ldflags; defaults for local go build.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// String returns a human-readable version line.
func String() string {
	short := Commit
	if len(short) > 7 {
		short = short[:7]
	}
	return fmt.Sprintf("%s (commit %s, built %s)", Version, short, Date)
}
