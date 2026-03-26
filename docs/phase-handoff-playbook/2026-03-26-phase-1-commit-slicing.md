# Phase 1 Commit Slicing Plan

更新时间：2026-03-26

## 1. 文档目的

这份文档不是实际 git 提交记录，而是当前 `phase1-control-plane` 工作树应当整理成的稳定提交清单。

它的用途有两个：

- 为当前模型后续整理 Phase 1 稳定版本提供明确边界
- 为 Claude 并行启动 Phase 2 工作提供可靠的交付契约，避免在脏工作区和未切片改动上协作

## 2. 切片原则

- 每个 commit 必须只覆盖一个清晰能力面
- 每个 commit 都必须能独立通过对应 targeted tests
- 优先按依赖顺序切，不按文件夹随意分堆
- 文档提交应贴近对应实现，不要全部堆到最后

## 3. 推荐提交清单

### C01 平台基础与运行骨架

范围：

- `cmd/api-server`
- `cmd/agent-gateway`
- `cmd/worker`
- `cmd/scheduler`
- `cmd/migrate`
- `internal/bootstrap/*`
- `configs/*`
- `docker-compose.yml`
- `Makefile`
- 基础健康检查与 smoke 流程

验收：

- `go test ./... -run TestAPIHealth...` 或等价基础验证

### C02 控制面第一批目录基线

范围：

- `migrations/000002_init_control_plane_catalog.*`
- `internal/domain/controlplane` 中 tenant/user/agent profile 初始切片
- `internal/application/controlplane` 中 tenant/user/agent profile create/list
- `internal/infrastructure/persistence/postgres/controlplane_repository.go`
- `internal/interfaces/http/router/admin.go` 的第一批 admin catalog 路由

验收：

- `go test ./test/integration -run 'Test(ControlPlaneMigrationContainsCatalogTables|AdminCreateTenantRoute)' -count=1`

### C03 治理核心

范围：

- `migrations/000003_phase1_governance_core.*`
- `internal/platform/policy/*`
- `internal/platform/audit/*`
- `internal/application/governance/*`
- `internal/infrastructure/persistence/postgres/policy_audit_repository.go`

验收：

- `go test ./internal/platform/policy ./internal/platform/audit ./internal/application/governance/... -count=1`
- `go test ./test/integration -run 'TestGovernance.*' -count=1`

### C04 Agent Runtime Control

范围：

- `migrations/000004_phase1_agent_runtime_control.*`
- `internal/domain/agentruntime/*`
- `internal/application/agentruntime/service.go`
- `internal/infrastructure/persistence/postgres/agent_runtime_repository.go`
- `internal/interfaces/ws/workspace_gateway.go` 的广播基线

验收：

- `go test ./internal/application/agentruntime ./internal/interfaces/ws ./internal/infrastructure/persistence/postgres -count=1`
- `go test ./test/integration -run 'TestAgentRuntime(Control|Failure).*' -count=1`

### C05 Capability Governance Baseline

范围：

- `migrations/000005_phase1_capability_governance.*`
- `internal/domain/capability/model_*`
- `internal/application/capability/*model*`
- `internal/infrastructure/persistence/postgres/capability_repository.go` 中 model catalog 部分

验收：

- `go test ./internal/domain/capability ./internal/application/capability ./internal/infrastructure/persistence/postgres -count=1`
- `go test ./test/integration -run TestCapabilityGovernanceMigrationDefinesModelCatalogTable -count=1`

### C06 Reliability Hardening

范围：

- `migrations/000006_phase1_reliability_hardening.*`
- `internal/application/shared/outbox/*`
- `internal/infrastructure/persistence/postgres/outbox_repository.go`
- `internal/infrastructure/observability/outbox_poll.go`
- `cmd/worker/main.go`

验收：

- `go test ./internal/application/shared/outbox ./internal/infrastructure/observability ./internal/infrastructure/persistence/postgres -count=1`
- `go test ./test/integration -run 'TestOutbox.*' -count=1`

### C07 Tenant IAM Expansion

范围：

