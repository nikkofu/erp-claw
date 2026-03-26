# Phase 1 明日继续工作清单与交接总结

更新时间：2026-03-25

## 1. 文档目的

这份文档同时承担 4 个角色：

- 保存今天这轮 Phase 1 开发后的真实状态
- 形成明天可以直接接着做的最小颗粒度任务清单
- 总结今天这轮并行开发的经验、教训和反思
- 作为后续继续推进 Phase 1 控制面主线的操作指南

本文只描述当前 `phase1-control-plane` 工作树的状态，不把 root `main` 的状态误判成当前开发主线。

## 2. 当前快照

### 2.1 工作位置

- 工作树路径：`<project-root>/.worktrees/phase1-control-plane`
- 当前分支：`feature/phase1-control-plane`
- 代码状态：脏工作区，包含大量未提交的 Phase 1 增量实现

### 2.2 已完成的本轮切片

这轮并行补齐了 4 个关键子项目：

- Governance Core 增量：
  - 策略规则持久化查询
  - 策略规则启停生命周期
  - 治理命令处理器
  - 审计持久化查询基线
- Agent Runtime Control 增量：
  - session/task 仓储与状态流转
  - `CloseSession`
  - `FailTask`
  - `CancelTask`
  - workspace runtime event seam
- Capability Governance 基线：
  - tenant-scoped model catalog baseline
  - `000005_phase1_capability_governance`
  - capability domain/application/postgres repository
- Reliability And Hardening 增量：
  - outbox dispatcher / retry / failed terminal state
  - worker poll observability seam
  - failed outbox operator recovery

### 2.3 Fresh 验证结果

本轮结束前已经 fresh 执行：

```bash
GOCACHE=$(pwd)/.cache/go-build go test ./... -count=1
```

结果：全绿，通过。

### 2.4 仍然未完成的 Phase 1 主线

虽然控制面已经不是纯骨架，但 `Phase 1` 依然明确未完成，剩余缺口主要集中在：

- 真正的 Tenant / IAM 实体化
- Plugin / Tool Registry
- quota / feature flag / capability governance 剩余部分
- workspace 真正的流式协议与事件回放
- approval / workflow 基线
- live Postgres 集成验证与更完整的运维可靠性

## 3. 明日开工前必做事项

这 5 步不要跳过。

1. 进入工作树，不要在 root `main` 上继续开发。
2. 先看两个状态：
   - 在当前工作树执行：`git status --short`
   - 在项目根目录执行：`git -C <project-root> status --short`
3. 再跑一次 fresh 验证：
   - `GOCACHE=$(pwd)/.cache/go-build go test ./... -count=1`
4. 明确今天误落在 root `main` 的 capability 文件如何处理，不要让 root/worktree 双份实现继续漂移。
5. 从下面 `P0/P1` 清单里只选一个工作流开始，不要并行改动互相耦合的文件组。

## 4. 明日推荐顺序

推荐顺序不是“想到哪写到哪”，而是：

1. 先做仓库卫生和集成边界确认
2. 再做 Tenant / IAM
3. 再做 Tool Registry / Capability 剩余部分
4. 再做 Workspace / Runtime read side
5. 再做 Approval / Workflow
6. 最后补 live DB integration 和更深的可靠性

原因很直接：

- Tenant / IAM 是后续所有控制面权限、审批和工具授权的前置条件
- Tool Registry 依赖 tenant / actor / policy 语义
- Approval / Workflow 如果先做，很容易围绕占位 actor 和权限模型空转

## 5. 最小颗粒度任务清单

下面的任务按“能直接开做”的颗粒度拆分。每个任务都给出目的、验收口径、建议触碰文件和并行建议。

### 5.1 P0 仓库卫生与边界确认

| ID | 优先级 | 任务 | 详细说明 | 验收口径 | 建议文件 | 并行性 |
| --- | --- | --- | --- | --- | --- | --- |
| T00-01 | P0 | 清点 root `main` 上误落的 capability 文件 | 今天有一次 capability 实现误写到 root `main`，必须先明确哪些文件只保留在工作树，哪些需要后续迁移或删除。 | 形成一条明确结论：`以工作树为准`，并列出 root 侧待清理文件清单。 | root 与 worktree 下的 `internal/domain/capability/*`、`internal/application/capability/*`、`migrations/000005*`、`test/integration/capability_governance_test.go` | 不并行 |
| T00-02 | P0 | 按切片整理提交边界 | 当前工作树是大量未提交改动，后续如果继续堆开发，回滚和审查都会越来越困难。 | 形成提交计划，至少按 `controlplane`、`governance`、`agentruntime`、`capability`、`outbox` 5 个逻辑切片整理。 | `git status`、相关目录 | 可与文档整理并行，不建议与新功能并行 |
| T00-03 | P0 | 确认文档与代码口径一致 | `phase-1-coverage-status.md` 已更新，但后续开发必须以这份文档为准，不再按旧“纯骨架”口径判断。 | 重新通读 [phase-1-coverage-status.md](../phase-1-coverage-status.md) 并确认没有新的事实遗漏。 | `docs/phase-1-coverage-status.md`、`docs/phase-roadmap-overview.md` | 可并行 |

