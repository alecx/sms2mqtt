#!/usr/bin/env bash
# Cross-compile the per-arch service binaries shipped inside the add-on. Run this
# before copying the add-on to the Yellow (local install) or before committing
# for the GitHub-repo on-device build.
set -euo pipefail
cd "$(dirname "$0")"

OUT=sms2mqtt/bin
mkdir -p "$OUT"

build() { # goarch  suffix
  echo "building $2 ..."
  GOOS=linux GOARCH="$1" go build -trimpath -ldflags="-s -w" -o "$OUT/sms2mqtt.$2" ./cmd/sms2mqtt
}

build arm64 aarch64
build amd64 amd64

ls -lh "$OUT"
