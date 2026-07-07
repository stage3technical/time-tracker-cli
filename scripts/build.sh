#!/usr/bin/env bash
# Build tt with version metadata from git (repo root).
set -euo pipefail
cd "$(dirname "$0")/.."

version=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
commit=$(git rev-parse HEAD 2>/dev/null || echo "none")
date=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

ldflags="-s -w"
ldflags+=" -X github.com/stage3technical/time-tracker-cli/internal/version.Version=${version}"
ldflags+=" -X github.com/stage3technical/time-tracker-cli/internal/version.Commit=${commit}"
ldflags+=" -X github.com/stage3technical/time-tracker-cli/internal/version.Date=${date}"

out=tt
if [[ "${1:-}" == "--windows" ]] || [[ "$(go env GOOS)" == "windows" ]]; then
  out=tt.exe
fi

go build -ldflags "$ldflags" -o "$out" ./cmd/tt
"./$out" version