- `migrations/000007_phase1_tenant_iam_extension.*`
- `internal/domain/controlplane/iam.go`
- `internal/application/controlplane/command/create_role.go`
- `internal/application/controlplane/command/create_department.go`
- `internal/application/controlplane/command/assign_user_role.go`
- `internal/application/controlplane/command/assign_user_department.go`
- `internal/application/controlplane/query/list_roles.go`
- `internal/application/controlplane/query/list_departments.go`
- `internal/bootstrap/control_plane_catalog.go`
- `internal/infrastructure/persistence/postgres/controlplane_repository.go`
- `internal/interfaces/http/router/admin.go`

验收：

- `go test ./internal/domain/controlplane ./internal/application/controlplane/... ./internal/infrastructure/persistence/postgres -count=1`
- `go test ./test/integration -run 'Test(TenantIAMExtensionMigrationContainsDepartmentAndBindingTables|AdminRoleDepartmentLifecycleRoutes)' -count=1`

### C08 Tool Catalog Expansion

范围：

- `migrations/000008_phase1_tool_catalog.*`
- `internal/domain/capability/tool_*`
- `internal/application/capability/*tool*`
- `internal/infrastructure/persistence/postgres/capability_repository.go` 中 tool catalog 部分

验收：

- `go test ./internal/domain/capability ./internal/application/capability ./internal/infrastructure/persistence/postgres -count=1`
- `go test ./test/integration -run TestCapabilityGovernanceMigrationDefinesToolCatalogTable -count=1`

### C09 Runtime Read Side

范围：

- `internal/application/agentruntime/read_side.go`
- `internal/infrastructure/persistence/postgres/agent_runtime_repository.go` 中 list seam
- `internal/interfaces/ws/workspace_gateway.go` 中 replay/history seam
- `internal/bootstrap/agent_runtime_catalog.go`
- `internal/interfaces/http/router/workspace.go`
- `test/integration/agent_runtime_read_side_test.go`
- `test/integration/workspace_api_test.go`

验收：

- `go test ./internal/application/agentruntime ./internal/interfaces/ws ./internal/infrastructure/persistence/postgres -count=1`
- `go test ./test/integration -run 'Test(AgentRuntimeReadSideListsSessionsTasksAndReplaysEvents|WorkspaceRoutesExposeSessionsTasksAndEvents)' -count=1`

### C10 Approval Baseline

范围：

- `migrations/000009_phase1_approval_baseline.*`
- `internal/domain/approval/*`
- `internal/application/approval/*`
- `internal/infrastructure/persistence/postgres/approval_repository.go`
- `internal/application/shared/pipeline.go` 中 approval starter seam
- `test/integration/approval_*`

验收：

- `go test ./internal/domain/approval ./internal/application/approval ./internal/infrastructure/persistence/postgres -count=1`
- `go test ./test/integration -run 'Test(ApprovalBaselineMigrationContainsDefinitionInstanceAndTaskTables|PipelineRequireApprovalStartsApprovalInstanceAndTask)' -count=1`

### C11 文档与阶段状态同步

范围：

- `docs/phase-1-coverage-status.md`
- `docs/phase-roadmap-overview.md`
- `docs/phase-handoff-playbook/*`

验收：

- 文档内容与代码快照一致
- 不包含绝对工作目录泄露

## 4. 给 Claude 的协作契约

当 Phase 2 交给 Claude 并行推进时，推荐以下工作方式：

- Claude 不要直接基于当前脏工作区做 Phase 2
- 先以 Phase 1 稳定 commit 清单为基础建立独立 worktree 或分支
- Phase 2 只消费已经稳定切好的 Phase 1 能力，不反向改动它们的接口约定

Claude 在启动 Phase 2 前，应默认已存在并可依赖的 Phase 1 契约：

- tenant / iam 最小控制面
- governance rule / audit query baseline
- agent runtime control + read side baseline
- model / tool catalog baseline
- approval baseline + `REQUIRE_APPROVAL` starter seam
- outbox reliability baseline

## 5. 当前最值得先整理成稳定提交的 4 个切片

如果只优先整理最关键的稳定版本，建议先切：

1. `C03 Governance Core`
2. `C04 Agent Runtime Control`
3. `C07 Tenant IAM Expansion`
4. `C10 Approval Baseline`

原因：

- 这 4 个切片共同构成了 Phase 1 控制面最关键的协作契约
- Claude 做 Phase 2 时，最需要依赖的是治理、审批、runtime 和 IAM 的稳定边界
