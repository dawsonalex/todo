#!/usr/bin/env bash
#
# next-version.sh — resolve the next semver from commit-message keywords.
#
# Scans commits since the most recent vX.Y.Z tag and picks the highest bump:
#   major: (major), a conventional breaking marker (type!: / )!:), or BREAKING CHANGE
#   minor: (minor), (feat)
#   patch: (patch), (fix), (bug)
#   none:  (chore), (docs), or anything unmatched
#
# Emits machine-readable key=value lines to STDOUT (suitable for appending to
# $GITHUB_OUTPUT); all human-facing diagnostics go to STDERR. Run it directly to
# preview what the next release would be:
#
#   ./scripts/next-version.sh
#
set -euo pipefail

log() { printf '%s\n' "$*" >&2; }

# Most recent semver tag, or v0.0.0 if the repo has never been tagged.
PREV_TAG="$(git tag --list 'v*' --sort=-v:refname | head -n1)"
if [ -z "$PREV_TAG" ]; then
  PREV_TAG="v0.0.0"
  RANGE="HEAD"
else
  RANGE="${PREV_TAG}..HEAD"
fi
log "Previous tag: ${PREV_TAG}"
log "Scanning commits in range: ${RANGE}"

# Highest bump wins: 3=major, 2=minor, 1=patch, 0=none.
bump=0
while IFS= read -r sha; do
  [ -z "$sha" ] && continue
  msg="$(git log -1 --format='%B' "$sha")"

  level=0
  if printf '%s' "$msg" | grep -qiE '\(major\)|BREAKING[ -]CHANGE|\)!:|^[a-z]+!:'; then
    level=3
  elif printf '%s' "$msg" | grep -qiE '\(minor\)|\(feat\)'; then
    level=2
  elif printf '%s' "$msg" | grep -qiE '\(patch\)|\(fix\)|\(bug\)'; then
    level=1
  fi
  # (chore), (docs) and anything unmatched contribute nothing.

  if [ "$level" -gt "$bump" ]; then bump=$level; fi
done < <(git rev-list "$RANGE")

if [ "$bump" -eq 0 ]; then
  log "No (major|minor|patch|feat|fix|bug) commits since ${PREV_TAG} — nothing to release."
  echo "release=false"
  exit 0
fi

# Split PREV_TAG (vMAJOR.MINOR.PATCH) into components.
ver="${PREV_TAG#v}"
MAJOR="${ver%%.*}"; rest="${ver#*.}"
MINOR="${rest%%.*}"; PATCH="${rest#*.}"

case "$bump" in
  3) MAJOR=$((MAJOR + 1)); MINOR=0; PATCH=0 ;;
  2) MINOR=$((MINOR + 1)); PATCH=0 ;;
  1) PATCH=$((PATCH + 1)) ;;
esac
NEW_TAG="v${MAJOR}.${MINOR}.${PATCH}"
log "Bump level ${bump} -> new version ${NEW_TAG}"

echo "release=true"
echo "prev_tag=${PREV_TAG}"
echo "new_tag=${NEW_TAG}"
echo "range=${RANGE}"
