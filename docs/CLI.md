# `tt` command reference

Cross-platform CLI for the [BLVD Timesheet API](https://github.com/stage3technical/time-tracker-api).

See also: [ACTIONS.md](ACTIONS.md) (commands by API group), [PLAN.md](PLAN.md) (design), [README](../README.md) (install).

## Install

```bash
go install github.com/stage3technical/time-tracker-cli/cmd/tt@latest
```

Or build from source (embeds git tag/commit):

```bash
./scripts/build.sh          # Linux / macOS / WSL
```

```powershell
.\scripts\build.ps1         # Windows
```

Plain `go build -o tt.exe ./cmd/tt` shows `dev (commit none, built unknown)`.

**Version:** `tt version` or `tt --version`. Merges to `main` auto-tag **semver** releases (`v0.1.0`, …) via GitHub Actions — bump rules: `feat:` → minor, breaking/`!` → major, else patch.

## Configuration

### Interactive setup

```bash
tt configure
```

Prompts for profile name, API base URL, and JWT. Writes `~/.tt/config` (Windows: `%USERPROFILE%\.tt\config`).

### Non-interactive

```bash
tt configure set --profile dev \
  --base-url https://8igr6pspqh.execute-api.us-east-1.amazonaws.com \
  --token "<JWT>"
```

### List profiles

```bash
tt configure list
```

Tokens are masked in output. `*` marks the default profile.

### Overrides

| Source | Variable / flag |
|--------|-----------------|
| Flag | `--profile`, `--base-url`, `--token` |
| Env | `TT_PROFILE`, `TT_API_BASE_URL`, `TT_API_TOKEN` |
| File | `~/.tt/config` |

Resolution order: **flags → env → config file**.

### Auth note

JWTs expire. When you see `401`, paste a fresh token from your browser session (Network tab → any API call → `Authorization` header) and run `tt configure set --profile dev --token "<new-jwt>"`.

## Global flags

| Flag | Description |
|------|-------------|
| `--profile` | Config profile name |
| `--base-url` | Override API base URL |
| `--token` | Override JWT |
| `--output json\|pretty` | Output format (default: pretty on TTY, json when piped) |
| `--quiet` | Suppress non-essential stderr |

## Commands

### `tt version`

Print build version (no config required).

```bash
tt version
tt --version
```

`tt version` includes commit and build date when compiled with release `-ldflags`.

### `tt health`

`GET /health` — no auth required.

```bash
tt health
tt health --base-url https://8igr6pspqh.execute-api.us-east-1.amazonaws.com
```

### `tt me`

`GET /me` — current user from JWT.

```bash
tt me --profile dev
```

### `tt api`

Generic HTTP call for endpoints not yet wrapped.

```bash
tt api GET /api/v1/persons
tt api GET /api/v1/persons --query status=active --query type=W2
tt api PUT /api/v1/persons/UUID/manager --body '{"managerId":"..."}'
tt api POST /api/v1/persons/import --query onDuplicate=update --body @person.json
```

Body: inline JSON or `@file.json`.

### `tt persons list`

`GET /api/v1/persons`

```bash
tt persons list
tt persons list --status active --type W2
tt persons list --output json
```

Pretty mode columns: `ID`, `NAME`, `EMAIL`, `ROLE`, `TEAM`.

### `tt persons get`

`GET /api/v1/persons/{id}`

```bash
tt persons get a091a3d5-f18a-4071-8ad0-454d9fe61cde
```

### `tt persons update`

`PUT /api/v1/persons/{id}`

```bash
tt persons update PERSON_ID \
  --name "Nicholaus Chipping" \
  --email nicholaus.chipping@blvdinteractive.com \
  --primary-role "AEM Architect" \
  --employment-type W2 \
  --team Engineering
```

At least one field flag is required.

### `tt persons import`

`POST /api/v1/persons/import?onDuplicate=update|skip|fail`

```bash
tt persons import --file person.json --on-duplicate update
```

`person.json` example:

```json
{
  "name": "Jane Doe",
  "email": "jane.doe@example.com",
  "primaryRole": "Analyst",
  "employmentType": "W2",
  "team": "Delivery"
}
```

### `tt persons manager get`

`GET /api/v1/persons/{id}/manager`

```bash
tt persons manager get PERSON_ID
```

### `tt persons manager set`

`PUT /api/v1/persons/{id}/manager`

```bash
tt persons manager set PERSON_ID --manager-id MANAGER_UUID
```

### `tt persons subordinates list`

`GET /api/v1/persons/{id}/subordinates`

```bash
tt persons subordinates list MANAGER_ID
```

### `tt timesheets`

Timesheet workflow — see **[ACTIONS.md](ACTIONS.md)** § Advanced Workflow.

`--week-start` defaults to **this Monday** (local timezone) when omitted.

```bash
tt timesheets list --email marlene.bockler@blvdinteractive.com
tt timesheets list --email marlene.bockler@blvdinteractive.com --before 2026-07-06
tt timesheets get --email pam@blvdinteractive.com
tt timesheets unlock --profile prod --email pam@blvdinteractive.com
tt timesheets purge --profile prod --email marlene.bockler@blvdinteractive.com --week-start 2026-06-30 --confirm
tt timesheets purge --profile prod --email marlene.bockler@blvdinteractive.com --before 2026-07-06 --confirm
tt timesheets submit --person-id UUID --week-start 2026-07-06
tt timesheets approve --person-id UUID
tt timesheets reject --email user@blvdinteractive.com
```

Pretty mode columns for `list`: `WEEK_START`, `ENTRIES`, `HOURS`, `SUBMISSION`, `WEEK_LOCK`.

### `tt entries`

Time reporting entries — see **[ACTIONS.md](ACTIONS.md)** § Time Reporting.

```bash
tt entries list --email user@blvdinteractive.com
tt entries list --email user@blvdinteractive.com --work-date 2026-07-07
tt entries get ENTRY_ID
tt entries create --email user@blvdinteractive.com \
  --project-name OOO --work-date 2026-07-07 --role PM --hours 8
tt entries update ENTRY_ID --hours 4
tt entries delete ENTRY_ID --confirm
```

Pretty mode columns for `list`: `ID`, `WORK_DATE`, `PROJECT`, `ROLE`, `HOURS`, `STATUS`.

### `tt projects`

Project CRUD — see **[ACTIONS.md](ACTIONS.md)** § Projects.

Projects use `canonicalName` as the display name. `--code` is an exact match on `canonicalName` (there is no separate code field).

```bash
tt projects list
tt projects list --status active
tt projects get PROJECT_ID
tt projects get --name "OOO/Holiday"
tt projects get --code OOO
tt projects create --name OOO --bill-type N-BIL-I --allowed-roles "PM,AEM Architect"
tt projects update PROJECT_ID --name "OOO (updated)"
tt projects archive PROJECT_ID --confirm
tt projects archive --name "OOO/Holiday" --confirm
```

Pretty mode columns for `list`: `ID`, `NAME`, `BILL_TYPE`, `STATUS`.

Destructive commands (`archive`, `entries delete`) require `--confirm`.

### `tt company-roles`

Company role registry — see **[ACTIONS.md](ACTIONS.md)** § Company Roles.

```bash
tt company-roles list
tt company-roles list --profile prod
tt company-roles get ROLE_ID
tt company-roles create --name "AEM Architect" --description "..."
tt company-roles update ROLE_ID --name "AEM Architect"
tt company-roles delete ROLE_ID --confirm
```

Pretty mode columns for `list`: `ID`, `NAME`, `DESCRIPTION`.

### `tt api`

Generic HTTP call for endpoints not yet wrapped (escape hatch).

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General / 5xx error |
| 2 | 400 Bad Request |
| 4 | 404 Not Found |
| 5 | 401 Unauthorized |
| 6 | 403 Forbidden |
| 9 | 409 Conflict |

## Scripting

Pipe JSON to `jq`:

```bash
tt persons list --output json | jq '.[].email'
tt persons get ID --output json | jq -r .name
```

Use env vars in CI (never commit tokens):

```bash
export TT_API_BASE_URL=https://...
export TT_API_TOKEN=eyJ...
tt me
```

## Windows notes

- Config path: `%USERPROFILE%\.tt\config`
- Build: `go build -o tt.exe ./cmd/tt`
- PowerShell env: `$env:TT_API_TOKEN = "eyJ..."`
