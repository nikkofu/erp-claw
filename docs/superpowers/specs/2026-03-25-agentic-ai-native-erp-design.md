# Agentic AI-Native ERP Design

Date: 2026-03-25

## 1. Overview

This document defines the target architecture for a multi-tenant, AI-native, enterprise ERP platform rebuilt in Go with Gin as the primary HTTP framework. The project priority is platform-first, not feature breadth-first. The first version must establish a durable foundation for:

- multi-tenant SaaS operation
- hybrid tenant isolation
- policy-bounded agent autonomy
- dual first-class entry points: Web Console and Agent Workspace
- supply-chain transaction loop as the first business closure

The design is informed by the reference project under `references/openclaw-main`. The reference is not used as a language or framework template. It is used as an architectural reference for:

- gateway-centric control-plane design
- explicit runtime facades
- session and execution context handling
- plugin and capability registration
- method and event catalogs
- execution lifecycle tracking

The new ERP should not become "a traditional ERP with an AI chat box attached." It should be designed as an ERP platform whose transaction core, execution plane, and governance plane are all first-class and explicitly modeled.

## 2. Confirmed Decisions

The following decisions were confirmed during the design discussion and are treated as baseline requirements:

- Architecture priority: platform foundation first
- Delivery model: multi-tenant SaaS
- Isolation model: layered hybrid isolation
- Agent execution mode: policy-bounded autonomy
- First business closure: supply-chain transaction loop
- First-class user surfaces: Web Console and Agent Workspace
- Third-party infrastructure for local development and integration: unified through project-root `docker-compose.yml`

These decisions constrain both the runtime architecture and the codebase structure.

## 3. Design Goals

### 3.1 Primary Goals

- Build a large-scale Go codebase that remains evolvable under enterprise complexity.
- Preserve transaction correctness in ERP core domains.
- Make Agent capabilities native to the platform without letting them bypass business rules.
- Support shared-tenant operation while allowing higher-isolation tenant routing for larger customers.
- Separate control-plane governance from domain execution and data storage concerns.
- Provide a clean transition path from a modular monolith to selective service extraction later.
- Standardize development and local integration on Docker Compose.

### 3.2 Non-Goals for Phase 1

- Full manufacturing, MRP, or production planning
- Full general ledger and accounting engine
- Full microservice decomposition
- Dedicated search clusters and dedicated vector databases from day one
- General-purpose low-code workflow platform
- Unrestricted autonomous agents

Phase 1 optimizes for a stable platform kernel, a correct supply-chain domain core, and a controlled AI execution model.

## 4. Reference Project Learnings

The reference project demonstrates several architectural choices worth translating into the ERP design.

### 4.1 Gateway as Control Plane

The reference gateway is not just a transport layer. It acts as a control plane that concentrates runtime coordination, session handling, and registered methods. The ERP should follow the same principle:

- centralize agent-facing entry
- centralize platform governance and capability discovery
- avoid scattering orchestration across domain modules

### 4.2 Runtime Facades

The reference runtime facade pattern prevents plugins and tools from binding directly to internal implementation details. The ERP should mirror this by exposing:

- domain command facade
- domain query facade
- tool runtime facade
- policy runtime facade
- integration runtime facade

This prevents the Agent runtime and plugins from becoming tightly coupled to persistence or HTTP implementation details.

### 4.3 Explicit Run Context

The reference run-context logic normalizes identity and execution metadata before command execution. The ERP requires a stronger version of the same pattern:

- tenant context
- actor context
- policy context
- business context
- execution context
- knowledge context

No agent or workflow execution should run without an assembled execution context.

### 4.4 Plugin and Capability Registry

The reference plugin registry treats tools, routes, handlers, and providers as explicit runtime registrations. This is directly applicable to:

- connectors
- enterprise integrations
- external tool adapters
- tenant-enabled plugins
- model/tool catalogs for Agent use

### 4.5 Session Lifecycle Tracking

The reference lifecycle model tracks execution states and transitions. The ERP must do the same for:

- agent tasks
- approval instances
- workflows
- long-running integration jobs

The ERP version must be stricter because task execution impacts real transaction state.

## 5. System Context and Top-Level Architecture

The platform is structured around five planes.

### 5.1 Experience Plane

The user-facing plane consisting of:

- Web Console
- Agent Workspace
- public/integration API surfaces
- webhook ingress

This plane is responsible for interaction only. It must not contain domain logic.

### 5.2 Control Plane