### 5.2 P1 Tenant / IAM 实体化

| ID | 优先级 | 任务 | 详细说明 | 验收口径 | 建议文件 | 并行性 |
| --- | --- | --- | --- | --- | --- | --- |
| T01-01 | P1 | 写角色/部门迁移契约测试 | 先把 `iam_role`、`iam_department`、`iam_user_role`、`iam_user_department` 的表结构契约固定下来。 | 先看到 migration contract test RED，再看到 GREEN。 | `test/integration/*tenant*` 或新增 `tenant_iam_migration_test.go` | 可与 T02 系列并行 |
| T01-02 | P1 | 增加 Tenant/IAM 扩展 migration | 新增一版 migration，补真实角色、部门和用户关联结构。 | `migrate` 可编译，迁移契约测试通过。 | `migrations/000007_*` | 依赖 T01-01 |
| T01-03 | P1 | 落 `Role` 领域对象 | 先把角色定义、名称、租户边界和必要校验固化。 | `internal/domain/*` 单测通过。 | `internal/domain/controlplane/*` 或拆新文件 | 依赖 T01-02 可弱依赖 |
| T01-04 | P1 | 落 `Department` 领域对象 | 补部门定义、层级或父子关系的最小校验。 | `internal/domain/*` 单测通过。 | `internal/domain/controlplane/*` | 可与 T01-03 并行 |
| T01-05 | P1 | 增加 user-role / user-department repository seam | 把用户和角色/部门的绑定从“概念”变成 repository contract。 | 仓储接口、内存测试替身和 Postgres compile test 通过。 | `internal/domain/controlplane/*`、`internal/infrastructure/persistence/postgres/*` | 依赖 T01-03/T01-04 |
| T01-06 | P1 | 增加 role/department create/list handlers | 给控制面补最小 command/query seams，为后面 admin API 接线做准备。 | handler 单测通过。 | `internal/application/controlplane/command/*`、`query/*` | 依赖 T01-05 |
| T01-07 | P1 | 增加 admin 集成测试 | 只做最小 happy path：创建角色、创建部门、绑定用户。 | targeted integration tests 通过。 | `test/integration/*control_plane*` | 依赖 T01-06 |

### 5.3 P1 Capability Governance 剩余部分

| ID | 优先级 | 任务 | 详细说明 | 验收口径 | 建议文件 | 并行性 |
| --- | --- | --- | --- | --- | --- | --- |
| T02-01 | P1 | 写 tool catalog migration 契约测试 | 现在只有 model catalog，没有 tool catalog。先把表结构契约固定下来。 | RED/GREEN 明确。 | `test/integration/tool_catalog_migration_test.go` | 可与 T01 系列并行 |
| T02-02 | P1 | 增加 tool catalog migration | 补 `tool_catalog_entries` 或等价结构，至少覆盖 tenant、tool key、risk、status。 | migration contract 通过。 | `migrations/000008_*` | 依赖 T02-01 |
| T02-03 | P1 | 落 ToolCatalogEntry 领域对象 | 把工具目录从“设计概念”变成可验证的 domain object。 | domain tests 通过。 | `internal/domain/capability/*` | 依赖 T02-02 |
| T02-04 | P1 | 增加 ToolCatalogRepository | 与 model catalog 风格保持一致，先做最小存取。 | postgres package tests 通过。 | `internal/infrastructure/persistence/postgres/*capability*` | 依赖 T02-03 |
| T02-05 | P1 | 增加 tool create/list handlers | 做成和 model catalog 类似的最小 application seam。 | application tests 通过。 | `internal/application/capability/*` | 依赖 T02-04 |
| T02-06 | P1 | 增加 model/tool policy binding 草案 | 为后面 Agent 工具授权预留结构，不要直接耦合 runtime。 | 至少形成一个 domain contract 和测试。 | `internal/domain/capability/*`、可能补 `internal/platform/policy/*` | 建议单独做，不与 T03 并行 |

### 5.4 P1 Workspace / Runtime Read Side

