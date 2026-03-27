# Changelog

All notable changes to this project will be documented in this file.

## [0.2.19] - 2026-03-27

### Added

- Runtime telemetry setup seam at `internal/infrastructure/observability/otel/setup.go` with validation + deterministic shutdown hook.
- Bootstrap telemetry wiring (`internal/bootstrap/runtime.go`) and runtime entrypoint integration for `api-server`, `agent-gateway`, `worker`, `scheduler`, and `migrate`.
- Unit coverage for telemetry setup behavior in `internal/infrastructure/observability/otel/setup_test.go` and `internal/bootstrap/runtime_test.go`.

### Changed

- Refreshed `docs/phase-1-coverage-status.md` to match current repository reality (including the existing Phase 2 minimal supply-chain baseline and Phase 1 unfinished control-plane goals).

## [0.2.18] - 2026-03-27

### Added

- Transfer-order list filtering, sorting, and pagination baseline on the Phase 2 admin inventory path.

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
