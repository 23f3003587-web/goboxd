#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

GOFLAGS="-buildvcs=false" go test ./tests/... -count=1
GOFLAGS="-buildvcs=false" go test ./tests/integration -count=1
