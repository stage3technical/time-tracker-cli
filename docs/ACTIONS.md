# CLI actions by API group

Commands call the **BLVD Timesheet HTTP API** only — no direct DynamoDB. Groupings match the API Swagger tags ([time-tracker-api `v1_stub_tags.py`](https://github.com/stage3technical/time-tracker-api/blob/develop/src/time_tracker_api/routes/v1_stub_tags.py)).

Legend: **implemented** = first-class `tt` subcommand · **api** = use `tt api` until wrapped · **—** = not in API yet

---

## Setup (not an API tag)

| Action | Command | API |
|--------|---------|-----|
| Configure profile | `tt configure` / `tt configure set` | — |
| List profiles | `tt configure list` | — |
| Health check | `tt health` | `GET /health` |
| Current user | `tt me` | `GET /me` |
| Generic HTTP | `tt api METHOD PATH` | any |

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
| List / get / create / update / delete | api | `tt api GET /api/v1/projects` etc. |
| Project roles, managers, approvers | api | `tt api ...` under `/api/v1/projects/{id}/...` |

---

## Company Roles

| Action | Status | Command |
|--------|--------|---------|
| CRUD company roles | api | `tt api GET /api/v1/company-roles` etc. |

---

## Documentation

| Action | Status | Command |
|--------|--------|---------|
| CRUD explanation documents | api | `tt api POST /api/v1/documentation/explanation --body '{...}'` |

---

## Time Reporting

| Action | Status | Command |
|--------|--------|---------|
| List entries | api | `tt api GET "/api/v1/time-reporting/entries?personId=UUID&weekStartDate=2026-07-06"` |
| Create entry | api | `tt api POST /api/v1/time-reporting/entries --body '{...}'` |
| Update entry | api | `tt api PUT /api/v1/time-reporting/entries/ENTRY_ID --body '{...}'` |
| Delete entry | api | `tt api DELETE /api/v1/time-reporting/entries/ENTRY_ID` |

---

## Accounts

| Action | Status | Command |
|--------|--------|---------|
| CRUD accounts | api | `tt api GET /api/v1/accounts` etc. |

---

## Advanced Workflow (timesheets)

Week start is always **Monday** (`YYYY-MM-DD`). Use `--person-id` or `--email` on every command.

| Action | Status | Command |
|--------|--------|---------|
| Get week | **implemented** | `tt timesheets get --email user@blvdinteractive.com --week-start 2026-07-06` |
| Submit week | **implemented** | `tt timesheets submit --email ... --week-start 2026-07-06` |
| Approve (lock) week | **implemented** | `tt timesheets approve --person-id UUID --week-start 2026-07-06` |
| Reject submission | **implemented** | `tt timesheets reject --email ... --week-start 2026-07-06` |
| **Unlock (admin)** | **implemented** | `tt timesheets unlock --email ... --week-start 2026-07-06` |
| Bulk approve | api | `tt api POST /api/v1/timesheets/bulk-approve --body '{...}'` |

### Admin unlock notes

- Reverts the **target person's** entries and submission to `draft`.
- If the **week is globally locked**, unlock **opens the week for everyone** (week locks are not per-person).
- Requires API with `POST /api/v1/timesheets/{personId}/unlock` deployed.

---

## Reporting & Analytics · Task Management · Real-Time Tracking · Billing · Resource & Leave · Audit & Compliance

| Action | Status | Command |
|--------|--------|---------|
| All sheet routes | stub only | API returns mock JSON; no `tt` wrappers yet |

---

## Prod example: unlock Pam's timesheet this week

Today is **2026-07-07** (Tuesday) → week starts **2026-07-06** (Monday).

```powershell
# One-time prod profile (JWT from prod login — localStorage authToken)
tt configure set --profile prod `
  --base-url https://timeapi.blvdinteractive.com `
  --token "<JWT>"

# Inspect current state
tt timesheets get --profile prod --email pam@blvdinteractive.com --week-start 2026-07-06

# Unlock
tt timesheets unlock --profile prod --email pam@blvdinteractive.com --week-start 2026-07-06
```

If unlock is not deployed yet, **submitted-only** (week not locked) can use reject today:

```powershell
tt timesheets reject --profile prod --email pam@blvdinteractive.com --week-start 2026-07-06
```

Or via generic API (same as above once `tt timesheets` is built):

```powershell
tt api POST "/api/v1/timesheets/PERSON_UUID/unlock?weekStartDate=2026-07-06" --profile prod
```

---

## See also

- [CLI.md](CLI.md) — install, config, global flags
- [PLAN.md](PLAN.md) — roadmap
- [api-implementation-status.md](https://github.com/stage3technical/time-tracker-api/blob/develop/docs/api-spec/api-implementation-status.md) — which API routes are real vs stub
