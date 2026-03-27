# Changelog

All notable changes to this project will be documented in this file.

## [0.2.33] - 2026-03-27

### Added

- Phase 1 inbox idempotency storage baseline:
  - migration `000011_init_phase1_inbox_idempotency_tables` adds `inbox` table
  - unique key `(tenant_id, message_key)` for dedupe claims
  - index `idx_inbox_status_received_at` for status/latency style scans
- New postgres inbox store helper at `internal/infrastructure/persistence/postgres/inbox.go`:
  - `ClaimMessage` for insert-once idempotency claims
  - `MarkProcessed` and `MarkFailed` for consumer-side processing state updates
- Unit tests for inbox store claim/duplicate/validation/state-update behavior in `internal/infrastructure/persistence/postgres/inbox_test.go`.

### Changed

- `docs/phase-1-coverage-status.md` now tracks inbox idempotency baseline as delivered and marks consumer-chain integration as remaining work.

## [0.2.32] - 2026-03-27

### Added

- Worker now emits dead-letter events when outbox publish attempts are exhausted:
  - after marking the record `failed`, emits `platform.outbox.dead_letter`
  - dead-letter payload includes outbox identity, original topic/event type, payload bytes, attempts, error, and failed timestamp
- Expanded worker outbox unit tests for terminal-failure dead-letter behavior:
  - success path where dead-letter event is published
  - fallback path where dead-letter publish also fails but outbox record still transitions to `failed`

### Changed

- `docs/phase-1-coverage-status.md` now records dead-letter publication as part of the outbox governance baseline.

## [0.2.31] - 2026-03-27

### Added

- Admin purchase-order list query endpoint:
  - `GET /api/admin/v1/procurement/purchase-orders`
  - supports `status=draft|pending_approval|approved|rejected|received`
  - supports `sort=id_asc|id_desc`, `page`, `page_size`
- Supply-chain service query capability:
  - `ListPurchaseOrders` with status/sort/pagination validation and tenant scope
- Procurement repository contract extended with tenant-scoped list:
  - `ListByTenant(ctx, tenantID)`
- New tests for purchase-order list query behavior:
  - `TestServiceListPurchaseOrdersSupportsStatusSortAndPagination`
  - `TestServiceListPurchaseOrdersFailsForInvalidQuery`
  - `TestAdminPurchaseOrderListSupportsStatusSortAndPagination`
  - `TestAdminPurchaseOrderListRejectsInvalidQuery`

### Changed

- Updated `README.md` and `docs/phase-2-coverage-status.md` to include purchase-order list query coverage.

## [0.2.30] - 2026-03-27

### Added

- Admin approval query now supports stable sorting and pagination:
  - `GET /api/admin/v1/approvals?sort=id_asc|id_desc&page=<n>&page_size=<n>`
- Supply-chain approval query input now supports:
  - `sort`, `page`, `page_size`
- New tests for approval sort/pagination behavior:
  - `TestServiceListApprovalRequestsSupportsSortAndPagination`
  - `TestServiceListApprovalRequestsFailsForInvalidSortAndPagination`
  - `TestAdminApprovalListSupportsSortAndPagination`
  - `TestAdminApprovalListRejectsInvalidSortAndPaginationQuery`

### Changed

- Updated `README.md` and `docs/phase-2-coverage-status.md` to reflect approval list sorting/pagination support.

## [0.2.29] - 2026-03-27

### Added

- Admin approval query endpoint:
  - `GET /api/admin/v1/approvals` (supports `status=pending|approved|rejected`)
- Supply-chain service query capability:
  - `ListApprovalRequests` with tenant scope + status filter validation
- Approval repository contract extended with tenant-scoped list:
  - `ListByTenant(ctx, tenantID)`
- New tests for approval query behavior:
  - `TestServiceListApprovalRequestsSupportsStatusFilter`
  - `TestServiceListApprovalRequestsFailsForInvalidStatus`
  - `TestAdminApprovalListSupportsStatusFilter`
  - `TestAdminApprovalListRejectsInvalidStatusFilter`

### Changed

- Updated `README.md` and `docs/phase-2-coverage-status.md` to reflect approval list query coverage.

## [0.2.28] - 2026-03-27

### Added

- Worker outbox claim now supports lease-based recovery:
  - claim set expanded to `status in ('pending', 'processing')`
  - claim window is bounded by `available_at <= readyBefore`
  - claimed records are moved to `processing` with lease expiry written to `available_at`
