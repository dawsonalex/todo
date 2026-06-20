#!/usr/bin/env bash
#
# build-release.sh — cross-compile release binaries and package them.
#
# Builds the ./cmd binary for a standard matrix of OS/arch targets, bundles
# each with LICENSE + README into a per-target archive (.tar.gz for Unix,
# .zip for Windows), and writes SHA256SUMS.txt over the archives. Everything
# lands in ./dist, ready to attach to a GitHub Release.
#
# Inputs (env):
#   VERSION  version string used in archive names (default: latest v* tag, or
#            v0.0.0). Also stamped into the binary via -ldflags if the program
#            exposes a `main.version` variable (no-op otherwise).
#
# Usage:
#   VERSION=v1.2.3 ./scripts/build-release.sh
#
set -euo pipefail

BIN_NAME="todo"
PKG="./cmd"
DIST="dist"

VERSION="${VERSION:-$(git tag --list 'v*' --sort=-v:refname | head -n1)}"
[ -z "$VERSION" ] && VERSION="v0.0.0"

# Standard desktop/server target matrix: 64-bit Linux, macOS (Intel + Apple
# Silicon), and Windows. Add lines here to widen coverage.
TARGETS=(
  "linux/amd64"
  "linux/arm64"
  "darwin/amd64"
  "darwin/arm64"
  "windows/amd64"
)

rm -rf "$DIST"
mkdir -p "$DIST"

echo "Building ${BIN_NAME} ${VERSION} for ${#TARGETS[@]} targets..."

for target in "${TARGETS[@]}"; do
  GOOS="${target%/*}"
  GOARCH="${target#*/}"

  bin="$BIN_NAME"
  [ "$GOOS" = "windows" ] && bin="${BIN_NAME}.exe"

  stage="$(mktemp -d)"
  echo "  -> ${GOOS}/${GOARCH}"

  # Static build (CGO off) for portable, dependency-free binaries.
  # -trimpath keeps build paths out of the binary; -s -w strip debug info.
  CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" \
    go build -trimpath \
      -ldflags "-s -w -X main.version=${VERSION}" \
      -o "${stage}/${bin}" "$PKG"

  # Include licence + docs in every archive.
  cp LICENSE README.md "$stage/" 2>/dev/null || true

  archive_base="${BIN_NAME}_${VERSION}_${GOOS}_${GOARCH}"
  if [ "$GOOS" = "windows" ]; then
    ( cd "$stage" && zip -q -r "${OLDPWD}/${DIST}/${archive_base}.zip" . )
  else
    tar -czf "${DIST}/${archive_base}.tar.gz" -C "$stage" .
  fi

  rm -rf "$stage"
done

# Checksums over the produced archives, for download verification.
( cd "$DIST" && sha256sum ./* > SHA256SUMS.txt )

echo "Artifacts:"
ls -1 "$DIST"
