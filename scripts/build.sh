#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

docker compose -f deployments/docker/docker-compose.yml build goboxd
