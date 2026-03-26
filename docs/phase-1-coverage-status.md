# Phase 1 覆盖领域、模块、功能清单、优先级与实现情况

更新时间：2026-03-26

## 1. 文档目的

本文用于把当前项目的 Phase 1 范围整理成一份可以持续维护的执行视图，回答四个问题：

- Phase 1 设计上覆盖哪些领域与模块
- 这些领域与模块的功能边界是什么
- 它们在当前阶段的优先级如何排序
- 代码仓库里已经实现到了什么程度

本文的判断基于三个来源：

- 设计蓝图：`docs/superpowers/specs/2026-03-25-agentic-ai-native-erp-design.md`
- 平台基础实施计划：`docs/superpowers/plans/2026-03-25-platform-foundation-implementation-plan.md`
- 当前代码快照：`cmd/`、`internal/`、`migrations/`、`scripts/`、`test/`

## 2. 口径说明

### 2.1 Phase 1 的执行口径

当前仓库已经完成的是“Phase 0/1 平台底座切片”，不是“Phase 1 全量业务闭环”。

更准确地说，当前状态是：

- 平台底座已经具备可运行骨架
- 控制面已经落地 tenant/user/role/department/agent profile 目录、user-role/user-department 绑定、policy/audit 持久化查询与最小 Admin 管理面，以及 capability 的 model/tool catalog 基线与最小 Admin 管理面
- Agent 执行与工作台入口已经具备 session/task 仓储、状态机、workspace event seam，以及最小 write/read side/replay query/SSE stream
- Approval baseline 已经具备 definition / instance / task 模型、start / approve / reject 闭环、最小 Admin HTTP 管理面，以及 `REQUIRE_APPROVAL` 的最小接线
- Outbox 已经具备 dispatcher、重试、终态失败、人工 recovery 基线，以及最小 Admin operator list/requeue 管理面
- 供应链交易闭环仍处于设计完成、代码未开始阶段

因此，本文对“实现情况”的判断分为两层：

- 目标覆盖：该领域或模块是否属于 Phase 1 应覆盖的范围
- 当前实现：该领域或模块在代码中是否已经落地

### 2.2 优先级定义

- `P0`：Phase 1 的阻塞性基础能力，没有它就无法继续推进控制面、执行面和业务域
- `P1`：Phase 1 的主线能力，需要在平台底座之上优先补齐
- `P2`：仍在 Phase 1 范围内，但可以排在主线之后实现

### 2.3 实现状态定义

- `已实现`：有实际代码、可编译或可运行，并已通过当前验证
- `部分实现`：已有骨架、占位或最小闭环，但距离业务可用仍有明显缺口
- `仅设计`：在设计文档中已定义，但仓库中尚无实际实现

## 3. 当前总体结论

### 3.1 已完成的 P0 平台底座

以下能力已经进入“已实现”或“可运行骨架”状态：

- 五个运行时入口：`api-server`、`agent-gateway`、`worker`、`scheduler`、`migrate`
- 本地第三方资源合同：`docker-compose.yml`、`configs/local/docker.env`、`Makefile`
- 配置加载与容器装配：`internal/bootstrap/*`
- 基础 HTTP 接口层、健康检查与中间件链
- 租户请求上下文与租户解析缝合点
- PostgreSQL / Redis / NATS / MinIO 客户端骨架
- 平台基础迁移与控制面表结构占位
- 事件总线、worker、scheduler 的运行时骨架
- 策略、审计、事务、命令管道基础实现
- 控制面实体化第一批切片：tenant / user / role / department / agent profile catalog 与用户绑定关系
- 治理核心第一批切片：policy rule 持久化、生命周期、audit query，以及最小 Admin create/list/activate/deactivate 管理面
- Agent runtime 第一批切片：session/task 仓储、状态流转、workspace event、session-scoped SSE stream
- Capability governance 第一批切片：tenant-scoped model catalog 与 tool catalog baseline，以及最小 Admin create/list 管理面
- Outbox reliability 第一批切片：dispatcher、retry、failed recovery、tenant-scoped Admin operator list/requeue，以及 poll observability seam
- Agent Gateway 的 workspace 事件入口骨架
- 本地 smoke 脚本与 live 健康验证流程

