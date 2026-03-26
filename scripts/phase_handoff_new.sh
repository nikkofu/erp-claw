#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 1 ]]; then
  echo "usage: $0 <topic-kebab-case>"
  exit 1
fi

TOPIC="$1"
DATE_STR="$(date +%F)"
BRANCH="$(git branch --show-current)"
HEAD_SHA="$(git rev-parse --short HEAD)"
ORIGIN_MAIN_SHA="$(git rev-parse --short origin/main 2>/dev/null || echo "unknown")"
ORIGIN_BRANCH_SHA="$(git rev-parse --short "origin/${BRANCH}" 2>/dev/null || echo "unknown")"
WORKTREE_PATH="$(pwd)"
OUT_DIR="docs/phase-handoff-playbook"
OUT_PATH="${OUT_DIR}/${DATE_STR}-${TOPIC}-handoff.md"

mkdir -p "${OUT_DIR}"

if [[ -f "${OUT_PATH}" ]]; then
  echo "file already exists: ${OUT_PATH}"
  exit 1
fi

cat > "${OUT_PATH}" <<EOF
# ${TOPIC} Handoff (${DATE_STR})

## 1. Snapshot

- worktree_path: \`${WORKTREE_PATH}\`
- branch: \`${BRANCH}\`
- head_sha: \`${HEAD_SHA}\`
- origin_main_sha: \`${ORIGIN_MAIN_SHA}\`
- origin_phase_branch_sha: \`${ORIGIN_BRANCH_SHA}\`
- open_pr:
- release_tag:
- merge_safety: \`safe_to_merge\` | \`blocked\` (+ reason)

## 2. Fresh Verification

| Command | Result | Notes |
| --- | --- | --- |
| \`go test ./...\` | PASS/FAIL |  |

## 3. Completed Today

- item 1

## 4. Still Unfinished

- item 1

## 5. Tomorrow Kickoff

\`\`\`bash
git fetch origin --prune
git checkout ${BRANCH}
git pull --ff-only origin ${BRANCH}
go test ./...
\`\`\`

First execution task tomorrow:

- <one smallest task>

## 6. Smallest-Next-Task Checklist

| ID | Priority | Task | Why | Acceptance | File Scope | Dependency | Parallel-safe |
| --- | --- | --- | --- | --- | --- | --- | --- |
| T01 | P0 |  |  |  |  | none | yes/no |

## 7. Risks

- risk 1

## 8. Do Not Do Next

- trap 1

## 9. Reflection

- what worked:
- what slowed down:
- one process improvement for next session:
EOF

echo "created: ${OUT_PATH}"
