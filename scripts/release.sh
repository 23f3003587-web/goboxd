#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

if [ -z "${1:-}" ]; then
  echo "Usage: ./scripts/release.sh <tag>" >&2
  exit 1
fi

docker build -t "goboxd:${1}" .
