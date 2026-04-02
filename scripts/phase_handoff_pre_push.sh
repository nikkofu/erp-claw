#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

upstream_ref="$(git rev-parse --abbrev-ref --symbolic-full-name "@{u}" 2>/dev/null || true)"
if [[ -z "${upstream_ref}" ]]; then
	upstream_ref="origin/main"
fi

git fetch --no-tags origin >/dev/null 2>&1 || true

changed_docs=""
while IFS= read -r file; do
	[[ -z "${file}" ]] && continue
	if [[ "$(basename "${file}")" == "README.md" ]]; then
		continue
	fi
	changed_docs="${changed_docs}${file}"$'\n'
done < <(git diff --name-only "${upstream_ref}...HEAD" -- "docs/phase-handoff-playbook/*.md")

if [[ -z "${changed_docs}" ]]; then
	echo "No changed handoff docs relative to ${upstream_ref}; pre-push handoff check skipped."
	exit 0
fi

count=0
while IFS= read -r doc; do
	[[ -z "${doc}" ]] && continue
	count=$((count + 1))
done <<< "${changed_docs}"

echo "Running handoff checks for ${count} changed doc(s) against ${upstream_ref}:"
while IFS= read -r doc; do
	[[ -z "${doc}" ]] && continue
	echo "- ${doc}"
	./scripts/phase_handoff_check.sh "${doc}"
done <<< "${changed_docs}"
