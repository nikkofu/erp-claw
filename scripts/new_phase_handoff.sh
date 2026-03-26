#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TEMPLATE_PATH="${ROOT_DIR}/skills/phase-handoff-playbook/handoff-template.md"
OUTPUT_DIR="${ROOT_DIR}/docs/phase-handoff-playbook"

usage() {
	echo "Usage: $0 <topic-slug>"
	echo "Example: $0 phase1-governance-wrapup"
}

if [[ $# -ne 1 ]]; then
	usage
	exit 1
fi

if [[ ! -f "${TEMPLATE_PATH}" ]]; then
	echo "template not found: ${TEMPLATE_PATH}" >&2
	exit 1
fi

raw_topic="$1"
topic_slug="$(printf "%s" "${raw_topic}" | tr '[:upper:]' '[:lower:]' | sed -E 's/[^a-z0-9]+/-/g; s/^-+//; s/-+$//')"

if [[ -z "${topic_slug}" ]]; then
	echo "invalid topic slug: ${raw_topic}" >&2
	exit 1
fi

mkdir -p "${OUTPUT_DIR}"

date_prefix="$(date +%F)"
output_path="${OUTPUT_DIR}/${date_prefix}-${topic_slug}-handoff.md"

if [[ -e "${output_path}" ]]; then
	echo "handoff file already exists: ${output_path}" >&2
	exit 1
fi

cp "${TEMPLATE_PATH}" "${output_path}"

relative_output="${output_path#${ROOT_DIR}/}"

cat <<EOF
Created handoff template:
- ${relative_output}

Next steps:
1. Fill in snapshot, task checklist, and risk sections.
2. Run: ./scripts/phase_handoff_check.sh ${relative_output}
EOF
