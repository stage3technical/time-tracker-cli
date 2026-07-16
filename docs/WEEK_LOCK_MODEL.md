# Week lock model (agreed)

Canonical product rules for who can edit a timesheet week. Source of truth for API, admin UI, and CLI.

## Two layers (do not conflate)

| Layer | What it is | What it controls |
|-------|------------|------------------|
| **`WeekLock`** | One record per Monday week: `open` or `locked` | **Company-wide editability** for that week |
| **Person / entry status** | Per person: submission `draft` \| `submitted`; entries `draft` \| `submitted` \| `locked` | Bookkeeping / UI chips; **not** a second week lock |

The API’s hard gate on create / update / delete is **only** “is this week open?” (`is_week_open` / PR-19). Entry status alone does not block edits once the week is open.

## Edit rules (agreed)

| `WeekLock` | Person submission | Can they edit that week? |
|------------|-------------------|---------------------------|
| `locked` | draft (or anything) | **No** |
| `open` | draft | **Yes** |
| `open` | submitted | My Week UI: **no** (still submitted). API does not use submission as the week gate. |

So: **global week locked ⇒ nobody edits**, including drafts. **Week open + draft ⇒ editable.**

## How a week becomes globally locked

Same `WeekLock` switch in all cases:

1. **Auto-lock (ops)** — Monday-morning scheduler / `POST /timesheets/weeks/lock-prior` locks the **prior** Monday week. This is the intended deadline close. Same as “global week locked” for that prior week.
2. **Admin lock week** — `POST /timesheets/weeks/{weekStartDate}/lock` (admin UI **Re-lock week**, `tt timesheets lock-week`). Recovery when the week is open.
3. **Admin re-lock person (last exception)** — clearing the last `WeekUnlockException` via relock restores the global week lock.
4. **Approve (legacy / lame)** — see below.

## Auto-lock

Auto-lock is fine and intentional: it sets **global** `WeekLock` for the prior Monday. Not “manager approved this person.”

## Unlock (admin) — agreed behavior

`POST /timesheets/{personId}/unlock`:

1. Puts **that person** back to editable draft (submission + their entries).
2. If the week was globally locked → **opens the week** (`WeekLock` → `open`) so they can edit (PR-19).
3. Records a **`WeekUnlockException`** for that person/week (bookmark: “this person was admin-unlocked”).

Side effect of (2): while the week is open, the company-wide edit gate is open. Other people’s row statuses are left alone (they may still show submitted/locked on entries), but the hard API gate is week open/closed, not their entry status.

Unlock is the right admin tool for “let this person fix their sheet.” Opening the week is required for them to edit under the current gate.

## Re-lock person / re-lock week (exception model)

See [ADMIN_PERSON_RELOCK_PLAN.md](ADMIN_PERSON_RELOCK_PLAN.md).

- **Re-lock person** — clear that person’s exception; set their entries to `locked` (no resubmit). If no exceptions remain → restore global week lock.
- **Re-lock week** — admin sets global `WeekLock` only (recovery when open with no exception rows).

## Approve — lame; candidate to remove

`POST /timesheets/{personId}/approve` today:

- Requires that person submitted
- Sets **their** entries → `locked`
- **Also** sets **global** `WeekLock` → `locked`

Approving one person closing the week for **everyone** does not make sense as product behavior. It is a historical coupling, not the intended “lock only this user” action.

**Status:** still implemented (API / CLI / OpenAPI) but **not** the recommended way to close a week. Prefer:

- **Auto-lock** for the deadline close
- **Admin unlock → correct → re-lock person** (and last exception / **lock week**) for admin corrections

**Likely follow-up:** remove or gut approve (and bulk-approve) until a real per-person approve exists that does **not** flip global `WeekLock`. Do not teach approve as “lock this user’s timesheet.”

## Submit / reject (for context)

- **Submit** — person marks their week submitted (entries → submitted). Does not lock the week.
- **Reject** — manager path: back to draft for that person. Does not change global `WeekLock`.

## Related docs

- [ADMIN_PERSON_RELOCK_PLAN.md](ADMIN_PERSON_RELOCK_PLAN.md) — exception + re-lock implementation plan
- Frontend admin UI: `time-tracker-frontend-01/docs/ADMIN_TIMESHEETS_SCREEN_PLAN.md`
- CLI actions: [ACTIONS.md](ACTIONS.md)
