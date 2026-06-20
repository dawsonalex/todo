#!/usr/bin/env bash
#
# release-notes.sh — generate GitHub Release notes for the current HEAD commit.
#
# Output (to STDOUT):
#   <first line of the HEAD commit>
#
#   Commit: [<short-sha>](<commit-url>)
#
#   ## Changes since <prev-tag>
#   - <subject> ([<short-sha>](<commit-url>))
#   ...
#
# Inputs come from the environment, each with a fallback so the script can be
# run locally to preview notes:
#   REPO_URL  base repo URL (default: derived from the origin remote)
#   PREV_TAG  previous release tag (default: latest v* tag, or v0.0.0)
#   RANGE     git revision range for the changelog (default: PREV_TAG..HEAD)
#   GITHUB_SHA  release commit (default: HEAD)
#
set -euo pipefail

# Base repository URL, e.g. https://github.com/owner/repo
if [ -z "${REPO_URL:-}" ]; then
  origin="$(git config --get remote.origin.url || true)"
  # Normalise git@github.com:owner/repo(.git) and https forms to an https URL.
  origin="${origin%.git}"
  origin="${origin/git@github.com:/https://github.com/}"
  REPO_URL="$origin"
fi

# Previous tag + range, mirroring next-version.sh defaults for standalone runs.
if [ -z "${PREV_TAG:-}" ]; then
  PREV_TAG="$(git tag --list 'v*' --sort=-v:refname | head -n1)"
  [ -z "$PREV_TAG" ] && PREV_TAG="v0.0.0"
fi
if [ -z "${RANGE:-}" ]; then
  if git rev-parse -q --verify "refs/tags/${PREV_TAG}" >/dev/null; then
    RANGE="${PREV_TAG}..HEAD"
  else
    RANGE="HEAD"
  fi
fi

SHA="${GITHUB_SHA:-HEAD}"
SUMMARY="$(git log -1 --format='%s' "$SHA")"
SHORT="$(git rev-parse --short "$SHA")"
COMMIT_URL="${REPO_URL}/commit/$(git rev-parse "$SHA")"

printf '%s\n\n' "$SUMMARY"
printf 'Commit: [%s](%s)\n\n' "$SHORT" "$COMMIT_URL"
printf '## Changes since %s\n' "$PREV_TAG"
git log "$RANGE" --format="- %s ([%h](${REPO_URL}/commit/%H))"
