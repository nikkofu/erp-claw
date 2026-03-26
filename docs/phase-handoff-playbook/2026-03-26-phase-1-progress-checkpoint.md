# Phase 1 Progress Checkpoint

更新时间：2026-03-26

## 1. 今日完成

### 1.1 仓库边界核对

- 已确认 root 仓库误落的 `capability` 文件与 worktree 对应文件当前内容一致。
- 真实的 Phase 1 开发主线仍以 `phase1-control-plane` worktree 为准。
- root `main` 仍保留一份误落的未跟踪 capability 文件，后续需要单独清理，但今天没有做任何破坏性操作。

### 1.2 Tenant / IAM 第二批切片

已完成从 migration 到 Admin API 的最小闭环：

- `000007_phase1_tenant_iam_extension`
- `Role`、`Department`、`UserRoleBinding`、`UserDepartmentBinding` 领域对象
- control-plane repository seam 扩展
- Postgres repository 实现
- in-memory control-plane catalog 回退实现
- application command/query handlers：
  - `CreateRole`
  - `ListRoles`
  - `CreateDepartment`
  - `ListDepartments`
  - `AssignUserRole`
  - `AssignUserDepartment`
- Admin API endpoints：
  - `POST /api/admin/v1/roles`
  - `GET /api/admin/v1/roles`
  - `POST /api/admin/v1/departments`
  - `GET /api/admin/v1/departments`
  - `POST /api/admin/v1/user-role-bindings`
  - `POST /api/admin/v1/user-department-bindings`

### 1.3 Capability / Tool Catalog 切片

已完成 tool catalog baseline：

- `000008_phase1_tool_catalog`
- `ToolCatalogEntry` 领域对象
- tool catalog repository contract
- application create/list handlers
- Postgres repository `SaveTool` / `ListToolsByTenant`
- migration / domain / application / repository tests

### 1.4 文档同步

- 已更新 `docs/phase-1-coverage-status.md`
- 口径已经反映 Tenant/IAM 与 Tool Catalog 的最新进度

### 1.5 Runtime Read Side 切片

已完成 runtime read side 的最小 query/replay 基线：

- `internal/application/agentruntime/read_side.go`
- `ListSessions`
- `ListTasks`
- `ReplayWorkspaceEvents`
- `WorkspaceGateway.ListWorkspaceEvents`
- in-memory integration read side 验证
- Postgres `AgentRuntimeRepository` 的 `ListSessions` / `ListTasks` seam

### 1.6 Approval Baseline 切片

已完成 approval/workflow 的最小审批闭环：

- `000009_phase1_approval_baseline`
- `approval_definition` / `approval_instance` / `approval_task`
- `internal/domain/approval/*`
- `internal/application/approval/*`
- `internal/infrastructure/persistence/postgres/approval_repository.go`
- `REQUIRE_APPROVAL` 到审批实例创建的最小 pipeline 接线
- start / approve / reject handler tests
- approval migration / pipeline integration tests

### 1.7 Workspace HTTP Surface 切片

已完成 workspace 最小只读接口：

- `GET /api/workspace/v1/sessions`
- `GET /api/workspace/v1/tasks`
- `GET /api/workspace/v1/events`
- `internal/bootstrap/agent_runtime_catalog.go`
- `internal/interfaces/http/router/workspace.go`
- `test/integration/workspace_api_test.go`

## 2. Fresh 验证

今天完成后 fresh 执行：

```bash
GOCACHE=$(pwd)/.cache/go-build go test ./... -count=1
```

结果：全绿，通过。

## 3. 当前仍未完成的 Phase 1 主线

### 3.1 P0 / 工程边界

- root `main` 上误落 capability 文件的清理动作尚未执行
- 大量 worktree 改动仍未按逻辑切片整理提交边界

### 3.2 P1 / Capability Governance

- model/tool policy binding 草案
- tenant enablement / plugin registry
- quota / feature flag

### 3.3 P1 / Workspace & Runtime Read Side

- workspace 命令写入面
- 更完整的 replay 协议
- 真实 WebSocket / streaming protocol
- runtime live integration

### 3.4 P1 / Approval / Workflow

- approval HTTP / admin surface
- 多级审批与更复杂的 approver resolution
- workflow orchestration / pause-resume

### 3.5 P1 / Reliability / Live DB

- policy audit live Postgres round-trip
- agent runtime live Postgres tests
- capability live Postgres tests
- outbox live Postgres tests
- DLQ / replay operator / consumer idempotency

## 4. 下一步推荐顺序

1. `T00-02` 明确当前 worktree 的 commit slicing 方案
2. `T02-06` 补 model/tool policy binding contract
3. `T05-*` 做 live DB round-trip 与 reliability hardening
4. 补 approval 管理面与更完整 workflow orchestration
5. 补 workspace 写入面和真实 streaming protocol

## 5. 结论

今天的推进结果不是 “Phase 1 完成”，而是：

- Tenant / IAM 从第一批 catalog 基线推进到了可操作的 role/department/binding 闭环
- Capability governance 从 model catalog 推进到了 model + tool catalog 双基线
- Runtime read side 已经补到 application/gateway/postgres 的最小 query/replay 闭环
- Approval baseline 已经补到 domain/application/postgres/pipeline 的最小闭环
- Workspace HTTP surface 已经从空壳推进到最小只读接口
- Phase 1 下一批高优先级焦点已经进一步收敛到 `policy binding`、`live DB reliability`、更完整的 workflow orchestration，以及 workspace 写入面/streaming protocol
