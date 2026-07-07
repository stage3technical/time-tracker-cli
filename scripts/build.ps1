# Build tt.exe with version metadata from git.
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"
Set-Location (Join-Path $PSScriptRoot "..")

$version = "dev"
try {
    $version = (git describe --tags --always --dirty 2>$null)
    if (-not $version) { $version = "dev" }
} catch { }

$commit = "none"
try { $commit = git rev-parse HEAD } catch { }

$date = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")

$ldflags = @(
    "-s", "-w",
    "-X", "github.com/stage3technical/time-tracker-cli/internal/version.Version=$version",
    "-X", "github.com/stage3technical/time-tracker-cli/internal/version.Commit=$commit",
    "-X", "github.com/stage3technical/time-tracker-cli/internal/version.Date=$date"
)

go build -ldflags ($ldflags -join " ") -o tt.exe ./cmd/tt
& .\tt.exe version