The platform governance plane responsible for:

- tenant management
- identity and access management
- policy definition and evaluation
- quotas and feature flags
- model and tool catalogs
- plugin registry
- audit governance
- execution/session metadata

This plane acts as the platform brain.

### 5.3 Domain Plane

The ERP transaction core responsible for business truth:

- master data
- sales
- procurement
- inventory
- receivables
- payables
- approval references

This plane owns transactional correctness and business invariants.

### 5.4 Execution Plane

The AI and automation plane responsible for:

- agent runtime
- task planning
- workflow orchestration
- policy-gated execution
- human escalation
- tool invocation
- knowledge retrieval

This plane coordinates work, but does not own domain truth.

### 5.5 Data and Integration Plane

The persistence and integration plane responsible for:

- OLTP persistence
- event publication and consumption
- cache
- object storage
- search
- vector retrieval
- third-party integrations

### 5.6 Top-Level Flow

```text
Web Console / Agent Workspace / API / Webhook
                    |
                    v
        API Entry / Agent Gateway / Ingress
                    |
          +---------+---------+
          |                   |
          v                   v
      Control Plane      Execution Plane
          |                   |
          +---------+---------+
                    |
                    v
               Domain Plane
                    |
                    v
         Data and Integration Plane
```

The key rule is:

- Domain Plane owns transaction truth.
- Execution Plane owns planning and orchestration.
- Control Plane owns governance.

Neither workflow nor agent orchestration is allowed to become the source of business truth.

## 6. Multi-Tenant Control Plane Design

### 6.1 Control-Plane Subdomains

The control plane should contain the following core subdomains:

- tenant management
- organization and identity
- authorization
- policy
- agent governance
- plugin and integration registry
- audit and compliance
- runtime session control

### 6.2 Tenant Isolation Model

The system uses hybrid isolation:

- default tenants operate in shared infrastructure
- high-compliance or large tenants can route to dedicated schema or dedicated database

Isolation must be enforced in three layers:

- logical isolation
- data isolation
- execution isolation

Logical isolation means every request, task, event, or tool call is explicitly bound to a tenant context.

Data isolation means storage is routed by tenant isolation policy.

Execution isolation means:

- queue namespaces
- cache namespaces
- object storage prefixes
- vector namespaces
- plugin availability
- model/tool policies

must all be scoped by tenant.

### 6.3 Tenant Routing Model

The control plane should define these platform objects:

- `Tenant`
- `TenantCell`
- `TenantDataRoute`

`TenantDataRoute` is resolved per request and determines:

- control-plane database route
- tenant OLTP database route
- schema name
- cache namespace
- object storage prefix
- encryption scope
- vector namespace
- event topic prefix

This route is a first-class runtime object and must be injected into the platform context.

## 7. Identity, Authorization, and Governance

### 7.1 Authorization Model

Traditional RBAC alone is not sufficient. The platform should combine:

- RBAC for menus and coarse function rights
- ABAC for data scope and contextual restrictions
- policy rules for dynamic evaluation, especially for Agent and workflow actions

### 7.2 Example Evaluation Dimensions

Policy checks may consider:

- tenant
- organization
- department
- legal entity
- warehouse
- partner
- amount thresholds
- document type
- risk level
- execution channel
- time window
- actor type

### 7.3 Agent Governance

Agents are governed entities. Each configured agent should have:

- identity
- allowed models
- allowed tools
- risk class
- execution bounds
- escalation rules
- audit level

An agent is not a privileged bypass path. It is a constrained executor operating through platform policy.

## 8. Go Architecture and Layering Strategy

The codebase should use a combination of:

- DDD for bounded contexts and aggregates
- Hexagonal Architecture for interface isolation
- CQRS for separation of command and read concerns
- Outbox pattern for reliable domain event publication
- workflow/saga patterns for long-running business processes

Gin is only an interface adapter. It must not become the business architecture.

### 8.1 High-Level Code Layout

```text
erp-claw/
  cmd/
  internal/
    bootstrap/
    platform/
    domain/
    application/
    interfaces/
    infrastructure/
  configs/
  deployments/
  docs/
  api/
  migrations/
  scripts/
  test/
  docker-compose.yml
  Makefile
  go.mod
```

### 8.2 Layer Responsibilities

#### `internal/bootstrap`

Application assembly:

- configuration loading
- wiring
- server startup
- worker startup
- infrastructure initialization

#### `internal/platform`

Shared platform kernel:

- tenant
- iam
- policy
- audit
- runtime
- plugin
- eventbus
- knowledge
- storage
- observability
- idempotency
- locking

#### `internal/domain`

Pure business model:

- entities
- value objects
- aggregates
- domain services
- domain events
- repository interfaces

#### `internal/application`

Use-case orchestration:

- command handlers
- query handlers
- transaction management
- policy enforcement
- audit integration
- event outbox coordination

#### `internal/interfaces`

External entry points:

- HTTP
- WebSocket
- jobs
- webhooks
- event consumers
- agent entry adapters

#### `internal/infrastructure`

Technical implementation:

- PostgreSQL repositories
- Redis cache
- NATS messaging
- object storage
- vector retrieval
- model integrations

### 8.3 Architectural Constraints

- Domain packages must not import Gin, Redis, SQL drivers, or NATS SDKs.
- Application layer must not depend on transport-specific types.
- Interfaces layer must not implement business rules.
- Infrastructure layer must satisfy interfaces defined upward, never reverse the dependency direction.

## 9. Domain Model and Bounded Contexts

Phase 1 should define eight primary bounded contexts.

### 9.1 Tenant and IAM Context

Responsibilities:

- tenant metadata
- organization structures
- users
- roles
- department scope
- policy references

### 9.2 Master Data Context

Responsibilities:

- customer
- supplier
- product
- warehouse
- location
- tax code
- currency
- unit of measure
- price list

### 9.3 Sales Context

Responsibilities:

- quotation
- sales order
- delivery planning
- order lifecycle

### 9.4 Procurement Context

Responsibilities:

- purchase requisition
- purchase order
- supplier transaction state
- receipt planning

### 9.5 Inventory Context

Responsibilities:

- inbound receipt
- outbound issue
- stock reservation
- stock transfer
- inventory ledger
- available versus reserved stock state

### 9.6 Receivable and Payable Context

Responsibilities:

- receivable bill
- payable bill
- invoice application
- payment plan

### 9.7 Approval and Workflow Context

Responsibilities:

- approval definitions
- approval instances
- human tasks
- workflow state progression

### 9.8 Agent Task and Automation Context

Responsibilities:

- agent profile
- agent session
- agent task
- execution plan
- tool execution record
- policy decision record

### 9.9 Context Interaction Rules

- Master Data provides references to transaction contexts.
- Sales and Procurement produce fulfillment intent.
- Inventory owns stock movement truth.
- Approval controls progression permission, not business truth.
- Agent Automation invokes application commands; it never mutates transaction storage directly.

## 10. Agentic Execution Architecture

The AI-native behavior is implemented through a dedicated execution model, not through ad hoc LLM calls inside handlers.

### 10.1 Core Components

- Agent Gateway
- Session Context Assembler
- Tool Catalog
- Tool Runtime
- Policy Engine
- Workflow Orchestrator
- Knowledge Retrieval Layer
- Command Translator
- Execution State Machine

### 10.2 Agent Gateway

The Agent Gateway is the primary entry for:

- workspace tasks
- event-triggered tasks
- scheduled tasks
- webhook-triggered automation

It is responsible for:

- creating sessions and tasks
- resolving runtime context
- selecting model and tool policy
- streaming state and result events
- coordinating with application-layer commands

It should be deployed as its own process under `cmd/agent-gateway`.

### 10.3 Session Context

Every task execution must include:

- tenant context
- actor context
- business context
- policy context
- knowledge context
- execution context

Task execution without a resolved context is invalid by design.

### 10.4 Tool Model

Tools should be grouped into:

- query tools
- command tools
- action tools
- knowledge tools

Every tool must declare:

- ID
- category
- input schema
- output schema
- required permissions
- risk level
- idempotency rule
- audit requirement
- approval requirement

### 10.5 Policy Enforcement

Every Agent-proposed action must pass through policy evaluation with one of these decisions:

- `ALLOW`
- `ALLOW_WITH_GUARD`
- `REQUIRE_APPROVAL`
- `DENY`

No direct database mutation is allowed from the Agent runtime.

### 10.6 Workflow Orchestration

Long-running processes should be modeled as workflows rather than database transactions. Typical Phase 1 workflow candidates:

- replenishment suggestion to procurement execution
- stock shortage handling
- approval-driven purchase creation
- collections escalation

### 10.7 Execution State Machine

Suggested states:

