# Phase 1 Effective Capability Policy Surface Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a Phase 1 runtime-side capability resolution baseline that exposes the effective active model/tool allowlist for an agent profile without changing the stored capability policy contract.

**Architecture:** Keep the stored policy (`capability-policy`) unchanged and add a separate effective read surface. The capability application layer will join stored bindings with the current tenant-local model/tool catalogs, return only active bindings as effective, and surface deactivated/missing bindings as stale. Admin HTTP remains a thin adapter over this resolver so future runtime orchestration can consume the same contract.

**Tech Stack:** Go, Gin, in-memory catalogs for integration tests, `go test`

---

### Task 1: Effective Policy Domain and Resolver

**Files:**
- Create: `internal/domain/capability/effective_agent_capability_policy.go`
- Create: `internal/domain/capability/effective_agent_capability_policy_test.go`
- Create: `internal/application/capability/resolve_effective_agent_capability_policy.go`
- Create: `internal/application/capability/resolve_effective_agent_capability_policy_test.go`

- [ ] **Step 1: Write the failing test**

Add tests for:
- effective active model/tool IDs
- stale model/tool IDs after deactivation
- empty policy fallback for an existing profile

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$(pwd)/.cache/go-build go test ./internal/domain/capability ./internal/application/capability -run 'EffectiveAgentCapabilityPolicy|ResolveEffectiveAgentCapabilityPolicy' -count=1`
Expected: FAIL because the new domain type and resolver do not exist yet.

- [ ] **Step 3: Write minimal implementation**

Add the effective policy aggregate and a resolver handler that reads:
- existing agent profile
- stored capability policy
- current tenant-local model/tool catalogs

- [ ] **Step 4: Run test to verify it passes**

Run: `GOCACHE=$(pwd)/.cache/go-build go test ./internal/domain/capability ./internal/application/capability -run 'EffectiveAgentCapabilityPolicy|ResolveEffectiveAgentCapabilityPolicy' -count=1`
Expected: PASS

- [ ] **Step 5: Commit**

Run:

```bash
git add internal/domain/capability internal/application/capability docs/superpowers/plans/2026-03-26-phase1-effective-capability-policy-surface.md
git commit -m "feat: add phase1 effective capability policy resolver"
```

### Task 2: Admin API Effective Read Surface

**Files:**
- Modify: `internal/interfaces/http/router/admin.go`
- Modify: `test/integration/admin_control_plane_test.go`

- [ ] **Step 1: Write the failing test**

Add integration coverage for:
- `GET /api/admin/v1/agent-profiles/:profile_id/capability-policy/effective`
- active entries appearing under effective lists
- deactivated entries appearing under stale lists, not effective lists

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$(pwd)/.cache/go-build go test ./test/integration -run 'AdminEffectiveAgentCapabilityPolicy' -count=1`
Expected: FAIL because the route does not exist yet.

- [ ] **Step 3: Write minimal implementation**

Wire the new admin read route to the resolver handler.

- [ ] **Step 4: Run test to verify it passes**

Run: `GOCACHE=$(pwd)/.cache/go-build go test ./test/integration -run 'AdminEffectiveAgentCapabilityPolicy' -count=1`
Expected: PASS

- [ ] **Step 5: Commit**

Run:

```bash
git add internal/interfaces/http/router/admin.go test/integration/admin_control_plane_test.go
git commit -m "feat: add phase1 effective capability policy admin surface"
```

### Task 3: Docs, Version, and Full Verification

**Files:**
- Modify: `README.md`
- Modify: `CHANGELOG.md`
- Modify: `VERSION`
- Modify: `docs/phase-1-coverage-status.md`
- Modify: `docs/phase-handoff-playbook/2026-03-26-phase-1-stable-baseline.md`

- [ ] **Step 1: Update docs**

Document that Phase 1 now has a runtime-consumable effective capability read surface, while plugin registry, true runtime enforcement, workflow orchestration, and WebSocket protocol still remain pending.

- [ ] **Step 2: Run full verification**

Run: `GOCACHE=$(pwd)/.cache/go-build go test ./... -count=1`
Expected: PASS

- [ ] **Step 3: Commit**

Run:

```bash
git add README.md CHANGELOG.md VERSION docs
git commit -m "docs: sync phase1 effective capability policy status"
```

## Review Checklist

Before marking this slice complete, confirm all of the following:

- stored `capability-policy` semantics remain unchanged
- effective read semantics are additive, not breaking
- inactive or missing entries are surfaced as stale rather than silently disappearing without explanation
- no workflow engine or real runtime executor code is introduced
- docs still clearly mark plugin registry and runtime-side enforcement beyond read resolution as pending

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-03-26-phase1-effective-capability-policy-surface.md`.

This session is continuing inline in the existing `feature/phase1-control-plane` worktree with TDD and fresh verification before any completion claim.
