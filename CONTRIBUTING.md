# Contributing — time-tracker-cli

## Scope

This CLI wraps the **BLVD Timesheet HTTP API only** — no direct DynamoDB access. See `.cursor/rules/cli-via-api.mdc`.

## Pull requests

- Use a feature branch (`feature/…`, `fix/…`, `docs/…`, `sync/…`, or `promote/…`) — not `chore/`
- Do not use `chore:` in PR titles — prefer `feature:`, `fix:`, `docs:`, `sync:`, or `promote:`
- Open PRs as **ready for review** (not draft) — `gh pr create --base main` without `--draft` (this repo has no `develop` branch)
- **User merges** — do not run `gh pr merge` unless explicitly asked
- When merging, default to **Create a merge commit** unless squash/rebase is requested

Canonical rules: **time-tracker-api** `docs/PROJECT_APPROACH.md` § Git and PR workflow and § Releases (semver).

## Shell

Windows agents use PowerShell — see api `PROJECT_APPROACH.md` § Shell environment.

## Build and test

```powershell
go test ./...
go build -o tt.exe ./cmd/tt
```

See [docs/CLI.md](docs/CLI.md) for command reference.

## Agent rules

See `.cursor/rules/` for git/PR workflow and CLI-via-API constraints.