- `draft`
- `planned`
- `policy_checking`
- `awaiting_approval`
- `executing`
- `partially_succeeded`
- `succeeded`
- `failed`
- `rolled_back`
- `cancelled`
- `escalated_to_human`

Each transition must record:

- input
- output
- policy decision
- tools used
- timing
- actor
- evidence
- failure cause

## 11. Data Architecture and Storage Strategy

### 11.1 Phase 1 Storage Stack

Recommended baseline:

- PostgreSQL
- Redis
- NATS JetStream
- S3-compatible object storage
- OpenTelemetry

Optional later upgrades:

- Temporal
- OpenSearch
- dedicated vector database

### 11.2 Logical Data Stores

Three logical stores are recommended:

#### Control Plane DB

Stores:

- tenant and subscription metadata
- tenant routing
- global policy definitions
- model and plugin registry
- quota
- feature flags

#### Tenant OLTP DB

Stores:

- transaction truth
- master data
- sales
- procurement
- inventory
- receivables and payables
- approval references
- outbox

#### Execution and Analytics Store

Phase 1 may still place this in PostgreSQL, logically separated, for:

- agent session records
- tool execution records
- policy decisions
- workflow evidence
- retrieval chunks
- workspace read models

### 11.3 CQRS Strategy

Phase 1 should use light CQRS:

- commands go through application command handlers
- queries use dedicated read services and projection tables

Recommended read models:

- Backoffice Read Model
- Workspace Read Model
- Monitoring Read Model
- Analytics Snapshot Read Model

### 11.4 Outbox Pattern

All cross-module asynchronous communication must rely on outbox publication. Outbox messages should include:

- event ID
- tenant ID
- aggregate type
- aggregate ID
- event type
- payload
- occurred at
- trace ID
- causation ID
- correlation ID

### 11.5 Cache Strategy

Redis should be used for:

- query caching
- distributed locks
- idempotency keys
- rate limiting
- task deduplication
- hot read model acceleration

All keys must be tenant- and/or cell-scoped.

### 11.6 Search and Vector Retrieval

Phase 1 should keep search and vector retrieval simple:

- PostgreSQL full-text search
- trigram support
- `pgvector`

This minimizes platform sprawl while preserving tenant-aware filtering and permission-aware retrieval.

## 12. API and Interface Design

API surfaces should be separated by interaction semantics.

### 12.1 Entry Groups

- `Admin API`
- `Workspace API`
- `Platform API`
- `Integration API`

Suggested route prefixes:

- `/api/admin/v1/...`
- `/api/workspace/v1/...`
- `/api/platform/v1/...`
- `/api/integration/v1/...`
- `/ws/workspace`
- `/ws/agent-events`

### 12.2 Command and Query Rules

- Commands mutate state and must go through application command handlers.
- Queries are read-only and must use query/read-model services.
- Business actions should use explicit action-oriented endpoints instead of generic PATCH semantics.

Examples:

- `POST /admin/v1/purchase-orders/{id}:submit`
- `POST /admin/v1/purchase-orders/{id}:approve`
- `POST /admin/v1/purchase-orders/{id}:reject`

### 12.3 Middleware Stack

Recommended middleware order:

- recovery
- request ID and tracing
- structured logging
- tenant resolution
- authentication
- authorization and data scope bootstrap
- rate limit and idempotency
- audit metadata injection

Handlers should only:

- bind request DTOs
- validate basic input
- fetch platform context
- call application handlers
- present DTO responses

## 13. Docker Compose and Local Resource Strategy

All third-party local infrastructure must be managed through project-root `docker-compose.yml`.

### 13.1 Required Phase 1 Compose Services

- PostgreSQL
- Redis
- NATS
- MinIO
- OpenTelemetry Collector
- Prometheus
- Grafana

Optional services depending on implementation pace:

- Temporal or Temporalite
- MailHog
- Mock server

### 13.2 Compose Role

`docker-compose.yml` is a first-class development contract and should support:

- local development
- integration testing
- demo startup
- onboarding

The Go processes themselves may still run locally from the host during development, while third-party infrastructure is standardized through Compose.

## 14. Deployment Blueprint

Phase 1 should not start as full microservices, but runtime roles should still be separated.

### 14.1 Runtime Units

- `api-server`
- `agent-gateway`
- `worker`
- `scheduler`
- `migrate`

### 14.2 Logical Topology