### 3.2 尚未完成的 Phase 1 主线

以下能力虽然在设计范围内，但当前仍未达到可交付状态：

- 更完整的 Tenant / IAM / Policy 控制面模型
- 插件注册、租户启用、model/tool policy binding、quota / feature flag 与租户级能力治理的剩余部分
- Agent session / task 的真实 WebSocket 协议、跨进程流式回放、执行记录、证据模型与 live integration
- Approval / Workflow 的更完整业务流程、多级编排与 runtime orchestration
- 供应链闭环所需的主数据、销售、采购、库存、应收应付上下文

### 3.3 当前阶段判断

可以将当前项目状态理解为：

- 平台底座：已完成第一批可运行基线
- Phase 1 控制面：已完成第一批可执行切片，尚未完成控制面主线
- 供应链交易闭环：尚未进入实际开发

## 4. Phase 1 领域覆盖矩阵

这一节按设计文档中的有界上下文整理“领域”层面的覆盖情况。

| 有界上下文 | 功能范围 | 优先级 | 当前实现情况 | 说明 |
| --- | --- | --- | --- | --- |
| Tenant and IAM | 租户元数据、组织结构、用户、角色、部门范围、策略引用 | P1 | 部分实现 | 已有 `internal/platform/tenant/*` 与 `internal/platform/iam/actor.go` 的请求上下文基线，同时已经补上 `internal/domain/controlplane/*`、`internal/application/controlplane/*`、`internal/interfaces/http/router/admin.go` 和 `migrations/000007_phase1_tenant_iam_extension.*`，落地了 role/department catalog、user-role/user-department 绑定、最小 Admin create/list/bind 闭环，以及 tenant root existence enforcement；但 actor 注入、ABAC/RBAC、组织树治理与更深的租户控制面仍未完成。 |
| Master Data | 客户、供应商、商品、仓库、库位、税码、币种、计量单位、价目表 | P1 | 仅设计 | 设计已定义，但仓库中没有 `internal/domain/masterdata` 或对应应用层实现。 |
| Sales | 报价、销售订单、发运计划、订单生命周期 | P1 | 仅设计 | 设计已定义，当前没有销售领域模型、命令处理器或 API。 |
| Procurement | 请购、采购订单、供应商事务状态、收货计划 | P1 | 仅设计 | 设计已定义，当前没有采购上下文代码。 |
| Inventory | 入库、出库、预留、调拨、库存台账、可用/预留库存状态 | P1 | 仅设计 | 设计已定义，当前没有库存聚合、库存流水或库存查询模型。 |
| Receivable and Payable | 应收单、应付单、开票申请、付款计划 | P2 | 仅设计 | 设计已定义，当前没有应收应付上下文实现。 |
| Approval and Workflow | 审批定义、审批实例、人工任务、流程推进 | P1 | 部分实现 | 已补 `internal/domain/approval/*`、`internal/application/approval/*`、`internal/infrastructure/persistence/postgres/approval_repository.go`、`internal/bootstrap/approval_catalog.go` 与 `migrations/000009_phase1_approval_baseline.*`，落地 definition / instance / task 模型、start / approve / reject 闭环、租户级 list query，以及最小 Admin HTTP create/list/approve/reject 管理面，并通过 `internal/application/shared/pipeline.go` 的 approval starter seam 把 `REQUIRE_APPROVAL` 接到了审批实例创建；但仍没有 workflow engine 或更复杂的多级审批编排。 |
| Agent Task and Automation | Agent profile、session、task、execution plan、tool record、policy decision record | P1 | 部分实现 | 已有 agent profile 目录、`agent_session` / `agent_task` 迁移、runtime service、状态流转、workspace event 广播 seam，以及围绕 `session_key` 统一的 workspace write/read-side/query 契约，但还没有流式协议、执行记录、tool record 与完整证据模型。 |

## 5. Phase 1 模块覆盖矩阵

这一节按模块和代码结构整理当前实现情况。

### 5.1 Experience Plane / 接口入口层

