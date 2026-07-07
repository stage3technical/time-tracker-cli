#!/usr/bin/env bash
# Compute next semver tag from conventional commits since the latest v* tag.
# Outputs the new tag (e.g. v0.2.0) on stdout.
set -euo pipefail

last_tag() {
  git describe --tags --match 'v*' --abbrev=0 2>/dev/null || echo ""
}

bump=patch
if last=$(last_tag); [[ -n "$last" ]]; then
  range="${last}..HEAD"
else
  range="HEAD"
  echo "v0.1.0"
  exit 0
fi

mapfile -t subjects < <(git log "$range" --pretty=format:%s)

for s in "${subjects[@]}"; do
  if [[ "$s" =~ ^[a-z]+(\(.+\))?!: ]] || [[ "$s" == *"BREAKING CHANGE"* ]]; then
    bump=major
    break
  fi
  if [[ "$s" =~ ^feat(\(.+\))?: ]]; then
    if [[ "$bump" != "major" ]]; then
      bump=minor
    fi
  fi
done

ver="${last#v}"
IFS=. read -r major minor patch <<<"$ver"
major=${major:-0}
minor=${minor:-0}
patch=${patch:-0}

case "$bump" in
  major)
    major=$((major + 1))
    minor=0
    patch=0
    ;;
  minor)
    minor=$((minor + 1))
    patch=0
    ;;
  patch)
    patch=$((patch + 1))
    ;;
esac

echo "v${major}.${minor}.${patch}"
