# time-tracker-cli (`tt`)

Cross-platform command-line client for the [BLVD Timesheet API](https://github.com/stage3technical/time-tracker-api). Single static Go binary for Windows and Linux/macOS.

## Quick start

```powershell
# Build (Windows) — embeds git tag/commit
.\scripts\build.ps1

# Or plain build (shows version "dev")
# go build -o tt.exe ./cmd/tt

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

## Features

- AWS-style profile config at `~/.tt/config`
- `tt configure`, `tt configure list`, `tt configure set`
- `tt health`, `tt me`
- `tt persons` — list, get, update, import, manager get/set, subordinates list
- `tt timesheets` — list, get, submit, approve, reject, unlock, purge
- `tt entries` — list, get, create, update, delete (destructive requires `--confirm`)
- `tt projects` — list, get, create, update, archive (lookup by `--name` / `--code`)
- `tt company-roles` — list, get, create, update, delete
- `tt version` — print semver / build info
- `tt api` — generic escape hatch for endpoints not yet wrapped
- `--output json|pretty`, env overrides (`TT_API_*`)

## Documentation

- [docs/CLI.md](docs/CLI.md) — install and config
- [docs/ACTIONS.md](docs/ACTIONS.md) — **commands by API action group**
- [docs/PLAN.md](docs/PLAN.md) — design and roadmap
- [CONTRIBUTING.md](CONTRIBUTING.md) — PR workflow

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

Internal BLVD Interactive tooling.