| 领域 | 模块 | 功能清单 | 优先级 | 当前实现情况 | 代码锚点 |
| --- | --- | --- | --- | --- | --- |
| Experience Plane | API Server 运行时 | 加载配置、装配容器、启动 Gin HTTP 服务 | P0 | 已实现 | `cmd/api-server/main.go` |
| Experience Plane | HTTP 路由分组 | `Admin API`、`Workspace API`、`Platform API`、`Integration API` 的分组入口 | P0 | 部分实现 | `internal/interfaces/http/router/router.go`、`admin.go`、`workspace.go`、`platform.go`、`integration.go`；当前 `Platform API` 已提供健康检查，`Admin API` 已提供控制面目录、policy rule / audit、审批定义/实例/任务、model/tool catalog，以及 outbox message operator 的最小写入/查询接口，`Workspace API` 已提供 sessions/tasks/events/stream 的最小写入/查询/SSE 接口，`Integration API` 仍是占位 |
| Experience Plane | 健康检查接口 | `/api/platform/v1/health/livez`、`/readyz` | P0 | 已实现 | `internal/interfaces/http/router/health.go`、`internal/platform/health/service.go` |
| Experience Plane | 中间件链 | request ID、logging、tenant、auth、audit | P0 | 已实现 | `internal/interfaces/http/middleware/*` |
| Experience Plane | Workspace Gateway seam | 工作台会话注册、事件广播骨架 | P1 | 部分实现 | `internal/interfaces/ws/workspace_gateway.go`；已经具备 session-aware subscription、广播和最小事件 replay/query/SSE stream seam，并已通过 `internal/interfaces/http/router/workspace.go` 暴露最小写入/查询/stream HTTP surface；但仍未实现真实 WebSocket 协议和跨进程回放。 |
| Experience Plane | Agent Workspace / Web Console 业务接口 | 工作台命令、会话查询、事件订阅、控制台业务操作 | P1 | 部分实现 | 已补 `internal/interfaces/http/router/workspace.go`，提供 sessions/tasks/events/stream 的最小写入/查询/SSE 接口；但仍没有真实 WebSocket 协议或前端应用。 |

### 5.2 Control Plane / 平台控制面

| 领域 | 模块 | 功能清单 | 优先级 | 当前实现情况 | 代码锚点 |
| --- | --- | --- | --- | --- | --- |
| Control Plane | Tenant 解析与 CellRoute | 租户标识、隔离信息、数据库/缓存/存储前缀路由 | P0 | 部分实现 | `internal/platform/tenant/resolver.go`、`cell_route.go`；当前只支持基于 `X-Tenant-ID` 的简单解析 |
| Control Plane | IAM Actor 基线 | 请求上下文中的 actor 注入 | P0 | 部分实现 | `internal/platform/iam/actor.go`、`internal/interfaces/http/middleware/auth.go`；当前只有占位 `system` actor |
| Control Plane | Policy Engine | 策略决策枚举、评估接口、规则生命周期、命令治理缝合点 | P1 | 部分实现 | `internal/platform/policy/*`、`internal/application/governance/*`、`internal/infrastructure/persistence/postgres/policy_audit_repository.go`、`internal/bootstrap/governance_catalog.go`；当前已具备 repository-backed rule evaluator、activate/deactivate 生命周期、治理命令处理器，以及最小 Admin create/list/activate/deactivate 管理面，但还没有真实 ABAC / RBAC / rules engine |
| Control Plane | Audit 基线 | 审计记录模型、Recorder / Store、持久化查询 | P1 | 部分实现 | `internal/platform/audit/*`、`internal/infrastructure/persistence/postgres/policy_audit_repository.go`、`internal/bootstrap/governance_catalog.go`；已具备持久化 store、查询服务和最小 Admin audit list 接口，但还没有更完整的审计治理、分页和 retention 策略 |
| Control Plane | Agent session/task 元数据 | session/task 表、仓储、状态机、workspace event identity | P1 | 部分实现 | `internal/domain/agentruntime/*`、`internal/application/agentruntime/*`、`internal/infrastructure/persistence/postgres/agent_runtime_repository.go`、`migrations/000004_phase1_agent_runtime_control.*`；已具备仓储、状态流转、close/fail/cancel 流程，以及以 `session_key` 为查询契约的 create/list/update/replay/stream workspace 最小 HTTP surface，但还没有真实 WebSocket 协议、跨进程流式回放与执行记录模型。 |
| Control Plane | Plugin / Tool Registry | 插件注册、工具目录、租户启用、风险级别、输入输出 schema | P1 | 部分实现 | 已补 `internal/domain/capability/tool_catalog_entry.go`、`internal/application/capability/*tool*`、`internal/application/capability/*agent_capability_policy*`、`internal/infrastructure/persistence/postgres/capability_repository.go`、`internal/bootstrap/capability_catalog.go`、`internal/interfaces/http/router/admin.go` 与 `migrations/000008_phase1_tool_catalog.*`、`migrations/000010_phase1_agent_capability_policy.*`，落地 tenant-scoped tool catalog baseline 以及 agent profile -> allowed tools 最小 Admin 管理/查询面；但 plugin registry、tenant enablement、tool schema/runtime 仍未完成。 |
| Control Plane | Quota / Feature Flag / Model Catalog | 配额、租户特性开关、模型与工具可用性治理 | P2 | 部分实现 | `internal/domain/capability/*`、`internal/application/capability/*`、`internal/infrastructure/persistence/postgres/capability_repository.go`、`internal/bootstrap/capability_catalog.go`、`internal/interfaces/http/router/admin.go`、`migrations/000005_phase1_capability_governance.*`、`migrations/000008_phase1_tool_catalog.*`、`migrations/000010_phase1_agent_capability_policy.*`；当前已经落地 tenant-scoped model catalog、tool catalog 与最小 agent capability policy binding（按 agent profile 管理 allowed models/tools）Admin 面，quota、feature flag、tenant enablement 与 runtime-side capability enforcement 仍未开始。 |

