# Changelog

All notable changes to this project will be documented in this file.

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
