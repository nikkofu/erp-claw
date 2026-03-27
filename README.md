# ERP Claw

ERP Claw is the Go-based runtime foundation for the AI-native ERP platform. It defines five runtime entrypoints so each role can boot with the shared bootstrap helpers.

## Repository And Module

- GitHub repository: `https://github.com/nikkofu/erp-claw`
- Go module path: `github.com/nikkofu/erp-claw`
- Install dependency: `go get github.com/nikkofu/erp-claw`

The local `references/` directory is used only as an in-workspace research input. It is excluded from git and is not published to the remote repository.

## Runtime Entrypoints

- `cmd/api-server` → API server role (`api-server`)
- `cmd/agent-gateway` → Agent gateway role (`agent-gateway`)
- `cmd/worker` → Background worker role (`worker`)
- `cmd/scheduler` → Scheduler role (`scheduler`)
- `cmd/migrate` → Migration/bootstrapper role (`migrate`)

Each entrypoint uses `internal/bootstrap` to surface the shared runtime metadata and stubbed launch logic.

## Runtime Role Purpose

- `api-server` serves tenant-aware HTTP APIs for platform, admin, workspace, and integration surfaces.
- `agent-gateway` is the runtime seam for workspace presence and agent event ingress/egress.
- `worker` owns asynchronous processing such as outbox draining and later background execution.
- `scheduler` emits timed orchestration signals instead of mutating business data directly.

## Dependencies

Third-party services (databases, message brokers, etc.) are expected to be brought up through `docker-compose.yml` before running any Go runtime.

## Local Development

Local infrastructure is defined in `docker-compose.yml` and uses the defaults from `configs/local/docker.env`. `make infra-up`/`infra-down` pass that file into `docker compose`, so editing it changes the Compose-side environment values (Postgres, MinIO, etc.) and the published host ports (default: Postgres 5432, Redis 6379, NATS 4222, MinIO 9000/9001, OTEL gRPC 4317, Prometheus 9090, Grafana 3000). Those edits only affect the Compose stack; host-run Go processes read `configs/local/app.yaml` (or explicit `ERP_*` overrides). If you change Compose ports inside `docker.env`, follow up by adjusting `configs/local/app.yaml` or by setting matching `ERP_*` variables before starting the runtime.

1. Run `make infra-up` to start the supporting services in detached mode.
2. Use `docker compose ps` or `docker compose port <service> <container-port>` (run with the same env file) before connecting your Go programs, so you can confirm each published port.
3. Run `make infra-down` when you are done; it stops services without removing their volumes. If you need a clean slate, `make infra-reset` removes the named volumes and then brings the stack back up.

Go runtimes (`make api`, `make gateway`, `make worker`, `make scheduler`) always run outside of Docker Compose in local development, so keep `make infra-*` running for the duration of your work or tests that depend on those services.

Use `make test` to run the full Go test suite and `make test-integration` to rerun the compose contract integration test that ensures the services remain in `docker-compose.yml`.

## Local Workflow

The intended local verification loop for the platform foundation is:

1. `make infra-up`
2. `go run ./cmd/api-server`
3. `go run ./cmd/agent-gateway`
4. `make test`
5. `make smoke`

## Phase Handoff Playbook

Use the playbook whenever a delivery wave is paused and another session (or agent) needs to continue safely.

1. Create a dated handoff document:
   - `./scripts/new_phase_handoff.sh <topic-slug>`
   - compatibility alias: `./scripts/phase_handoff_new.sh <topic-slug>`
   - or `make handoff-new HANDOFF_TOPIC=<topic-slug>`
2. Fill the generated markdown under `docs/phase-handoff-playbook/`.
3. Run the quality gate before you stop:
   - `./scripts/phase_handoff_check.sh <handoff-doc-path>`
   - or `make handoff-check HANDOFF_DOC=<handoff-doc-path>`
4. Optional resume helper for next morning:
   - `./scripts/phase_resume_from_latest.sh`
5. Optional local pre-push gate:
   - `./scripts/phase_handoff_pre_push.sh`
   - or `make handoff-prepush`

PR/main now enforces changed handoff docs via `.github/workflows/handoff-quality-gate.yml`.

