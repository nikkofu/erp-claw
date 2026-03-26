# Changelog

All notable changes to this project will be documented in this file.

## [0.2.6] - 2026-03-26

### Added

- Minimal workspace SSE stream route that replays tenant-scoped session history and then streams live workspace events.
- Atomic workspace gateway subscription support with history snapshots and multi-subscriber fan-out for the same session.
- Integration and gateway coverage for workspace stream replay/live delivery behavior.

### Changed

- Phase 1 docs and README now reflect that the workspace surface includes a minimal session-scoped streaming protocol, not only write/read-side HTTP endpoints.

## [0.2.5] - 2026-03-26

### Added

- Outbox catalog bootstrap with Postgres runtime wiring and in-memory test fallback for tenant-scoped operator reads and failed-message recovery.
- Admin API routes for listing tenant-scoped outbox messages by status and requeueing failed messages from the control plane.
- Integration coverage for the outbox operator surface, including a cross-tenant requeue rejection check.

### Changed

- Phase 1 docs and README now reflect that the reliability baseline includes an Admin outbox operator surface in addition to dispatcher, retry, and recovery seams.

## [0.2.4] - 2026-03-26

### Added

- Governance catalog bootstrap with Postgres runtime wiring and in-memory test fallback for policy rules and audit events.
- Admin API routes for policy rule create/list, activate/deactivate, and audit event list queries.
- Integration coverage for governance admin lifecycle routing on top of the existing policy rule and audit services.

### Changed

- Phase 1 docs now reflect that governance baseline includes an Admin HTTP management surface instead of only application-level handlers and repository seams.

## [0.2.3] - 2026-03-26

### Added

- Workspace write-side HTTP routes for creating sessions and tasks, transitioning tasks through start/complete/fail/cancel, and closing sessions.
- Integration coverage for a full workspace write flow that creates runtime state and verifies replayed workspace events through the existing read APIs.

### Changed

- Phase 1 docs now reflect that the workspace surface is no longer read-only; it now includes a minimal command/write seam on top of the existing session/task runtime service.

## [0.2.2] - 2026-03-26

### Added

- Capability catalog bootstrap with Postgres runtime wiring and in-memory test fallback for model catalog and tool catalog repositories.
- Admin API routes for tenant-scoped model catalog entries and tool catalog entries with create/list coverage.
- Integration and repository tests that verify capability admin lifecycle routing and `*sql.DB` adapter construction for the capability repository.

### Changed

- Phase 1 coverage docs and README now reflect that capability governance includes a minimal Admin HTTP management surface for model/tool catalog baselines.

## [0.2.1] - 2026-03-26

### Added

- Approval management application handlers for saving definitions and listing tenant-scoped approval definitions, instances, and tasks.
- Approval catalog bootstrap with Postgres runtime wiring and in-memory test fallback, aligned with the existing control-plane and agent-runtime catalog contracts.
- Admin API routes for approval definitions, approval instances, approval task lists, and task approve/reject actions.
- Integration and bootstrap tests that cover approval admin lifecycle flows and approval catalog fail-fast behavior.

### Changed

- Phase 1 coverage documentation now reflects that the approval baseline includes an Admin HTTP management surface instead of only domain-level pipeline integration.

## [0.2.0] - 2026-03-26

### Added

- Phase 1 control-plane catalog baseline for tenant, user, role, department, agent profile, and tenant-scoped bindings.
- Governance baseline with persisted policy rules, audit event storage/query seams, and repository-backed rule evaluation.
- Agent runtime control/read-side baseline for sessions, tasks, workspace events, and workspace query APIs.
- Capability governance baseline for tenant-scoped model catalog and tool catalog entries.
- Approval baseline with definition, instance, task, and `REQUIRE_APPROVAL` pipeline integration.
- Reliability baseline for outbox dispatch, retry, failed recovery, and worker poll instrumentation seams.
- Phase handoff playbook skill assets plus project handoff documents for Phase 1 continuation and collaboration.

### Changed

- `README.md`, roadmap docs, and Phase 1 coverage docs now reflect the current executable scope instead of the original skeleton-only status.
- Phase 1 stable collaboration baseline is now suitable for handing off to a parallel `Phase 2` branch or worktree.
- Runtime bootstrap now fails fast on database catalog initialization errors outside explicit test-mode paths.
- Workspace session/task read-side contracts are normalized around `session_key` at the HTTP, service, and Postgres repository boundary.
- Control-plane subordinate writes now require an existing tenant root instead of allowing orphaned tenant-scoped records.
- Outbox dispatch now requeues stale `publishing` rows before polling, so worker crashes do not strand messages indefinitely.

## [0.1.0] - 2026-03-25

### Added

- Initial Go runtime foundation for `api-server`, `agent-gateway`, `worker`, `scheduler`, and `migrate`.
- Local infrastructure contract with PostgreSQL, Redis, NATS, MinIO, OpenTelemetry Collector, Prometheus, and Grafana.
- Bootstrap configuration loading, health endpoints, HTTP middleware stack, tenant request seam, and workspace gateway skeleton.
- Infrastructure clients, platform migrations, event bus seam, command pipeline skeleton, and local smoke verification workflow.
- Phase roadmap and per-phase coverage documents under `docs/`.

### Notes

- Go module path is now `github.com/nikkofu/erp-claw`.
- The local `references/` directory is intentionally excluded from version control and is not synchronized to GitHub.
