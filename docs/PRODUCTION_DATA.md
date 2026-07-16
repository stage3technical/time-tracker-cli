# Production data — do not casually rewrite the store

**BLVD Timesheet is live in production.** Dev, UAT, and Prod DynamoDB tables already hold data people rely on (entries, persons, projects, week locks, submissions).

## Rules for agents and PRs

1. **No silent key migrations.** Do not change `PK` / `SK` prefixes, GSI layouts, or entity key builders (`keys.py`) unless the change is the explicit purpose of the PR, with a migration plan and human approval.
2. **Doc regenerators are not schema changes.** Running `scripts/generate_data_model_docs.py` may refresh `entities.md` / `erd.mmd` from `entities.yaml`. That can **surface fields that already existed in yaml** (e.g. Project `shortcode`) in the ERD diff. That is **not** “adding shortcodes to prod.” Do not treat ERD catch-up as a product feature drop.
3. **Integrity / workflow PRs stay in their lane.** Unlock/re-lock / integrity-rule work must not “clean up” unrelated key-doc drift or invent new stored fields.
4. **When docs disagree with `keys.py`:** call it out; do not quietly rewrite yaml to match code (or code to match yaml) inside an unrelated PR. Live writes follow **code**. Prod rows already use whatever `keys.py` wrote historically.

## TimeEntry pointer SK (`ENTRY#` vs `TIMEENTRY#`) — clarification

- **Live write path** (`keys.time_entry_pointer_keys`): SK = `ENTRY#{workDate}#{entryId}`.
- **Profile row**: PK = `TIMEENTRY#{id}` / SK = `PROFILE`.
- Older `entities.yaml` text for the **pointer** SK said `TIMEENTRY#{workDate}#{…}` — that was **documentation drift**, not a second live format.
- **Nothing in the unlock/re-lock work migrates pointer rows.** Do not change pointer prefixes in prod without an explicit migration.

## Project shortcode

`Project.shortcode` / integrity rule `unique_project_shortcode` were **already** in `entities.yaml` / integrity rules before unlock/re-lock. Regenerating the ERD may show `shortcode` on the Project box if the diagram was stale. That is **not** a new prod feature introduced by unlock/re-lock.

## Related

- [WEEK_LOCK_MODEL.md](WEEK_LOCK_MODEL.md) — week lock / unlock product rules
- [data-model/entities.yaml](data-model/entities.yaml) — schema source of truth (docs)
- `src/time_tracker_api/db/keys.py` — what the API actually writes