- Worker outbox unit tests now assert claim arguments (`limit`, `readyBefore`, `leaseUntil`) to lock the lease behavior.

### Changed

- `docs/phase-1-coverage-status.md` now documents the outbox processing lease/reclaim semantics.

## [0.2.27] - 2026-03-27

### Added

- Workspace query surface now includes finance read models:
  - `GET /api/workspace/v1/payables`
  - `GET /api/workspace/v1/payables/:id`
  - `GET /api/workspace/v1/receivables`
  - `GET /api/workspace/v1/receivables/:id`
- Integration query surface now includes finance read models:
  - `GET /api/integration/v1/payables`
  - `GET /api/integration/v1/payables/:id`
  - `GET /api/integration/v1/receivables`
  - `GET /api/integration/v1/receivables/:id`
- Added integration tests for workspace/integration finance queries:
  - `TestWorkspaceFinanceQueriesReturnPayableAndReceivableReadModels`
  - `TestIntegrationFinanceQueriesReturnPayableAndReceivableReadModels`

### Changed

- Updated `README.md` Phase 2 query API sections and verification commands.
- Updated `docs/phase-2-coverage-status.md` with latest workspace/integration query coverage.

## [0.2.24] - 2026-03-27

### Added

- Outbox retry-governance migration `000010_init_phase1_outbox_retry_governance_tables`:
  - `attempts`, `last_error`, `failed_at` columns
  - `idx_outbox_status_available_at` index
- Worker outbox processing now tracks failure attempts and applies terminal failure status:
  - publish failure increments attempts
  - below threshold: back to `pending` with delayed `available_at`
  - threshold reached: mark as `failed`
- Expanded outbox worker unit tests for retry-vs-failed branching.

### Changed

- `docs/phase-1-coverage-status.md` now reflects outbox attempt tracking and failed terminal state.

## [0.2.23] - 2026-03-27

### Added

- Worker outbox polling now implements a minimal publish loop:
  - claim pending records with `FOR UPDATE SKIP LOCKED`
  - publish events through `eventbus.Bus`
  - mark success as `published`
  - mark failures back to `pending` with delayed `available_at`
- Unit tests for outbox polling behavior in `cmd/worker/main_test.go` (happy path, publish failure retry path, claim error path).

### Changed

- `docs/phase-1-coverage-status.md` now reflects outbox progression from placeholder loop to minimal executable processing path.

## [0.2.22] - 2026-03-27

### Added

- Phase 2 mainline merge #9 baseline on top of `origin/main`.

## [0.2.21] - 2026-03-27

### Added

- Runtime telemetry setup seam at `internal/infrastructure/observability/otel/setup.go` with validation + deterministic shutdown hook.
- Bootstrap telemetry wiring (`internal/bootstrap/runtime.go`) and runtime entrypoint integration for `api-server`, `agent-gateway`, `worker`, `scheduler`, and `migrate`.
- Unit coverage for telemetry setup behavior in `internal/infrastructure/observability/otel/setup_test.go` and `internal/bootstrap/runtime_test.go`.

### Changed

- Refreshed `docs/phase-1-coverage-status.md` to match current repository reality (including the existing Phase 2 minimal supply-chain baseline and Phase 1 unfinished control-plane goals).

## [0.2.18] - 2026-03-27

### Added

- Transfer-order list filtering, sorting, and pagination baseline on the Phase 2 admin inventory path.

## [0.2.19] - 2026-03-27

### Added

- Transfer-order cancellation workflow baseline on the Phase 2 admin inventory path.

## [0.2.20] - 2026-03-27

### Added

- Phase 2 mainline merge #8 release baseline on top of `origin/main`.

## [0.2.17] - 2026-03-27

### Added

- Executable phase handoff playbook workflow with reusable assets under `skills/phase-handoff-playbook/`.
- Handoff generation and validation scripts: `scripts/new_phase_handoff.sh`, `scripts/phase_handoff_check.sh`, and compatibility alias `scripts/phase_handoff_new.sh`.
- CI gate for changed handoff docs: `.github/workflows/handoff-quality-gate.yml`.
- Local pre-push handoff quality helper: `scripts/phase_handoff_pre_push.sh` and `make handoff-prepush`.

### Changed

- README/Makefile/playbook docs wiring for handoff quality workflow.

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
