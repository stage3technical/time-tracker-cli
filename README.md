# time-tracker-cli (`tt`)

Cross-platform command-line client for the [Time Tracker API](https://github.com/stage3technical/time-tracker-api). Single static Go binary for Windows and Linux/macOS.

## Quick start

```powershell
# Build (Windows)
go build -o tt.exe ./cmd/tt

# Configure dev profile (paste JWT from browser)
.\tt.exe configure set --profile dev `
  --base-url https://8igr6pspqh.execute-api.us-east-1.amazonaws.com `
  --token "<JWT>"

# Verify
.\tt.exe health
.\tt.exe me
.\tt.exe persons list --status active
```

```bash
# Linux / macOS / WSL
go build -o tt ./cmd/tt
tt configure set --profile dev \
  --base-url https://8igr6pspqh.execute-api.us-east-1.amazonaws.com \
  --token "$JWT"
tt health && tt me
```

## Features (Phase 1)

- AWS-style profile config at `~/.tt/config`
- `tt configure`, `tt configure list`, `tt configure set`
- `tt health`, `tt me`
- `tt persons` — list, get, update, import, manager get/set, subordinates list
- `tt api` — generic METHOD + path escape hatch
- `--output json|pretty`, env overrides (`TT_API_*`)

## Documentation

- [docs/CLI.md](docs/CLI.md) — full command reference
- [docs/PLAN.md](docs/PLAN.md) — design and roadmap

## Config

Profiles live in `~/.tt/config` (Windows: `%USERPROFILE%\.tt\config`):

```ini
[default]
profile = dev

[profile dev]
base_url = https://8igr6pspqh.execute-api.us-east-1.amazonaws.com
token = eyJ...
```

**Never commit tokens.** Use `tt configure set` or env vars in CI.

## Development

```bash
go test ./...
go build ./cmd/tt
```

CI runs tests and builds on Ubuntu and Windows (see [.github/workflows/ci.yml](.github/workflows/ci.yml)).

## License

Internal Stage3 Technical tooling.
