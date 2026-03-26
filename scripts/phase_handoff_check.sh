#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

usage() {
	echo "Usage: $0 <handoff-doc-path>"
	echo "Example: $0 docs/phase-handoff-playbook/2026-03-26-phase1-handoff.md"
}

if [[ $# -ne 1 ]]; then
	usage
	exit 1
fi

doc_arg="$1"
if [[ "${doc_arg}" = /* ]]; then
	doc_path="${doc_arg}"
else
	doc_path="$(pwd)/${doc_arg}"
fi

if [[ ! -f "${doc_path}" ]]; then
	echo "handoff doc not found: ${doc_path}" >&2
	exit 1
fi

declare -a errors=()
declare -a warnings=()

add_error() {
	errors+=("$1")
}

add_warning() {
	warnings+=("$1")
}

require_heading() {
	local heading="$1"
	if ! rg -q "^## ${heading}$" "${doc_path}"; then
		add_error "missing section heading: ## ${heading}"
	fi
}

require_field_value() {
	local label="$1"
	if ! awk -v label="${label}" '
function has_text(line) {
	gsub(/^[[:space:]]+|[[:space:]]+$/, "", line)
	return line != ""
}
BEGIN {
	found = 0
	has_value = 0
	in_field = 0
}
{
	if ($0 ~ "^## " && in_field == 1) {
		in_field = 0
	}

	if ($0 ~ ("^- " label ":")) {
		found = 1
		in_field = 1
		line = $0
		sub("^- " label ":[[:space:]]*", "", line)
		if (has_text(line)) {
			has_value = 1
		}
		next
	}

	if (in_field == 1) {
		if ($0 ~ "^- [^:]+:") {
			in_field = 0
			next
		}
		if ($0 ~ /^[[:space:]]*-[[:space:]]+[^[:space:]].*$/ || $0 ~ /^[[:space:]]*[0-9]+\.[[:space:]]+[^[:space:]].*$/ || $0 ~ /^[[:space:]]+[^[:space:]].*$/) {
			has_value = 1
		}
	}
}
END {
	if (found == 1 && has_value == 1) {
		exit 0
	}
	exit 1
}
' "${doc_path}"; then
		add_error "missing or empty field: ${label}"
	fi
}

require_heading "Snapshot"
require_heading "Tomorrow Kickoff"
require_heading "Minimal Task Checklist"
require_heading "Experience Summary"
require_heading "Deep Reflection"
require_heading "Do Not Do Tomorrow"

require_field_value "worktree path"
require_field_value "branch"
require_field_value "head sha"
require_field_value "origin/main sha"
require_field_value "origin/phase branch sha"
require_field_value "merge safety status"
require_field_value "fresh verification command"
require_field_value "fresh verification result"
require_field_value "today completed"
require_field_value "still unfinished"

if ! rg -q "^[0-9]+\\.[[:space:]]+" "${doc_path}"; then
	add_error "tomorrow kickoff has no numbered steps"
fi

task_row_count="$(
	awk -F'|' '
function trim(x) {
	gsub(/^[ \t]+|[ \t]+$/, "", x)
	return x
}
/^\|/ {
	if ($0 ~ /^\|[[:space:]]*ID[[:space:]]*\|/) next
	if ($0 ~ /^\|[[:space:]-]+\|/) next
	if (NF < 9) next

	id=trim($2)
	priority=trim($3)
	task=trim($4)
	why=trim($5)
	acceptance=trim($6)
	files=trim($7)
	dependency=trim($8)
	parallel=trim($9)

	if (id != "" && priority != "" && task != "" && why != "" && acceptance != "" && files != "" && dependency != "" && parallel != "") {
		complete++
	}
}
END {
	print complete + 0
}
' "${doc_path}"
)"

if [[ "${task_row_count}" -lt 1 ]]; then
	add_error "task checklist has no fully specified task row"
fi

if rg -q "<phase or topic>|T00-01 \\| P0 \\|[[:space:]]+\\|[[:space:]]+\\|[[:space:]]+\\|[[:space:]]+\\|[[:space:]]+\\|[[:space:]]+" "${doc_path}"; then
	add_error "template placeholders still exist"
fi

if ! rg -q 'merge safety status:[[:space:]]+`(safe_to_merge|blocked)`' "${doc_path}"; then
	add_warning "merge safety status should use `safe_to_merge` or `blocked`"
fi

if [[ "${doc_path}" != "${ROOT_DIR}"/docs/phase-handoff-playbook/* ]]; then
	add_warning "doc is outside docs/phase-handoff-playbook/"
fi

if [[ ${#warnings[@]} -gt 0 ]]; then
	echo "WARNINGS:"
	for warning in "${warnings[@]}"; do
		echo "- ${warning}"
	done
fi

if [[ ${#errors[@]} -gt 0 ]]; then
	echo "HANDOFF CHECK FAILED:"
	for error in "${errors[@]}"; do
		echo "- ${error}"
	done
	exit 1
fi

echo "HANDOFF CHECK PASSED: ${doc_path}"
