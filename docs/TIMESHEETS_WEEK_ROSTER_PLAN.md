# `tt timesheets week` / `lastweek` pretty roster

Add human-readable week roster commands to time-tracker-cli: `tt timesheets week` (defaults to this Monday) and `tt timesheets lastweek` (previous Monday), both calling `GET /api/v1/timesheets/weeks/{date}` with pretty tables and optional `--status` filter.

## Locked decisions

- Default for `week`: **this Monday** (`defaultWeekStart()`)
- Convenience alias: **`tt timesheets lastweek`** = previous Monday (`thisMonday - 7 days`)
- Human output via existing `--output pretty` / TTY default; JSON when piped or `--output json`
- Both commands are **`CapRead`** (available in `tt-ro`)
- Optional filter: `--status submitted|draft|all` (default **`all`**)

## Commands

```text
tt timesheets week [--week-start YYYY-MM-DD] [--status submitted|draft|all]
tt timesheets lastweek [--status submitted|draft|all]
```

Examples:

```cmd
tt timesheets week --profile prod
tt timesheets week --week-start 2026-07-06 --status submitted
tt timesheets lastweek --profile dev --status submitted
```

API: `GET /api/v1/timesheets/weeks/{weekStartDate}` (week roster).

## Pretty table shape

Reuse `internal/output` pattern (`PrintTimesheetWeeksList`).

```text
Week 2026-07-06  lock=open

NAME     EMAIL                    STATUS     HOURS  ENTRIES
Corinna  corinna@...              submitted  40.0   5
Leon     leon@...                 draft       8.0   1

Submitted: 12 / 28
```

- Keep API person order (sorted by name server-side)
- Summary line: count of `submitted` vs rows shown (filter-aware denominator)

## Implementation

1. `internal/cmd/timesheets.go` — `week` / `lastweek`, CapRead, shared `runWeekRoster`
2. `internal/cmd/resolve.go` — `lastWeekStart()`
3. `internal/output/output.go` — `PrintWeekRoster(data, statusFilter)`
4. Tests + docs (`ACTIONS.md`, `CLI.md`)

## Out of scope

- CSV export flag
- API changes
- Teaching `tt api` to pretty-print
- Month-close CSVs (stays in time-tracker-reports)