### 5.3 Execution Plane / Agent 与自动化执行面

| 领域 | 模块 | 功能清单 | 优先级 | 当前实现情况 | 代码锚点 |
| --- | --- | --- | --- | --- | --- |
| Execution Plane | Command Pipeline | 命令进入应用处理器、策略前置、事务边界、审计记录 | P0 | 部分实现 | `internal/application/shared/pipeline.go`、`command.go`、`transaction.go`；当前已经具备 policy decision、transaction boundary、audit record 与 `REQUIRE_APPROVAL` starter seam，但仍未扩展为更完整的 workflow orchestration pipeline |
| Execution Plane | Event Bus | 内存总线、NATS JetStream 总线 | P0 | 已实现 | `internal/platform/eventbus/*` |
| Execution Plane | Worker Runtime | 装配依赖、轮询 outbox 的后台进程骨架 | P0 | 部分实现 | `cmd/worker/main.go`、`internal/application/shared/outbox/*`；当前已接入 dispatcher、重试、失败恢复与最小 Admin operator seam，但仍缺少更完整的运行治理和操作面 |
| Execution Plane | Scheduler Runtime | 定时 tick 产生时间驱动事件 | P0 | 部分实现 | `cmd/scheduler/main.go`；当前只发送 `platform.scheduler.tick` 骨架事件 |
| Execution Plane | Agent Gateway Runtime | 工作台事件入口、会话 channel 注册、进程生命周期 | P1 | 部分实现 | `cmd/agent-gateway/main.go`、`internal/interfaces/ws/workspace_gateway.go` |
| Execution Plane | Session Context Assembler | tenant/actor/business/policy/knowledge/execution context 组装 | P1 | 部分实现 | `internal/platform/runtime/request_context.go` 只覆盖 request 级元数据，没有完整 session context assembler |
| Execution Plane | Workflow Orchestration | 长事务编排、审批暂停/恢复、人工接管、回滚与补偿 | P1 | 部分实现 | 当前已经具备 approval baseline 和 `REQUIRE_APPROVAL` -> approval instance 的最小接线，但还没有真正的 workflow engine、暂停/恢复状态机或 orchestration service。 |
| Execution Plane | Tool Runtime / Query Tool / Command Tool | 工具目录、输入输出 schema、权限、审批、风险控制 | P1 | 部分实现 | 当前已具备 tenant-scoped tool catalog baseline 和最小 Admin 管理面，但还没有 tool executor、tool policy binding、schema contract 和 runtime approval integration |

### 5.4 Data and Integration Plane / 数据与集成层

| 领域 | 模块 | 功能清单 | 优先级 | 当前实现情况 | 代码锚点 |
| --- | --- | --- | --- | --- | --- |
| Data Plane | PostgreSQL 基线 | 连接配置、SQL DB 构建、事务辅助、迁移执行 | P0 | 已实现 | `internal/infrastructure/persistence/postgres/*`、`cmd/migrate/main.go` |
| Data Plane | 控制面基础迁移 | `tenant`、`tenant_cell`、`audit_log`、`agent_session`、`agent_task`、`outbox` 表 | P0 | 已实现 | `migrations/000001_init_platform_tables.up.sql` |
| Data Plane | Redis seam | 客户端配置校验与构建 | P0 | 已实现 | `internal/infrastructure/cache/redis/client.go` |
| Data Plane | NATS seam | 连接配置校验与构建 | P0 | 已实现 | `internal/infrastructure/messaging/nats/client.go` |
| Data Plane | MinIO seam | 对象存储客户端配置校验与构建 | P0 | 已实现 | `internal/infrastructure/storage/minio/client.go` |
| Data Plane | Outbox Pattern 基础 | 表结构、worker 轮询、总线发布、失败重试与恢复 | P1 | 部分实现 | `internal/application/shared/outbox/*`、`internal/infrastructure/persistence/postgres/outbox_repository.go`、`migrations/000006_phase1_reliability_hardening.*`；已具备 dispatcher、retry、failed terminal state、stale `publishing` reclaim、tenant-scoped Admin operator list/requeue、cross-tenant requeue rejection 与 poll observability seam，但仍缺少 consumer-side idempotency、DLQ 与 live DB contention 验证 |
| Data Plane | CQRS Read Models | Backoffice、Workspace、Monitoring、Analytics 读模型 | P1 | 仅设计 | 当前没有 projection、read service 或读模型表 |
| Data Plane | Search / Vector Retrieval | PostgreSQL FTS、trigram、`pgvector` | P2 | 仅设计 | 设计已定义，当前迁移中没有相关扩展与索引结构 |
| Integration Plane | 外部系统连接器 | webhook、ERP 外部系统、插件式连接器 | P2 | 仅设计 | 只有 `Integration API` 路由分组占位，没有 connector/runtime 实现 |

### 5.5 Engineering / Runtime Foundation / 工程与运行时底座

| 领域 | 模块 | 功能清单 | 优先级 | 当前实现情况 | 代码锚点 |
| --- | --- | --- | --- | --- | --- |
| Foundation | 运行时入口 | `api-server`、`agent-gateway`、`worker`、`scheduler`、`migrate` | P0 | 已实现 | `cmd/*/main.go`、`internal/bootstrap/runtime.go` |
| Foundation | 配置与装配 | YAML 配置、环境变量覆盖、容器装配 | P0 | 已实现 | `internal/bootstrap/config.go`、`container.go`、`configs/local/app.yaml` |
| Foundation | Compose 开发合同 | PostgreSQL、Redis、NATS、MinIO、OTEL、Prometheus、Grafana | P0 | 已实现 | `docker-compose.yml`、`configs/local/docker.env` |
| Foundation | 本地开发命令 | `infra-up`、`infra-down`、`test`、`smoke`、`migrate-up`、`migrate-down` | P0 | 已实现 | `Makefile` |
| Foundation | Live Smoke Workflow | 本地健康检查脚本与集成测试 | P0 | 已实现 | `scripts/smoke_local.sh`、`test/integration/api_health_test.go` |
| Foundation | Observability 基础合同 | OTEL endpoint、Prometheus、Grafana | P1 | 部分实现 | Compose 与 config 已准备，且已有 `internal/infrastructure/observability/outbox_poll.go` 的 worker poll instrumentation seam，但还没有统一 OTEL bootstrap 与端到端 tracing 初始化 |