| ID | 优先级 | 任务 | 详细说明 | 验收口径 | 建议文件 | 并行性 |
| --- | --- | --- | --- | --- | --- | --- |
| T03-01 | P1 | 定义 workspace event query seam | 现在只有 append/broadcast，没有 query/replay contract。 | query contract + 单测通过。 | `internal/platform/runtime/*` 或 `internal/application/agentruntime/*` | 可与 T02 系列并行 |
| T03-02 | P1 | 增加 session/task list query handlers | 当前有状态机，没有真正的 read side handler。 | query handler tests 通过。 | `internal/application/agentruntime/*` | 依赖 T03-01 |
| T03-03 | P1 | 增加 workspace event replay handler | 最小支持按 sessionKey 拉取历史事件。 | query tests 通过。 | `internal/application/agentruntime/*`、可能补 repository seam | 依赖 T03-01 |
| T03-04 | P1 | 落 SSE 或最小 WebSocket replay 协议 | 不要先做复杂协议，只要能订阅和回放。 | 接口集成测试通过。 | `internal/interfaces/ws/*`、可能补 `internal/interfaces/http/router/workspace.go` | 依赖 T03-02/T03-03 |
| T03-05 | P1 | 增加 runtime live integration | 用真实 repo + gateway 验证 create/start/fail/cancel/close + replay。 | integration tests 通过。 | `test/integration/*runtime*` | 依赖 T03-04 |

### 5.5 P1 Approval / Workflow 基线

| ID | 优先级 | 任务 | 详细说明 | 验收口径 | 建议文件 | 并行性 |
| --- | --- | --- | --- | --- | --- | --- |
| T04-01 | P1 | 写 approval migration 契约测试 | 固定 `approval_definition`、`approval_instance`、`approval_task` 这类基础结构。 | RED/GREEN 清晰。 | `test/integration/approval_migration_test.go` | 建议在 T01 基础稳定后开始 |
| T04-02 | P1 | 增加 approval migration | 落最小结构，不要一开始就做复杂引擎。 | migration tests 通过。 | `migrations/000009_*` | 依赖 T04-01 |
| T04-03 | P1 | 定义 approval domain object | 先有状态定义、发起人、审批人、状态流转。 | domain tests 通过。 | `internal/domain/approval/*` 或等价目录 | 依赖 T04-02 |
| T04-04 | P1 | 增加 start approval handler | 让审批实例能被创建。 | application tests 通过。 | `internal/application/approval/*` | 依赖 T04-03 |
| T04-05 | P1 | 增加 approve/reject handler | 最小人工任务闭环。 | approve/reject tests 通过。 | `internal/application/approval/*` | 依赖 T04-04 |
| T04-06 | P1 | 把 `REQUIRE_APPROVAL` 与审批实例接上 | 这一步才把 policy decision 和 approval 关联起来。 | integration scenario 通过。 | `internal/application/shared/pipeline.go`、`internal/platform/policy/*`、`test/integration/*approval*` | 不建议与其它 pipeline 变更并行 |

### 5.6 P1 Live DB 集成与可靠性补强

| ID | 优先级 | 任务 | 详细说明 | 验收口径 | 建议文件 | 并行性 |
| --- | --- | --- | --- | --- | --- | --- |
| T05-01 | P1 | 给 `PolicyAuditRepository` 增加 live Postgres round-trip tests | 当前主要是 compile/unit seam，缺真实 DB round-trip。 | compose 环境下 targeted test 通过。 | `internal/infrastructure/persistence/postgres/*policy*`、`test/integration/*governance*` | 可单独并行 |
| T05-02 | P1 | 给 `AgentRuntimeRepository` 增加 live Postgres tests | 覆盖 create/get/update/status flow。 | targeted test 通过。 | `internal/infrastructure/persistence/postgres/agent_runtime_repository_test.go` | 可单独并行 |
| T05-03 | P1 | 给 `CapabilityRepository` 增加 live Postgres tests | 覆盖 save/listByTenant。 | targeted test 通过。 | `internal/infrastructure/persistence/postgres/capability_repository_test.go` | 可单独并行 |
| T05-04 | P1 | 给 `OutboxRepository` 增加 live Postgres tests | 覆盖 fetch/mark retry/fail/requeue。 | targeted test 通过。 | `internal/infrastructure/persistence/postgres/outbox_repository_test.go` | 可单独并行 |
| T05-05 | P1 | 定义 DLQ 或 replay operator 策略 | 现在有 failed recovery，但还没有更完整的失败运营方案。 | 至少形成设计说明或小型 plan。 | `docs/*` 或 `internal/application/shared/outbox/*` | 可与 DB tests 并行 |
| T05-06 | P1 | 增加 consumer-side idempotency 设计/基线 | outbox 端已有提升，但消费端幂等仍是缺口。 | 形成设计或最小 contract。 | `internal/platform/eventbus/*`、`docs/*` | 建议晚于 T05-04 |

