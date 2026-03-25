#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

export ERP_CLAW_LIVE_SMOKE=1

mkdir -p "${ROOT_DIR}/.cache/go-build"

cd "${ROOT_DIR}"
GOCACHE="${ROOT_DIR}/.cache/go-build" go test ./test/integration -run TestHealthRoutesLive -v
