# Submission + unlock model (agreed)

Canonical product rules for who can edit a timesheet week. Replaces the former global **week lock** / auto-lock / re-lock model (retired).

**Production is live** — see [PRODUCTION_DATA.md](PRODUCTION_DATA.md). This change does **not** rewrite DynamoDB key shapes. Existing `WeekLock` / `WeekUnlockException` rows may remain as unused orphans.

## Agreed decisions

| Topic | Decision |
|-------|----------|
| Edit gate | **Submission only** — `draft` → editable; `submitted` → not |
| Unlock scope | **Person + week** (`personId` + `weekStartDate`), not the person forever |
| Unlock effect | That week’s submission + entries → **draft**; person must **submit again** |
| Who unlocks | Cognito **admins** only |
| Global week lock | **Removed** — no company-wide open/locked week |
| Auto-lock | **Removed** (scheduler disabled / endpoint retired) |
| Re-lock person / re-lock week | **Removed** |
| Approve / bulk-approve | **Removed** (was implemented; locked one person *and* the global week) |
| Week “100% submitted” | Tracked later — **out of this work** |

## Edit rules

| WeekSubmission (person + week) | Can they edit that week? |
|--------------------------------|---------------------------|
| missing or `draft` | **Yes** |
| `submitted` | **No** (admin unlock required) |

Leftover entry status `locked` (from old approve/re-lock) is treated like non-editable until unlock sets the week back to draft. Prefer checking **submission** as the hard API gate.

## Unlock (admin)

`POST /api/v1/timesheets/{personId}/unlock?weekStartDate={monday}`

1. Sets that person’s **WeekSubmission** for that week → `draft`
2. Sets that person’s **TimeEntry** rows in that week → `draft` (clears `submittedAt`)
3. Does **not** write `WeekUnlockException`
4. Does **not** read or write `WeekLock`

## Data implications

| Entity | Keys | After this model |
|--------|------|------------------|
| **WeekSubmission** | `PK=PERSON#…`, `SK=WEEK#{monday}` | Source of truth for editability |
| **TimeEntry** (+ pointer) | unchanged | Status follows submit / unlock for that week |
| **WeekLock** | `PK=WEEK#{monday}`, `SK=LOCK` | Unused — stop writing; leave orphans |
| **WeekUnlockException** | `PK=WEEK#{monday}`, `SK=UNLOCK#{personId}` | Unused — stop writing; leave orphans |

No key-shape migration. No bulk status backfill. Stale `weekLockId` fields ignored.

**Semantic shift:** past weeks that were globally auto-locked while a person stayed `draft` become editable again (intentional — no week lock).

## Removed surfaces

- `POST /timesheets/weeks/lock-prior`
- `POST /timesheets/weeks/{weekStartDate}/lock`
- `POST /timesheets/{personId}/relock`
- `POST /timesheets/{personId}/approve`
- `POST /timesheets/bulk-approve`
- Infra Monday EventBridge week-lock scheduler (disabled / retired)
- Admin UI re-lock person / re-lock week / unlock-exception badge
- CLI `relock`, `lock-week`, approve / lock-prior as product commands

## Kept

- `POST /timesheets/submit`
- `POST /timesheets/{personId}/unlock` (admin, person + week)
- Reject (manager path → draft) if still present — does not involve week lock

## Integrity

Scanner must not require “no drafts in a locked week.” Write-time gate is submission status, not `WeekLock`.

## Related

- Plan / implementation note: this document
- Frontend admin: unlock row action only
- CLI: `tt timesheets unlock`