Reusable templates/checklists live in `skills/phase-handoff-playbook/`; project-specific outputs live in `docs/phase-handoff-playbook/`.

## Async Runtime Notes

The worker and scheduler are wired as separate runtime roles because they serve different responsibilities in the platform execution plane.

- `cmd/worker` is the long-running consumer that will drain the relational outbox and publish domain or integration events in later tasks.
- `cmd/scheduler` emits time-based commands onto the event bus and should not mutate business data directly.

## Command Pipeline

Application commands flow through a shared pipeline before any domain mutation happens.

- Commands enter application handlers through `internal/application/shared`.
- Policy evaluation runs before mutation so tenant or actor rules can stop unsafe work early.
- Audit recording wraps command execution so later observability and compliance hooks share one contract.
- Later ERP domain modules will plug into this same pipeline instead of bypassing it.

## Phase 2 Admin Flow (Wave 1-6 Baseline)

The first executable Phase 2 slice is now available through the admin surface. It currently uses in-memory repositories at runtime and forward-looking SQL migrations for the future PostgreSQL-backed implementation.

- `POST /api/admin/v1/master-data/suppliers`
- `POST /api/admin/v1/master-data/products`
- `POST /api/admin/v1/master-data/warehouses`
- `POST /api/admin/v1/procurement/purchase-orders`
- `POST /api/admin/v1/procurement/purchase-orders/:id/submit`
- `POST /api/admin/v1/procurement/purchase-orders/:id/receive`
- `POST /api/admin/v1/procurement/purchase-orders/:id/payable-bills`
- `GET /api/admin/v1/procurement/purchase-orders/:id`
- `GET /api/admin/v1/approvals` (supports `status=pending|approved|rejected`)
- `GET /api/admin/v1/inventory/ledger?product_id=<id>&warehouse_id=<id>`
- `GET /api/admin/v1/inventory/balances?product_id=<id>&warehouse_id=<id>` (returns `on_hand` / `reserved` / `available`)
- `POST /api/admin/v1/inventory/reservations`
- `POST /api/admin/v1/inventory/outbounds`
- `POST /api/admin/v1/inventory/transfers`
- `POST /api/admin/v1/inventory/transfer-orders`
- `GET /api/admin/v1/inventory/transfer-orders` (supports `status`, `sort=id_asc|id_desc`, `page`, `page_size`)
- `GET /api/admin/v1/inventory/transfer-orders/:id`
- `POST /api/admin/v1/inventory/transfer-orders/:id/execute`
- `POST /api/admin/v1/inventory/transfer-orders/:id/cancel`
- `POST /api/admin/v1/receivables`
- `GET /api/admin/v1/receivables`
- `GET /api/admin/v1/receivables/:id`
- `GET /api/admin/v1/payables`
- `GET /api/admin/v1/payables/:id`
- `POST /api/admin/v1/payables/:id/payment-plans`
- `POST /api/admin/v1/sales-orders`
- `GET /api/admin/v1/sales-orders`
- `GET /api/admin/v1/sales-orders/:id`
- `POST /api/admin/v1/sales-orders/:id/ship`
- `GET /api/admin/v1/read-models/overview`
- `POST /api/admin/v1/approvals/:id/approve`
- `POST /api/admin/v1/approvals/:id/reject`

Run `go test ./test/integration -run 'TestAdminSupplyChainFlow|TestAdminApprovalListSupportsStatusFilter|TestAdminInventoryReceiptFlow|TestAdminInventoryReservationFlow|TestAdminInventoryReservationRejectsExcessQuantity|TestAdminInventoryOutboundFlow|TestAdminInventoryOutboundRejectsExcessQuantity|TestAdminInventoryTransferFlow|TestAdminInventoryTransferRejectsExcessQuantity|TestAdminInventoryTransferOrderWorkflow|TestAdminInventoryTransferOrderListSupportsStatusSortAndPagination|TestAdminInventoryTransferOrderCancelFlow|TestAdminInventoryLedgerListFlow|TestAdminPayableFlow|TestAdminReceivableFlow|TestAdminSalesOrderShipFlow|TestAdminSalesOrderShipRejectsInsufficientInventory|TestAdminBackofficeOverviewReadModel' -v` to verify the end-to-end Phase 2 admin flow locally, including approval list query, inventory reservation/outbound/transfer, transfer-order execution/query/cancel (status/sort/pagination), and ledger query, payable/receivable basics, minimal sales shipment loop, and the backoffice overview read model.

## Workspace API (Phase 2 Minimal Query Slice)

- `GET /api/workspace/v1/inventory/balances?product_id=<id>&warehouse_id=<id>`
- `GET /api/workspace/v1/inventory/ledger?product_id=<id>&warehouse_id=<id>`
- `GET /api/workspace/v1/sales-orders`
- `GET /api/workspace/v1/sales-orders/:id`
- `GET /api/workspace/v1/payables`
- `GET /api/workspace/v1/payables/:id`
- `GET /api/workspace/v1/receivables`
- `GET /api/workspace/v1/receivables/:id`

Run `go test ./test/integration -run 'TestWorkspaceInventoryQueriesReturnBalanceAndLedger|TestWorkspaceSalesOrderQueriesReturnListAndDetail|TestWorkspaceFinanceQueriesReturnPayableAndReceivableReadModels' -v` to verify workspace inventory/sales/finance query routing and response shape.

## Integration API (Phase 2 Minimal Query Slice)

- `GET /api/integration/v1/read-models/overview`
- `GET /api/integration/v1/sales-orders`
- `GET /api/integration/v1/sales-orders/:id`
- `GET /api/integration/v1/payables`
- `GET /api/integration/v1/payables/:id`
- `GET /api/integration/v1/receivables`
- `GET /api/integration/v1/receivables/:id`

Run `go test ./test/integration -run 'TestIntegrationReadModelAndSalesQueries|TestIntegrationFinanceQueriesReturnPayableAndReceivableReadModels' -v` to verify integration overview/sales/finance query routing.

## Smoke Run

Start the API server and hit the health endpoint to confirm the router boots:

```bash
go run ./cmd/api-server
curl -sS http://127.0.0.1:8080/api/platform/v1/health/livez
```

Use `make smoke` to run the live HTTP health probe in `TestHealthRoutesLive` after the API server is already listening locally.

## Configuration

Bootstrap configuration is explicit: `LoadConfig(path string)` seeds defaults, optionally reads the requested YAML file, and then applies any `ERP_` environment overrides before returning the final `Config`. Runtimes that need the local stack behavior should pass `configs/local/app.yaml`; other profiles can reuse this file as their template, but there is no automatic profile discovery in `LoadConfig` yet.

YAML files currently mirror the `Config` shape with top-level `http`, `database`, `redis`, `nats`, `objectStorage`, and `telemetry` sections so that the Go code can unmarshal directly into those fields without additional nesting.

Environment variables whose names start with `ERP_` override any prior setting; the loader currently respects `ERP_ENV`, `ERP_HTTP_HOST`, `ERP_HTTP_PORT`, `ERP_STORE_DATABASE_DSN`, `ERP_STORE_REDIS_ADDR`, `ERP_TELEMETRY_METRICS_PATH`, and `ERP_TELEMETRY_TRACING_ENDPOINT`.

Keep `configs/local/app.yaml` synchronized with `internal/bootstrap/config.go` so the compose stack and the Go entrypoints remain aligned.

## Request Flow

Inbound HTTP calls pass through the Gin engine in `internal/interfaces/http/router`. The middleware stack installed before any handler runs is:

1. Request ID middleware generates `req-<timestamp>` IDs, stores them in `RequestContext`, and exposes them as `request_id` (request meta) plus response `meta.request_id`.
2. Logging middleware reads `RequestContext` to log request ID, tenant ID, method, path, status, and duration.
3. Tenant middleware reads `X-Tenant-ID` (local mode uses this header), resolves it via `internal/platform/tenant.Resolver`, mirrors the header downstream, and teaches the context about the tenant.
4. Auth middleware injects the placeholder `system` actor (until IAM is implemented) through the same context and adds `X-Actor-ID`.
5. Audit middleware copies `RequestContext.RequestID` into `X-Audit-Request-ID` for downstream correlation.

Handlers rely on `internal/platform/runtime.RequestContextKey` to fetch the shared context, so the middleware keeps that object synchronized with the headers.
