# Agent Operating Instructions (Repo-Level)

## Phase Handoff Startup Rule

When user explicitly says any of:

- `phase-handoff-playbook`
- `handoff`
- `交接`
- `明早开工`
- `继续昨天`

the assistant must treat it as a handoff-resume workflow first, before coding.

## Mandatory Startup Sequence

1. locate latest handoff under `docs/phase-handoff-playbook/`
2. restate snapshot (`worktree`, `branch`, `head`, `origin/main`)
3. rerun verification command(s) recorded in handoff
4. restate first execution task exactly as written
5. only then start implementation

## Source Of Truth

- Playbook skill docs: `skills/phase-handoff-playbook/`
- Project handoff docs: `docs/phase-handoff-playbook/`

If there are multiple handoff files, pick the newest date-prefixed file unless user specifies one.
