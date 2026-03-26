# Phase Handoff Template

Save to:

- `docs/phase-handoff-playbook/YYYY-MM-DD-<topic>-handoff.md`

---

## 1. Snapshot

- worktree_path:
- branch:
- head_sha:
- origin_main_sha:
- origin_phase_branch_sha:
- open_pr:
- release_tag:
- merge_safety: `safe_to_merge` | `blocked` (+ reason)

## 2. Fresh Verification

| Command | Result | Notes |
| --- | --- | --- |
| `go test ./...` | PASS/FAIL |  |
| `<other command>` | PASS/FAIL |  |

## 3. Completed Today

- item 1
- item 2

## 4. Still Unfinished

- item 1
- item 2

## 5. Tomorrow Kickoff

```bash
git fetch origin --prune
git checkout <branch>
git pull --ff-only origin <branch>
go test ./...
```

First execution task tomorrow:

- `<one smallest task>`

## 6. Smallest-Next-Task Checklist

| ID | Priority | Task | Why | Acceptance | File Scope | Dependency | Parallel-safe |
| --- | --- | --- | --- | --- | --- | --- | --- |
| T01 | P0 |  |  |  |  | none | yes/no |

## 7. Risks

- risk 1
- risk 2

## 8. Do Not Do Next

- trap 1
- trap 2

## 9. Reflection

- what worked:
- what slowed down:
- one process improvement for next session:
