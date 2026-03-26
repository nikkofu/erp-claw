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

## Admin API Control-Plane Slice (Phase 1)

The Admin API now includes the first control-plane catalog slice. Its purpose is to bootstrap and manage foundational catalog data needed before workspace and execution-plane workflows can be layered on top.

This is the first Phase 1 control-plane implementation, not the full control plane. It intentionally ships a narrow set of catalog capabilities so the runtime wiring and contracts can be validated incrementally.

Catalog entities in this slice:

- Tenants (control-plane catalog roots)
- IAM Users (tenant-scoped identities)
- Agent Profiles (governed AI agent definitions per tenant)
- Approval Definitions / Instances / Tasks (tenant-scoped governance workflow baseline)
- Model Catalog Entries / Tool Catalog Entries (tenant-scoped capability governance baseline)
- Policy Rules / Audit Events (tenant-scoped governance control baseline)
- Outbox Messages (tenant-scoped operator list/requeue surface for failed delivery recovery)

Bootstrap behavior for `bootstrap.NewContainer(cfg)`:

- Uses a Postgres-backed control-plane catalog repository when the configured database is reachable.
- Uses a Postgres-backed approval catalog repository for approval definitions, instances, and tasks when the configured database is reachable.
- Uses a Postgres-backed capability catalog repository for model catalog and tool catalog entries when the configured database is reachable.
- Uses a Postgres-backed governance catalog repository for policy rules and audit events when the configured database is reachable.
- Uses a Postgres-backed outbox catalog repository for tenant-scoped message inspection and failed-message requeue operations when the configured database is reachable.
- Fails fast when a runtime Postgres catalog cannot be initialized, so control-plane and workspace state do not silently degrade into ephemeral storage.
- Uses in-memory catalogs only for explicit test bootstrap paths such as `bootstrap.NewTestContainer()` or configurations with an empty database DSN.

## Workspace API Slice (Phase 1)

The Workspace API now exposes a minimal command surface in addition to the existing read-side endpoints.

- Create sessions
- Create tasks
- Start / complete / fail / cancel tasks
- Close sessions
- List sessions / tasks / replay events
- Stream session-scoped workspace events over SSE

This is still not a full real-time workspace protocol. It now includes the smallest session-scoped SSE seam that replays stored events and streams live updates over HTTP, but it does not yet provide a full WebSocket protocol or cross-process stream durability.

## Outbox Operator Slice (Phase 1)

The Admin API also exposes a minimal operator surface for outbox recovery.

- List tenant-scoped outbox messages filtered by status
- Requeue failed outbox messages by ID
- Reject cross-tenant requeue attempts before recovery is executed

This keeps the reliability baseline operable without expanding into a full dead-letter, replay, or idempotency management console yet.

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
