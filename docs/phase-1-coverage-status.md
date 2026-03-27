# Phase 1 覆盖领域、模块、功能清单、优先级与实现情况

更新时间：2026-03-27

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
- 控制面核心能力已经有基础缝合点
- Agent 执行与工作台入口已经有运行时占位
- 供应链交易闭环已在 Phase 2 形成最小可运行回路（主数据/采购/库存/应收应付/销售与相关 admin API），但不等于 Phase 1 主线目标已完成

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
- Agent Gateway 的 workspace 事件入口骨架
- 本地 smoke 脚本与 live 健康验证流程

### 3.2 尚未完成的 Phase 1 主线

以下能力虽然在设计范围内，但当前仍未达到 Phase 1 目标口径下的可交付状态：

- 真正的 Tenant / IAM / Policy 控制面模型
- 插件、工具、模型目录与租户级能力治理
- Agent session / task 的存储、状态机与流式协议
- Approval / Workflow 的真实业务流程
- 面向 Agent 原生执行面的审批/流程编排、会话状态机、能力治理等主线

### 3.3 当前阶段判断

可以将当前项目状态理解为：

- 平台底座：已完成第一批可运行基线
- Phase 1 控制面：完成约束骨架，尚未完成业务化实现
- 供应链交易闭环：已由 Phase 2 波次落地最小运行闭环，但与 Phase 1 控制面主线目标并行

## 4. Phase 1 领域覆盖矩阵

这一节按设计文档中的有界上下文整理“领域”层面的覆盖情况。

| 有界上下文 | 功能范围 | 优先级 | 当前实现情况 | 说明 |
| --- | --- | --- | --- | --- |
| Tenant and IAM | 租户元数据、组织结构、用户、角色、部门范围、策略引用 | P1 | 部分实现 | 已有 `internal/platform/tenant/*` 与 `internal/platform/iam/actor.go`，但目前只有基于请求头的租户解析和占位 `system` actor，没有真实租户目录、组织结构、用户和角色模型。 |
| Master Data | 客户、供应商、商品、仓库、库位、税码、币种、计量单位、价目表 | P1 | 部分实现 | 已有 `internal/domain/masterdata/*` 与 admin 接口落地（来自 Phase 2 业务波次），但仍缺少更完整主数据治理能力。 |
| Sales | 报价、销售订单、发运计划、订单生命周期 | P1 | 部分实现 | 已有最小销售订单与发运路径（来自 Phase 2），尚未覆盖完整生命周期与策略约束。 |
| Procurement | 请购、采购订单、供应商事务状态、收货计划 | P1 | 部分实现 | 已有采购订单创建/提交流程与收货路径（来自 Phase 2），仍缺少更完整流程编排。 |
| Inventory | 入库、出库、预留、调拨、库存台账、可用/预留库存状态 | P1 | 部分实现 | 已有库存流水/余额/预留/出库/调拨与 transfer order 最小能力（来自 Phase 2），仍缺少更深层库存治理。 |
| Receivable and Payable | 应收单、应付单、开票申请、付款计划 | P2 | 部分实现 | 已有最小应收应付与支付计划路径（来自 Phase 2），尚未覆盖完整财务流程。 |
| Approval and Workflow | 审批定义、审批实例、人工任务、流程推进 | P1 | 部分实现 | 已有审批实例通过/拒绝最小闭环（来自 Phase 2），但通用 workflow 编排与人工任务引擎未落地。 |
| Agent Task and Automation | Agent profile、session、task、execution plan、tool record、policy decision record | P1 | 部分实现 | 已有 `agent_session`、`agent_task` 表占位，`internal/interfaces/ws/workspace_gateway.go` 与 `internal/platform/runtime/workspace_event.go` 提供运行时骨架，但没有真实会话仓储、任务状态机、流式协议和工具执行记录。 |

## 5. Phase 1 模块覆盖矩阵

这一节按模块和代码结构整理当前实现情况。

### 5.1 Experience Plane / 接口入口层

| 领域 | 模块 | 功能清单 | 优先级 | 当前实现情况 | 代码锚点 |
| --- | --- | --- | --- | --- | --- |
| Experience Plane | API Server 运行时 | 加载配置、装配容器、启动 Gin HTTP 服务 | P0 | 已实现 | `cmd/api-server/main.go` |
| Experience Plane | HTTP 路由分组 | `Admin API`、`Workspace API`、`Platform API`、`Integration API` 的分组入口 | P0 | 部分实现 | `internal/interfaces/http/router/router.go`、`admin.go`、`workspace.go`、`platform.go`、`integration.go`；当前只有 `Platform API` 挂了健康检查 |
| Experience Plane | 健康检查接口 | `/api/platform/v1/health/livez`、`/readyz` | P0 | 已实现 | `internal/interfaces/http/router/health.go`、`internal/platform/health/service.go` |
| Experience Plane | 中间件链 | request ID、logging、tenant、auth、audit | P0 | 已实现 | `internal/interfaces/http/middleware/*` |
| Experience Plane | Workspace Gateway seam | 工作台会话注册、事件广播骨架 | P1 | 部分实现 | `internal/interfaces/ws/workspace_gateway.go`；未实现真实 WebSocket 协议与事件回放 |
| Experience Plane | Agent Workspace / Web Console 业务接口 | 工作台命令、会话查询、事件订阅、控制台业务操作 | P1 | 仅设计 | 目前只有路由占位与 gateway seam，没有业务 API 或前端应用 |

### 5.2 Control Plane / 平台控制面

| 领域 | 模块 | 功能清单 | 优先级 | 当前实现情况 | 代码锚点 |
| --- | --- | --- | --- | --- | --- |
| Control Plane | Tenant 解析与 CellRoute | 租户标识、隔离信息、数据库/缓存/存储前缀路由 | P0 | 部分实现 | `internal/platform/tenant/resolver.go`、`cell_route.go`；当前只支持基于 `X-Tenant-ID` 的简单解析 |
| Control Plane | IAM Actor 基线 | 请求上下文中的 actor 注入 | P0 | 部分实现 | `internal/platform/iam/actor.go`、`internal/interfaces/http/middleware/auth.go`；当前只有占位 `system` actor |
| Control Plane | Policy Engine | 策略决策枚举、评估接口、静态评估器 | P1 | 部分实现 | `internal/platform/policy/*`；当前是静态 evaluator，没有真实 ABAC / RBAC / rules engine |
| Control Plane | Audit 基线 | 审计记录模型、Recorder 接口、noop / in-memory recorder | P1 | 部分实现 | `internal/platform/audit/*`；没有持久化审计仓储和审计查询 |
| Control Plane | Agent session/task 元数据 | session/task 表、workspace event identity | P1 | 部分实现 | `migrations/000001_init_platform_tables.up.sql`、`internal/platform/runtime/workspace_event.go` |
| Control Plane | Plugin / Tool Registry | 插件注册、工具目录、租户启用、风险级别、输入输出 schema | P1 | 仅设计 | 设计已明确，代码中尚无 registry、catalog 或 capability registration |
| Control Plane | Quota / Feature Flag / Model Catalog | 配额、租户特性开关、模型与工具可用性治理 | P2 | 仅设计 | 设计已覆盖，但当前没有控制面数据结构与服务 |

### 5.3 Execution Plane / Agent 与自动化执行面

| 领域 | 模块 | 功能清单 | 优先级 | 当前实现情况 | 代码锚点 |
| --- | --- | --- | --- | --- | --- |
| Execution Plane | Command Pipeline | 命令进入应用处理器、策略前置、事务边界、审计记录 | P0 | 部分实现 | `internal/application/shared/pipeline.go`、`command.go`、`transaction.go`；当前只完成基础执行链路 |
| Execution Plane | Event Bus | 内存总线、NATS JetStream 总线 | P0 | 已实现 | `internal/platform/eventbus/*` |
| Execution Plane | Worker Runtime | 装配依赖、轮询 outbox 的后台进程骨架 | P0 | 部分实现 | `cmd/worker/main.go` 已实现 claim/publish/标记 published、失败回退 pending+delay、attempts 计数与阈值后 failed 终态，以及 `processing` 租约写入 `available_at` 的超时回收；DLQ 与消费幂等等治理仍待补齐。 |
| Execution Plane | Scheduler Runtime | 定时 tick 产生时间驱动事件 | P0 | 部分实现 | `cmd/scheduler/main.go`；当前只发送 `platform.scheduler.tick` 骨架事件 |
| Execution Plane | Agent Gateway Runtime | 工作台事件入口、会话 channel 注册、进程生命周期 | P1 | 部分实现 | `cmd/agent-gateway/main.go`、`internal/interfaces/ws/workspace_gateway.go` |
| Execution Plane | Session Context Assembler | tenant/actor/business/policy/knowledge/execution context 组装 | P1 | 部分实现 | `internal/platform/runtime/request_context.go` 只覆盖 request 级元数据，没有完整 session context assembler |
| Execution Plane | Workflow Orchestration | 长事务编排、审批暂停/恢复、人工接管、回滚与补偿 | P1 | 仅设计 | 当前没有 workflow engine、状态机或 orchestration service |
| Execution Plane | Tool Runtime / Query Tool / Command Tool | 工具目录、输入输出 schema、权限、审批、风险控制 | P1 | 仅设计 | 当前没有 tool catalog、tool executor、tool policy binding |

### 5.4 Data and Integration Plane / 数据与集成层

| 领域 | 模块 | 功能清单 | 优先级 | 当前实现情况 | 代码锚点 |
| --- | --- | --- | --- | --- | --- |
| Data Plane | PostgreSQL 基线 | 连接配置、SQL DB 构建、事务辅助、迁移执行 | P0 | 已实现 | `internal/infrastructure/persistence/postgres/*`、`cmd/migrate/main.go` |
| Data Plane | 控制面基础迁移 | `tenant`、`tenant_cell`、`audit_log`、`agent_session`、`agent_task`、`outbox` 表 | P0 | 已实现 | `migrations/000001_init_platform_tables.up.sql` |
| Data Plane | Redis seam | 客户端配置校验与构建 | P0 | 已实现 | `internal/infrastructure/cache/redis/client.go` |
| Data Plane | NATS seam | 连接配置校验与构建 | P0 | 已实现 | `internal/infrastructure/messaging/nats/client.go` |
| Data Plane | MinIO seam | 对象存储客户端配置校验与构建 | P0 | 已实现 | `internal/infrastructure/storage/minio/client.go` |
| Data Plane | Outbox Pattern 基础 | 表结构、worker 轮询骨架、总线发布路径 | P1 | 部分实现 | `outbox` 表与 worker 发布路径已落地（pending -> processing -> published，失败回退 pending+available_at，attempts 计数与 failed 终态，并支持 `processing` 记录租约到期后回收再处理）；仍缺少 DLQ/inbox 幂等等完整治理。 |
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
| Foundation | Observability 基础合同 | OTEL endpoint、Prometheus、Grafana | P1 | 部分实现 | Compose 与 config 已准备，现已新增 `internal/infrastructure/observability/otel/setup.go` 与 runtime 入口初始化 seam；真实 exporter/provider 仍待后续补齐。 |

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

- 真实的 Tenant / Organization / User / Role / Department 模型
- 真正的控制面策略引擎，而不是静态 evaluator
- 审计记录持久化、查询与审计治理
- Agent session/task 仓储、状态机、执行记录、证据模型
- Workspace 真正的流式协议与事件订阅
- Outbox 完整治理能力（DLQ、消费幂等、跨边界去重、可观测告警）
- 插件注册、工具目录、模型目录、租户能力治理
- Approval 基线（已部分落地），Workflow Orchestrator 仍未完成

### 6.3 P2：Phase 1 边界内但可顺延的能力

- 配额与 feature flag
- 读模型与监控分析投影
- `pgvector`、全文检索、知识检索
- 更完整的 observability 初始化
- 外部系统 connector 与 integration runtime

## 7. 供应链交易闭环的实现情况

“供应链交易闭环”已经不再是“仅设计”状态。

当前仓库已有可运行的最小业务路径（主要由 Phase 2 波次落地）：

- 主数据：供应商/商品/仓库最小能力
- 采购：采购订单创建/提交/收货
- 库存：流水、余额、预留、出库、调拨、transfer order
- 审批：审批实例通过/拒绝最小闭环
- 应收应付：最小查询与计划能力
- 销售：最小销售订单与发运路径

但这并不代表 Phase 1 已完成，因为 Phase 1 的主线目标仍聚焦在控制面、执行面和治理能力（session/task state machine、workflow orchestration、tool/model governance）。

## 8. 对 Phase 1 完成度的定性判断

如果按“平台底座是否可作为后续 Phase 1 开发起点”来判断，当前状态是：

- 可以开始下一批 Phase 1 主线开发

如果按“Phase 1 是否已经完成”来判断，当前状态是：

- 明确未完成

原因是：

- 平台底座已具备
- 控制面主线只完成了基础骨架
- 控制面与执行面主线只完成了基础骨架，尚未完成目标化交付
- 供应链最小闭环已落地，但属于跨阶段并行推进成果

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

- Workspace WebSocket / SSE 协议
- 执行状态流
- 人工接管与审批暂停/恢复
- Tool catalog / command translator / knowledge retrieval

## 10. 结论

当前项目已经完成了“Phase 1 可开工”的平台基础，但还没有完成“Phase 1 业务目标”本身。

最准确的表述应该是：

- 已完成：Phase 0/1 平台底座
- 部分完成：Phase 1 控制面骨架 + observability 初始化 seam +审批最小基线
- 尚未完成：Phase 1 控制面/执行面的目标能力闭环（workflow orchestration、task runtime、capability governance 等）

因此，后续所有业务域开发都应该以本文中的 `P1` 模块为第一优先级，而不是继续扩展更多技术骨架。
