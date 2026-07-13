# CLI actions by API group

Commands call the **BLVD Timesheet HTTP API** only — no direct DynamoDB. Groupings match the API Swagger tags ([time-tracker-api `v1_stub_tags.py`](https://github.com/stage3technical/time-tracker-api/blob/develop/src/time_tracker_api/routes/v1_stub_tags.py)).

Legend: **implemented** = first-class `tt` subcommand · **api** = use `tt api` until wrapped · **—** = not in API yet

**Conventions**

- `--week-start` defaults to **this Monday** (local timezone) on `tt timesheets` and `tt entries list`.
- Destructive commands require **`--confirm`**: `tt entries delete`, `tt projects archive`.
- Projects have no separate `code` field; `--code` matches `canonicalName` exactly (case-insensitive).

---

## Setup (not an API tag)

| Action | Command | API |
|--------|---------|-----|
| Configure profile | `tt configure` / `tt configure set` | — |
| List profiles | `tt configure list` | — |
| Health check | `tt health` | `GET /health` |
| Current user | `tt me` | `GET /me` |

---

## Persons

| Action | Status | Command |
|--------|--------|---------|
| List persons | **implemented** | `tt persons list [--status active] [--type W2]` |
| Get person | **implemented** | `tt persons get PERSON_ID` |
| Update person | **implemented** | `tt persons update PERSON_ID --name ...` |
| Import person | **implemented** | `tt persons import --file person.json` |
| Get manager | **implemented** | `tt persons manager get PERSON_ID` |
| Set manager | **implemented** | `tt persons manager set PERSON_ID --manager-id UUID` |
| List subordinates | **implemented** | `tt persons subordinates list MANAGER_ID` |
| Create person | api | `tt api POST /api/v1/persons --body '{...}'` |
| Deactivate person | api | `tt api DELETE /api/v1/persons/UUID` |

---

## Relationships

| Action | Status | Command |
|--------|--------|---------|
| CRUD employee relationships | api | `tt api POST /api/v1/employee-relationships/two-way --body '{...}'` |

---

## Projects

| Action | Status | Command |
|--------|--------|---------|
| List projects | **implemented** | `tt projects list [--status active]` |
| Get project | **implemented** | `tt projects get ID` or `tt projects get --name "OOO"` / `--code OOO` |
| Create project | **implemented** | `tt projects create --name OOO --bill-type N-BIL-I [--allowed-roles ...]` |
| Update project | **implemented** | `tt projects update ID --name ...` |
| Archive project | **implemented** | `tt projects archive ID --confirm` or `--name` / `--code` |
| Project roles, managers, approvers | api | `tt api ...` under `/api/v1/projects/{id}/...` |

---

## Company Roles

| Action | Status | Command |
|--------|--------|---------|
| List roles | **implemented** | `tt company-roles list` |
| Get role | **implemented** | `tt company-roles get ROLE_ID` |
| Create role | **implemented** | `tt company-roles create --name "AEM Architect"` |
| Update role | **implemented** | `tt company-roles update ROLE_ID --description "..."` |
| Delete role | **implemented** | `tt company-roles delete ROLE_ID --confirm` |

---

## Documentation

| Action | Status | Command |
|--------|--------|---------|
| CRUD explanation documents | api | `tt api POST /api/v1/documentation/explanation --body '{...}'` |

---

## Time Reporting (`tt entries`)

| Action | Status | Command |
|--------|--------|---------|
| List entries | **implemented** | `tt entries list --email user@...` (defaults to this week) |
| Get entry | **implemented** | `tt entries get ENTRY_ID` |
| Create entry | **implemented** | `tt entries create --email ... --project-name OOO --work-date 2026-07-07 --role PM --hours 8` |
| Update entry | **implemented** | `tt entries update ENTRY_ID --hours 4` |
| Delete entry | **implemented** | `tt entries delete ENTRY_ID --confirm` |

---

## Accounts

| Action | Status | Command |
|--------|--------|---------|
| CRUD accounts | api | `tt api GET /api/v1/accounts` etc. |

---

## Advanced Workflow (timesheets)

Week start is always **Monday** (`YYYY-MM-DD`). Use `--person-id` or `--email` on every command. Omit `--week-start` to use **this Monday**.

| Action | Status | Command |
|--------|--------|---------|
| List weeks | **implemented** | `tt timesheets list --email user@blvdinteractive.com` |
| Get week | **implemented** | `tt timesheets get --email user@blvdinteractive.com` |
| Submit week | **implemented** | `tt timesheets submit --email ...` |
| Approve (lock) week | **implemented** | `tt timesheets approve --person-id UUID` |
| Reject submission | **implemented** | `tt timesheets reject --email ...` |
| **Unlock (admin)** | **implemented** | `tt timesheets unlock --email ...` |
| **Lock prior week (ops)** | **implemented** | `TT_SCHEDULER_SECRET=... tt timesheets lock-prior --profile prod` |
| **Purge (admin)** | **implemented** | `tt timesheets purge --email ... --week-start 2026-06-30 --confirm` |
| Bulk approve | api | `tt api POST /api/v1/timesheets/bulk-approve --body '{...}'` |

`list` supports `--before` / `--after` (Monday `YYYY-MM-DD`). `purge` supports `--week-start` (one week) or `--before` (all prior weeks, exclusive). Purge requires `--confirm`.

### Admin unlock notes

- Reverts the **target person's** entries and submission to `draft`.
- If the **week is globally locked**, unlock **opens the week for everyone** (week locks are not per-person).
- Requires API with `POST /api/v1/timesheets/{personId}/unlock` deployed.

### Admin purge notes

- Deletes **all entries** and **WeekSubmission** for the person/week.
- Does **not** change global `WeekLock` (other people on that week are unaffected).
- Works on locked/submitted weeks without `unlock`.

---

## Example: clear Marlene's history before this week

```powershell
# See what exists
tt timesheets list --profile prod --email marlene.bockler@blvdinteractive.com --before 2026-07-06

# One week at a time
tt timesheets purge --profile prod --email marlene.bockler@blvdinteractive.com --week-start 2026-06-30 --confirm

# Or all weeks before this Monday (exclusive)
tt timesheets purge --profile prod --email marlene.bockler@blvdinteractive.com --before 2026-07-06 --confirm
```

---

## Reporting & Analytics · Task Management · Real-Time Tracking · Billing · Resource & Leave · Audit & Compliance

| Action | Status | Command |
|--------|--------|---------|
| All sheet routes | stub only | API returns mock JSON; no `tt` wrappers yet |

---

## Example: OOO/Holiday project split

```powershell
tt projects list --status active
tt projects create --name OOO --bill-type N-BIL-I --allowed-roles "PM,AEM Architect"
tt projects create --name Holiday --bill-type N-BIL-I --allowed-roles "PM,AEM Architect"
tt projects archive --name "OOO/Holiday" --confirm

tt entries list --email user@blvdinteractive.com
tt entries update ENTRY_ID --project-id <new-ooo-id>
```

---

## Example: unlock Pam's timesheet this week

```powershell
tt configure set --profile prod `
  --base-url https://timeapi.blvdinteractive.com `
  --token "<JWT>"

tt timesheets get --profile prod --email pam@blvdinteractive.com
tt timesheets unlock --profile prod --email pam@blvdinteractive.com
```

---

## Escape hatch

For endpoints not yet wrapped:

```powershell
tt api METHOD PATH [--query key=value] [--body @file.json]
```

---

## See also

- [CLI.md](CLI.md) — install, config, global flags
- [PLAN.md](PLAN.md) — roadmap
- [api-implementation-status.md](https://github.com/stage3technical/time-tracker-api/blob/develop/docs/api-spec/api-implementation-status.md) — which API routes are real vs stub
