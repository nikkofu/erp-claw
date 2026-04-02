#!/usr/bin/env bash
set -euo pipefail

HANDOFF_DIR="docs/phase-handoff-playbook"

if [[ ! -d "${HANDOFF_DIR}" ]]; then
  echo "handoff dir not found: ${HANDOFF_DIR}"
  exit 1
fi

LATEST_FILE="$(ls -1 "${HANDOFF_DIR}"/20??-??-??-*-handoff.md 2>/dev/null | sort | tail -n 1 || true)"
if [[ -z "${LATEST_FILE}" ]]; then
  echo "no dated handoff file found under ${HANDOFF_DIR}"
  exit 1
fi

echo "latest_handoff=${LATEST_FILE}"
echo "worktree=$(pwd)"
echo "branch=$(git branch --show-current)"
echo "head_sha=$(git rev-parse --short HEAD)"
echo "origin_main_sha=$(git rev-parse --short origin/main 2>/dev/null || echo unknown)"
echo "origin_branch_sha=$(git rev-parse --short "origin/$(git branch --show-current)" 2>/dev/null || echo unknown)"
echo
echo "---- first execution task (from handoff) ----"
awk '
  /First execution task tomorrow:/ {found=1; next}
  found && /^- / {print; exit}
' "${LATEST_FILE}"
echo
echo "---- verification commands (from handoff table, best-effort) ----"
awk -F"|" '
  /^\| `/ {
    cmd=$2
    gsub(/^ +| +$/, "", cmd)
    print cmd
  }
' "${LATEST_FILE}" | sed 's/^`//; s/`$//' || true
