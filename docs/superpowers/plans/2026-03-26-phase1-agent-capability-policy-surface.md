# Phase 1 Agent Capability Policy Surface Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a minimal Phase 1 control-plane surface for tenant-scoped agent capability policy binding so an `agent_profile` can declare allowed model and tool catalog entries through Admin API.

**Architecture:** Persist policy bindings in dedicated relational tables keyed by `tenant_id + agent_profile_id`, keep normalization in the capability domain, and validate referenced profile/model/tool IDs in the application layer using existing control-plane and capability catalog seams. This slice intentionally excludes runtime executor enforcement, quota, feature flags, plugin registry, and tenant enablement.

**Tech Stack:** Go, Gin, PostgreSQL migrations, in-memory bootstrap catalogs for tests, `go test`

---

### Task 1: Domain and Migration Baseline

**Files:**
- Create: `internal/domain/capability/agent_capability_policy.go`
- Create: `internal/domain/capability/agent_capability_policy_test.go`
- Create: `internal/domain/capability/agent_capability_policy_repository.go`
- Create: `migrations/000010_phase1_agent_capability_policy.up.sql`
- Create: `migrations/000010_phase1_agent_capability_policy.down.sql`
- Modify: `test/integration/capability_governance_test.go`

- [ ] **Step 1: Write the failing tests**

Add tests for policy normalization and for the new migration contract.

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$(pwd)/.cache/go-build go test ./internal/domain/capability ./test/integration -run 'AgentCapabilityPolicy|CapabilityGovernanceMigrationDefinesAgentCapabilityPolicyTables' -count=1`
Expected: FAIL because the policy type and migration do not exist yet.

- [ ] **Step 3: Write minimal implementation**

Add the policy aggregate constructor and migration tables:
- `agent_profile_allowed_model`
- `agent_profile_allowed_tool`

- [ ] **Step 4: Run test to verify it passes**

Run: `GOCACHE=$(pwd)/.cache/go-build go test ./internal/domain/capability ./test/integration -run 'AgentCapabilityPolicy|CapabilityGovernanceMigrationDefinesAgentCapabilityPolicyTables' -count=1`
Expected: PASS

- [ ] **Step 5: Commit**

Run:

```bash
git add internal/domain/capability migrations test/integration/capability_governance_test.go docs/superpowers/plans/2026-03-26-phase1-agent-capability-policy-surface.md
git commit -m "feat: add phase1 agent capability policy domain"
```

### Task 2: Repository and Application Handlers

**Files:**
- Create: `internal/application/capability/save_agent_capability_policy.go`
- Create: `internal/application/capability/get_agent_capability_policy.go`
- Create: `internal/application/capability/agent_capability_policy_test.go`
- Modify: `internal/application/capability/repository_fake_test.go`
- Modify: `internal/bootstrap/container.go`
- Modify: `internal/bootstrap/capability_catalog.go`
- Modify: `internal/infrastructure/persistence/postgres/capability_repository.go`
- Modify: `internal/infrastructure/persistence/postgres/capability_repository_test.go`

- [ ] **Step 1: Write the failing tests**

Add tests for save/get handler behavior and repository SQL expectations.

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$(pwd)/.cache/go-build go test ./internal/application/capability ./internal/infrastructure/persistence/postgres -run 'AgentCapabilityPolicy|CapabilityRepository' -count=1`
Expected: FAIL because save/get policy support is missing.

- [ ] **Step 3: Write minimal implementation**

Implement in-memory and Postgres save/get support for normalized policy bindings.

- [ ] **Step 4: Run test to verify it passes**

Run: `GOCACHE=$(pwd)/.cache/go-build go test ./internal/application/capability ./internal/infrastructure/persistence/postgres -run 'AgentCapabilityPolicy|CapabilityRepository' -count=1`
Expected: PASS

- [ ] **Step 5: Commit**

Run:

```bash
git add internal/application/capability internal/bootstrap internal/infrastructure/persistence/postgres
git commit -m "feat: add phase1 agent capability policy repository"
```

### Task 3: Admin API Validation Surface

**Files:**
- Modify: `internal/interfaces/http/router/admin.go`
- Modify: `test/integration/admin_control_plane_test.go`
- Modify: `README.md`
- Modify: `CHANGELOG.md`
- Modify: `VERSION`
- Modify: `docs/phase-1-coverage-status.md`
- Modify: `docs/phase-handoff-playbook/2026-03-26-phase-1-stable-baseline.md`

- [ ] **Step 1: Write the failing tests**

Add integration tests for:
- successful `PUT/GET /api/admin/v1/agent-profiles/:profile_id/capability-policy`
- rejection of unknown or cross-tenant profile/model/tool references
- replace semantics, duplicate normalization, and empty-list reads

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$(pwd)/.cache/go-build go test ./test/integration -run 'AdminAgentCapabilityPolicy' -count=1`
Expected: FAIL because the routes do not exist yet.

- [ ] **Step 3: Write minimal implementation**

Add Admin API routes and keep tenant/profile/catalog validation in the application layer so the HTTP router stays thin.

- [ ] **Step 4: Run test to verify it passes**

Run: `GOCACHE=$(pwd)/.cache/go-build go test ./test/integration -run 'AdminAgentCapabilityPolicy' -count=1`
Expected: PASS

- [ ] **Step 5: Commit**

Run:

```bash
git add internal/interfaces/http/router test/integration README.md CHANGELOG.md VERSION docs
git commit -m "feat: add phase1 agent capability policy surface"
```

## Review Checklist

Before marking this slice complete, confirm all of the following:

- policy bindings stay tenant-scoped and agent-profile-scoped
- allowed model/tool IDs are normalized, deduplicated, and stable on read
- Admin API rejects references to unknown tenant-local profiles, models, and tools
- no runtime execution semantics are introduced in this slice
- no Phase 2 supply-chain domain code is mixed into this slice
- documentation explicitly marks tenant enablement, quota, and plugin registry as still pending

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-03-26-phase1-agent-capability-policy-surface.md`.

This session is continuing inline execution inside the existing `feature/phase1-control-plane` worktree with TDD and frequent verification.
