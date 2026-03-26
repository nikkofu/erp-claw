#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Backward-compatible alias. Keep historical command working while reusing the new scaffold logic.
exec "${SCRIPT_DIR}/new_phase_handoff.sh" "$@"