```text
[Web Console] ----\
[Agent Workspace] -- > [Ingress / API Gateway] -> [api-server]
[Webhook / API] --/                            -> [agent-gateway]
                                                [worker]
                                                [scheduler]
                                                      |
                            +-------------------------+-------------------------+
                            |                         |                         |
                        [PostgreSQL]               [Redis]                   [NATS]
                            |                                                   |
                            +----------------------[Object Storage]------------+
```

### 14.3 Cell Evolution Path

Future evolution:

- shared cell A
- shared cell B
- dedicated enterprise cell X

The routing model must support that without redesigning the application layer.

## 15. Error Handling and Reliability

The platform should define a shared application error taxonomy:

- validation error
- domain error
- policy error
- infrastructure error
- execution error

Reliability rules:

- all commands support idempotency keys where relevant
- all cross-boundary async actions use outbox
- consumers use idempotent processing
- agent execution always records plan, policy, action, and evidence

## 16. Observability and Audit

The platform requires explicit observability and audit from day one.

### 16.1 Observability

- structured logs
- traces
- metrics
- task timeline visibility
- queue depth visibility
- policy decision metrics

### 16.2 Audit

All high-impact actions must retain:

- actor
- tenant
- source channel
- command or task identity
- target business object
- policy result
- before/after state summary
- trace ID

Actions generated by agents must always be distinguishable from human-initiated actions.

## 17. Testing Strategy

The test strategy should be layered.

### 17.1 Domain Unit Tests

Focus:

- aggregates
- value objects
- domain rules

### 17.2 Application Tests

Focus:

- command and query handlers
- transaction coordination
- policy checks
- audit integration

### 17.3 Repository Integration Tests

Focus:

- PostgreSQL persistence
- Redis behavior
- outbox and inbox processing
- isolation routing

### 17.4 API Contract Tests

Focus:

- Admin API
- Workspace API
- Platform API

### 17.5 Workflow and Agent Scenario Tests

Focus:

- task planning
- policy gating
- tool execution
- approval pause/resume
- escalation
- partial success and rollback paths

### 17.6 Security and Isolation Tests

Focus:

- tenant isolation
- data-scope enforcement
- tool authorization
- plugin boundary enforcement
- attachment and retrieval permission boundaries

## 18. Initial Project Skeleton Recommendation

The initial scaffold should optimize for architectural correctness, not for feature volume.

Suggested first wave of files and packages:

- `go.mod`
- `cmd/api-server/main.go`
- `cmd/agent-gateway/main.go`
- `cmd/worker/main.go`
- `cmd/scheduler/main.go`
- `internal/bootstrap/...`
- `internal/platform/tenant/...`
- `internal/platform/iam/...`
- `internal/platform/policy/...`
- `internal/platform/audit/...`
- `internal/platform/eventbus/...`
- `internal/interfaces/http/router/...`
- `internal/interfaces/http/middleware/...`
- `internal/infrastructure/persistence/postgres/...`
- `internal/infrastructure/cache/redis/...`
- `configs/local/...`
- `migrations/...`
- `docker-compose.yml`
- `Makefile`
- `README.md`

This creates a stable base for implementation planning.

## 19. Recommended Phase Roadmap

### Phase 0: Runtime and Engineering Skeleton

- repository bootstrap
- Compose infrastructure
- process entry points
- health checks
- configuration
- observability

### Phase 1: Platform Control Plane

- tenant and cell routing
- IAM
- policy
- audit
- plugin and tool registration
- agent session/task model

### Phase 2: Supply-Chain Transaction Loop

- master data
- sales
- procurement
- inventory
- approval
- receivable/payable basics

### Phase 3: AI-Native Execution Strengthening

- Agent Workspace
- knowledge retrieval
- policy-bounded autonomy
- workflow orchestration
- human handoff and escalation

### Phase 4: Enterprise Expansion

- open integration platform
- dedicated tenant cells
- advanced search
- advanced analytics
- deeper workflow and model governance

## 20. Final Recommendation

The recommended architecture is:

- a platform-oriented, modular Go codebase
- Gin as interface adapter only
- PostgreSQL-centered transaction and read-model strategy
- hybrid multi-tenant isolation through routed tenant cells
- explicit control plane for governance
- explicit execution plane for Agent runtime
- domain truth preserved in transaction contexts
- all third-party local infrastructure standardized through `docker-compose.yml`

This is the correct midpoint between:

- an overly simple ERP monolith that cannot absorb AI-native behavior
- an over-engineered microservice platform that collapses under its own coordination cost

It creates a realistic Phase 1 path and a clear Phase 2+ evolution route.
