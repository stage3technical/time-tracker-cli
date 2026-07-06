# Contributing — time-tracker-cli

## Scope

This CLI wraps the **Time Tracker HTTP API only** — no direct DynamoDB access. See `.cursor/rules/cli-via-api.mdc`.

## Git and PR workflow

Follow [time-tracker-api/docs/PROJECT_APPROACH.md](https://github.com/stage3technical/time-tracker-api/blob/main/docs/PROJECT_APPROACH.md) § Git and PR workflow and § Releases (semver).

- Branch from `develop` (or `main` if this repo has no `develop` yet).
- User merges PRs; agents do not run `gh pr merge` unless asked.

## Shell

Windows agents use PowerShell — see api `PROJECT_APPROACH.md` § Shell environment.

## Build and test

```powershell
go test ./...
go build -o tt.exe ./cmd/tt
```
