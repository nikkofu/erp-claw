# Phase 1 Capability Tenant Enablement Surface Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a Phase 1 tenant enablement baseline for capability catalog entries so model/tool entries can be activated or deactivated per tenant and future agent capability policy writes only accept active entries.

**Architecture:** Keep the slice inside the existing capability bounded context. Reuse the current model/tool catalog repositories, add minimal status helpers and application handlers for activate/deactivate semantics, and keep Admin HTTP as a thin adapter over those handlers. Existing agent capability policy save validation will be tightened so only tenant-local active entries can be bound.

**Tech Stack:** Go, Gin, PostgreSQL repository, in-memory bootstrap catalogs, `go test`

---

### Task 1: Status Lifecycle Contract

**Files:**
- Modify: `internal/domain/capability/model_catalog_entry.go`
- Modify: `internal/domain/capability/model_catalog_entry_test.go`
- Modify: `internal/domain/capability/tool_catalog_entry.go`
- Modify: `internal/domain/capability/tool_catalog_entry_test.go`
- Create: `internal/application/capability/set_model_catalog_entry_status.go`
- Create: `internal/application/capability/set_tool_catalog_entry_status.go`
- Create: `internal/application/capability/catalog_entry_status_test.go`

- [ ] **Step 1: Write the failing test**

Add domain/app tests for:
- active/inactive helpers
- activate/deactivate command behavior
- rejection when catalog entry is missing
- rejection when policy binding references inactive entries

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$(pwd)/.cache/go-build go test ./internal/domain/capability ./internal/application/capability -run 'CatalogEntryStatus|Inactive|Active' -count=1`
Expected: FAIL because lifecycle helpers and handlers do not exist yet.

- [ ] **Step 3: Write minimal implementation**

Add status helpers and minimal application handlers that locate tenant-local entries through the existing list seams and persist the new status through `Save`/`SaveTool`.

- [ ] **Step 4: Run test to verify it passes**

Run: `GOCACHE=$(pwd)/.cache/go-build go test ./internal/domain/capability ./internal/application/capability -run 'CatalogEntryStatus|Inactive|Active' -count=1`
Expected: PASS

- [ ] **Step 5: Commit**

Run:

```bash
git add internal/domain/capability internal/application/capability docs/superpowers/plans/2026-03-26-phase1-capability-tenant-enablement-surface.md
git commit -m "feat: add capability tenant enablement handlers"
```

### Task 2: Admin API and Integration Coverage

**Files:**
- Modify: `internal/interfaces/http/router/admin.go`
- Modify: `test/integration/admin_control_plane_test.go`

- [ ] **Step 1: Write the failing test**

Add integration coverage for:
- activating/deactivating model catalog entries
- activating/deactivating tool catalog entries
- capability policy binding rejection for inactive entries
- tenant-mutation rejection for unknown tenant scope
- explicit contract that deactivation does not rewrite existing capability bindings

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$(pwd)/.cache/go-build go test ./test/integration -run 'AdminCapabilityTenantEnablement|InactiveCapabilityPolicy' -count=1`
Expected: FAIL because activate/deactivate routes do not exist yet.

- [ ] **Step 3: Write minimal implementation**

Add `POST .../activate` and `POST .../deactivate` routes for model/tool catalog entries and wire them to the new application handlers.

- [ ] **Step 4: Run test to verify it passes**

Run: `GOCACHE=$(pwd)/.cache/go-build go test ./test/integration -run 'AdminCapabilityTenantEnablement|InactiveCapabilityPolicy' -count=1`
Expected: PASS

- [ ] **Step 5: Commit**

Run:

```bash
git add internal/interfaces/http/router/admin.go test/integration/admin_control_plane_test.go
git commit -m "feat: add capability tenant enablement admin surface"
```

### Task 3: Docs, Version, and Full Verification

**Files:**
- Modify: `README.md`
- Modify: `CHANGELOG.md`
- Modify: `VERSION`
- Modify: `docs/phase-1-coverage-status.md`
- Modify: `docs/phase-handoff-playbook/2026-03-26-phase-1-stable-baseline.md`

- [ ] **Step 1: Update docs**

Document that Phase 1 capability governance now includes tenant-local activation/deactivation baseline, while plugin registry, quota, feature flags, and runtime execution enforcement still remain pending.

- [ ] **Step 2: Run full verification**

Run: `GOCACHE=$(pwd)/.cache/go-build go test ./... -count=1`
Expected: PASS

- [ ] **Step 3: Commit**

Run:

```bash
git add README.md CHANGELOG.md VERSION docs
git commit -m "docs: sync phase1 capability tenant enablement status"
```

## Review Checklist

Before marking this slice complete, confirm all of the following:

- model/tool tenant enablement stays tenant-scoped
- Admin API supports explicit activate/deactivate lifecycle instead of only create/list
- inactive model/tool entries cannot be bound into agent capability policy
- no runtime executor or workflow orchestration logic is introduced
- docs still clearly mark plugin registry, quota, feature flag, and runtime-side enforcement as pending

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-03-26-phase1-capability-tenant-enablement-surface.md`.

This session is continuing in the existing `feature/phase1-control-plane` worktree with a subagent-assisted implementation flow and fresh verification before every completion claim.
