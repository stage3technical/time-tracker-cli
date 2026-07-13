# Auto semver across Time Tracker repos

Canonical plan: [time-tracker-api/docs/plans/auto-semver-release.md](https://github.com/stage3technical/time-tracker-api/blob/main/docs/plans/auto-semver-release.md)

This repo (**time-tracker-cli**) is the **reference implementation** and stays unchanged:

- `scripts/next-semver.sh`
- `.github/workflows/release.yml` (tags + `tt` / `tt-ro` binaries)

Other repos copy the same bump script and a notes-only `release.yml` without touching existing deploy workflows.
