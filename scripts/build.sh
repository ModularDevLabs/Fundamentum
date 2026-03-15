#!/usr/bin/env bash
set -euo pipefail

NAME=modbot
OUT_DIR=dist
mkdir -p "$OUT_DIR"

echo "Building local binary ./${NAME} for normal dashboard runs..."
CGO_ENABLED=0 go build -o "$NAME" ./cmd/modbot
echo "Built ./${NAME}"

build() {
  local os="$1"
  local arch="$2"
  local outfile="$OUT_DIR/${NAME}_${os}_${arch}"
  if [[ "$os" == "windows" ]]; then
    outfile+=".exe"
  fi
  GOOS="$os" GOARCH="$arch" CGO_ENABLED=0 go build -o "$outfile" ./cmd/modbot
  echo "Built $outfile"
}

build linux amd64
build linux arm64
build darwin amd64
build darwin arm64
build windows amd64
