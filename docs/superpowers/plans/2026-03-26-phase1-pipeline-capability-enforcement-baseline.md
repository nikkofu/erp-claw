# Phase 1 Pipeline Capability Enforcement Baseline Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add an additive Phase 1 capability enforcement seam to the shared command pipeline so commands can be blocked before mutation when they request models/tools that are not effectively allowed for an agent profile.

**Architecture:** Keep enforcement opt-in and additive. The shared pipeline gets an optional capability authorizer dependency; when a command payload explicitly carries `agent_profile_id` plus requested `model_entry_id` / `tool_entry_ids`, the authorizer resolves the effective capability policy and rejects stale or unbound entries before approval-start or handler execution. This does not add a tool runtime, plugin registry, or workflow engine; it only creates the pre-execution gating seam that later runtime commands can reuse.

**Tech Stack:** Go, Gin-free application seam, shared pipeline, capability application handlers, `go test`

---

### Task 1: Shared Pipeline Capability Gate

**Files:**
- Modify: `internal/application/shared/pipeline.go`
- Modify: `internal/application/shared/pipeline_test.go`

- [ ] **Step 1: Write the failing test**

Add a pipeline unit test that proves:
- a configured capability authorizer runs before the handler
- a capability denial returns `ErrCapabilityDenied`
- the handler is not executed when capability authorization fails

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$(pwd)/.cache/go-build go test ./internal/application/shared -run 'Pipeline.*Capability' -count=1`
Expected: FAIL because the pipeline does not yet support a capability authorizer.

- [ ] **Step 3: Write minimal implementation**

Add:
- `CapabilityAuthorizer` interface
- `Capabilities` dependency on `PipelineDeps`
- optional authorizer invocation after policy evaluation but before approval/transaction execution
- `ErrCapabilityDenied` sentinel handling with a dedicated audit outcome

- [ ] **Step 4: Run test to verify it passes**

Run: `GOCACHE=$(pwd)/.cache/go-build go test ./internal/application/shared -run 'Pipeline.*Capability' -count=1`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/application/shared/pipeline.go internal/application/shared/pipeline_test.go
git commit -m "feat: add phase1 pipeline capability guard"
```

### Task 2: Capability Pipeline Adapter

**Files:**
- Create: `internal/application/capability/pipeline_adapter.go`
- Create: `internal/application/capability/pipeline_adapter_test.go`
- Modify: `internal/application/capability/repository_fake_test.go`

- [ ] **Step 1: Write the failing test**

Add adapter tests that prove:
- effective model/tool requests are allowed
- stale model/tool requests return `shared.ErrCapabilityDenied`
- commands without explicit capability request keys are ignored
- commands with capability request keys but missing `agent_profile_id` are rejected

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$(pwd)/.cache/go-build go test ./internal/application/capability -run 'SharedCommandCapabilityAuthorizer' -count=1`
Expected: FAIL because the adapter does not exist yet.

- [ ] **Step 3: Write minimal implementation**

Add a `SharedCommandCapabilityAuthorizer` that:
- accepts a `ResolveEffectiveAgentCapabilityPolicyHandler`
- reads map payload keys `agent_profile_id`, `model_entry_id`, and `tool_entry_ids`
- skips commands that do not explicitly request a model/tool capability
- rejects missing/invalid `agent_profile_id`
- resolves the effective policy and returns `shared.ErrCapabilityDenied` for stale or unbound requested entries

- [ ] **Step 4: Run test to verify it passes**

Run: `GOCACHE=$(pwd)/.cache/go-build go test ./internal/application/capability -run 'SharedCommandCapabilityAuthorizer' -count=1`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/application/capability/pipeline_adapter.go internal/application/capability/pipeline_adapter_test.go internal/application/capability/repository_fake_test.go
git commit -m "feat: add phase1 capability pipeline adapter"
```

### Task 3: Integration, Docs, And Release Sync

**Files:**
- Modify: `test/integration/command_pipeline_test.go`
- Modify: `test/integration/approval_pipeline_integration_test.go`
- Modify: `README.md`
- Modify: `CHANGELOG.md`
- Modify: `VERSION`
- Modify: `docs/phase-1-coverage-status.md`
- Modify: `docs/phase-handoff-playbook/2026-03-26-phase-1-stable-baseline.md`

- [ ] **Step 1: Write the failing test**

Add integration coverage that proves:
- an allowed command with explicit `agent_profile_id` / `model_entry_id` / `tool_entry_ids` passes through the pipeline
- a stale or unbound request is blocked with `shared.ErrCapabilityDenied`
- approval is not started when capability authorization fails before the approval path

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$(pwd)/.cache/go-build go test ./test/integration -run 'CommandPipeline.*Capability|Pipeline.*Capability.*Approval' -count=1`
Expected: FAIL because the pipeline is not yet capability-aware end-to-end.

- [ ] **Step 3: Write minimal implementation**

Wire the integration tests to the new capability authorizer, then update docs and versioning to describe this slice as a pipeline-side enforcement baseline rather than full tool runtime enforcement.

- [ ] **Step 4: Run test to verify it passes**

Run: `GOCACHE=$(pwd)/.cache/go-build go test ./test/integration -run 'CommandPipeline.*Capability|Pipeline.*Capability.*Approval' -count=1`
Expected: PASS

- [ ] **Step 5: Run full verification**

Run: `GOCACHE=$(pwd)/.cache/go-build go test ./... -count=1`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add test/integration/command_pipeline_test.go test/integration/approval_pipeline_integration_test.go README.md CHANGELOG.md VERSION docs
git commit -m "feat: add phase1 pipeline capability enforcement baseline"
```

## Review Checklist

Before marking this slice complete, confirm all of the following:

- capability enforcement is opt-in and additive; commands without explicit capability request keys still behave as before
- enforcement uses explicit `capability-policy` bindings resolved through the effective policy surface; it does not infer `agent_profile.model`
- capability denial happens before approval creation and before handler execution
- no plugin registry, tool executor, or workflow engine is introduced in this slice
- docs still clearly mark full runtime-side tool execution and plugin registry as pending

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-03-26-phase1-pipeline-capability-enforcement-baseline.md`.

This session continues inline on `feature/phase1-control-plane` with subagent-driven development discipline, but the implementation will be executed locally on the critical path to keep the current slice moving.
