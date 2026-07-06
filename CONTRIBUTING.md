# Contributing to time-tracker-cli

## Pull requests

- Use a feature branch (`feature/…`, `fix/…`, or `docs/…`) — not `chore/`
- Open PRs as **ready for review** (not draft) — `gh pr create --base main` without `--draft` (this repo has no `develop` branch)
- **User merges** — do not run `gh pr merge` unless explicitly asked
- When merging, default to **Create a merge commit** unless squash/rebase is requested

## Development

```bash
go test ./...
go build ./cmd/tt
```

See [docs/CLI.md](docs/CLI.md) for command reference.

## Agent rules

See `.cursor/rules/` for git/PR workflow and CLI-via-API constraints.