## 6. 按优先级整理的功能清单

### 6.1 P0：已完成或已打通的基础能力

- 五个运行时角色与启动入口
- 本地 Compose 资源标准化
- 配置装配与健康检查
- Gin 路由分组与基础中间件链
- 请求级租户与 actor 上下文
- PostgreSQL / Redis / NATS / MinIO 客户端骨架
- 平台基础迁移与 `migrate` 运行时
- Event Bus、worker、scheduler 运行时骨架
- Policy / Audit / Transaction / Command Pipeline 基础实现
- Agent Gateway workspace seam
- `make smoke` 与 live health 验证

### 6.2 P1：下一批必须补齐的主线能力

- 更完整的 Tenant / Organization / User / Role / Department 模型与 actor/policy 联动
- 更完整的控制面策略引擎，而不仅是 repository-backed rule evaluator
- 更完整的审计治理、分页、retention 与合规能力
- Agent session/task 的真实 WebSocket 协议、跨进程流式回放、执行记录、证据模型与 live integration
- Workspace HTTP surface、真实 WebSocket 协议、跨进程事件回放与更完整的事件订阅
- Approval 的多级审批与 workflow orchestration
- Outbox 的 consumer-side idempotency、DLQ 与 live DB contention 验证
- 插件注册、租户启用、model/tool policy binding、quota / feature flag 与能力治理剩余部分

### 6.3 P2：Phase 1 边界内但可顺延的能力

- 配额与 feature flag
- 读模型与监控分析投影
- `pgvector`、全文检索、知识检索
- 更完整的 observability 初始化
- 外部系统 connector 与 integration runtime

## 7. 供应链交易闭环的实现情况

这是当前文档最重要的结论之一。

“供应链交易闭环”虽然是项目确认过的第一业务闭环，但在当前仓库中的实现状态仍然是：

- 设计已完成
- 执行架构已经为它预留了运行时与数据缝合点
- 实际业务领域代码仍未开始

具体来看：

- 主数据：未开始
- 销售：未开始
- 采购：未开始
- 库存：未开始
- 审批：平台级审批基线与 Admin 管理面已实现，但与供应链业务单据耦合的审批编排仍未开始
- 应收应付基础：未开始

也就是说，当前代码库还不能被视为“已进入供应链 ERP 功能开发阶段”，而应被视为“已经完成平台底座并进入 Phase 1 控制面实体化阶段”。

## 8. 对 Phase 1 完成度的定性判断

如果按“平台底座是否可作为后续 Phase 1 开发起点”来判断，当前状态是：

- 可以开始下一批 Phase 1 主线开发

如果按“Phase 1 是否已经完成”来判断，当前状态是：

- 明确未完成

原因是：

- 平台底座已具备
- 控制面主线虽然已有第一批可执行切片，但距离完整交付仍有明显缺口
- 供应链业务闭环尚未落地

## 9. 推荐的下一步实现顺序

建议后续实施按下面顺序推进。

### Wave 1：真实控制面落地

- Tenant / IAM 数据模型与仓储
- Policy engine 与授权模型
- Audit persistence 与审计查询
- Agent session / task 仓储与状态流转

### Wave 2：供应链闭环最小域模型

- Master Data
- Procurement
- Inventory
- Approval
- Sales 最小闭环

### Wave 3：Agent Workspace 强化

- Workspace WebSocket / SSE 强化与跨进程回放
- 执行状态流
- 人工接管与审批暂停/恢复
- Tool catalog / command translator / knowledge retrieval

## 10. 结论

当前项目已经完成了“Phase 1 可开工”的平台基础，并补齐了第一批控制面可执行切片，但还没有完成“Phase 1 业务目标”本身。

最准确的表述应该是：

- 已完成：Phase 0/1 平台底座
- 部分完成：Phase 1 控制面第一批可执行切片
- 尚未开始：Phase 1 供应链交易闭环业务实现

因此，后续所有业务域开发都应该以本文中的 `P1` 模块为第一优先级，而不是继续扩展更多技术骨架。