## 6. 明天最值得先做的 8 个任务

如果明天只打算推进一整天，建议优先做下面 8 个：

1. `T00-01` 清点 root/worktree capability 漂移
2. `T00-02` 整理提交边界
3. `T01-01` Tenant/IAM migration contract tests
4. `T01-02` Tenant/IAM migration
5. `T01-03` Role domain object
6. `T01-04` Department domain object
7. `T01-05` user-role / user-department repository seam
8. `T01-06` role/department create/list handlers

这是最合理的顺序，因为：

- 先把仓库边界搞清楚，避免在错误工作区里继续写
- 再把 Tenant / IAM 铺开，后续 capability、approval、tool registry 才不会建立在空 actor 模型上

## 7. 今天的经验总结

### 7.1 有效的方法

- 工作树隔离是对的。
- 4 个子代理并行也是对的。
- 绝对路径 + 明确 write scope 极其关键。
- 先跑 targeted tests，再跑 `go test ./...`，节奏是对的。
- 在阶段中途更新状态文档，而不是最后才补文档，能减少口径漂移。

### 7.2 真正帮到开发速度的做法

- 不是“多开代理”本身，而是“多开彼此不碰写入范围的代理”。
- 不是“让代理自己探索全部上下文”，而是提前给文件范围、目标和不允许修改的边界。
- 不是“看到代理说 DONE 就信”，而是统一回到主线程做 fresh 验证。

### 7.3 今天踩到的坑

- capability 这一条线第一次写到了 root `main`，不是工作树。
- 1C 的计划草案和最后实际落地结构不完全一致，后面才补了 plan note。
- 如果不及时更新 `phase-1-coverage-status.md`，就会继续按过时口径误判“只剩骨架”。
- 并行开发时，任何一个切片没有明确绝对路径，都会把上下文带偏。

## 8. 深刻反思

### 8.1 最大的问题不是代码，而是“工作上下文漂移”

今天最值得警惕的不是某个函数写错，而是 capability 任务一度脱离了正确工作树。这个问题说明：在多工作树、多分支、多子代理同时存在的情况下，默认上下文根本不可靠。以后任何并行任务都必须：

- 在提示里写绝对工作树路径
- 明确“不能写 root repo”
- 明确“不要触碰哪些文件”

否则不是实现慢，而是实现会落到错误位置，后面清理成本更高。

### 8.2 “计划存在”不等于“计划仍然正确”

1C 最后是按 root baseline 复刻到工作树，而不是完全按最初计划里的命名和分层落地。这件事说明：计划是起点，不是圣经。如果实现过程中发现已有事实上的基线已经存在，应该尽快把计划同步，而不是等到最后再解释差异。否则后面接手的人会同时看到两套口径。

### 8.3 并行开发真正的瓶颈是“集成与判断”

多代理确实能加快切片开发，但加速的前提是主线程持续做三件事：

- 收敛范围
- 审核结果
- fresh 验证

如果主线程只负责“转发消息”，那并行只会制造更多漂移；如果主线程持续收敛边界和口径，并行才真正有价值。

### 8.4 文档不是附属物，而是控制复杂度的基础设施

今天如果没有同步更新状态文档，明天继续时就很可能还会问“Phase 1 到底是不是只完成了骨架”。文档在这里不是记录，而是降低认知回放成本的基础设施。以后每完成一轮明显改变阶段判断的工作，都必须同步更新：

- 覆盖状态文档
- 阶段总览文档
- 当前切片计划文档

## 9. 明天继续时禁止做的事

- 不要直接在 root `main` 上继续开发
- 不要在 Tenant / IAM 没展开前就大做 Approval / Workflow
- 不要先做复杂多代理编排或高级检索
- 不要先接复杂前端界面
- 不要跳过 fresh `go test ./...`
- 不要把多个会改同一组文件的任务再并行化

## 10. 结论

今天这轮工作的正确结论不是“Phase 1 做完了”，而是：

- 平台底座已经稳住
- Phase 1 控制面已经从“纯骨架”进入“第一批可执行切片”
- 明天应继续补 `Tenant / IAM`，然后再推进 `Tool Registry / Workspace / Approval`

如果明天只做一件真正重要的事，那就是先把 `Tenant / IAM` 这条线拉起来。
