# Phase 3 覆盖领域、模块、功能清单、优先级与实现情况

更新时间：2026-03-25

## 1. Phase 3 定位

Phase 3 的名称是：

- `AI-Native Execution Strengthening`
- 中文可表述为：`AI-Native 执行面强化`

这个阶段不是简单“接入更多模型”，而是把平台里的 Agent、Workspace、Workflow、Knowledge、Policy、Human Handoff 做成真正能承载企业执行风险的系统。

## 2. Phase 3 目标

### 2.1 核心目标

- 让 Agent Workspace 成为一等入口
- 让 agent session / task / execution state 成为可观测、可审计、可暂停恢复的正式运行时
- 让工具调用、审批暂停、人工接管进入统一治理框架

### 2.2 目标边界

Phase 3 的重点不是增加更多 ERP 业务域，而是强化执行层。

因此它依赖：

- Phase 2 至少完成一条真实交易闭环
- 控制面里已有真实 policy、audit、session/task 模型

## 3. Phase 3 领域覆盖矩阵

| 有界上下文 / 能力域 | 功能范围 | 优先级 | 当前实现情况 | 说明 |
| --- | --- | --- | --- | --- |
| Agent Task and Automation | Agent profile、session、task、execution plan、tool execution record、policy decision record | P0 | 部分实现 | 已有 `agent_session` / `agent_task` 表占位和 workspace event skeleton，但没有真实状态机和仓储。 |
| Approval and Workflow | 审批暂停/恢复、人工任务、长事务编排、人工接管 | P0 | 仅设计 | 这是企业级 Agent 执行不可缺少的一层。 |
| Policy and Governance | `ALLOW`、`ALLOW_WITH_GUARD`、`REQUIRE_APPROVAL`、`DENY` 的真实动态执行 | P0 | 部分实现 | 决策枚举和 evaluator 接口已存在，但没有真实规则运行时。 |
| Knowledge Retrieval | 检索上下文、权限过滤、知识来源、证据归档 | P1 | 仅设计 | 这是 Agent 高质量执行的关键，但应建立在业务对象与权限模型之上。 |
| Workspace Experience | 实时事件流、任务时间线、状态播报、执行证据可视化 | P1 | 部分实现 | 只有 workspace gateway seam 和 typed event envelope。 |
| Tool Runtime | query tool、command tool、action tool、knowledge tool 目录与执行 | P1 | 仅设计 | 当前还没有工具目录、schema、权限和风险模型。 |
| Human Handoff / Escalation | 升级到人工、审批等待、失败证据、部分成功处理 | P1 | 仅设计 | 设计中明确要求，但代码未开始。 |

## 4. Phase 3 模块覆盖矩阵

### 4.1 Execution Plane

| 模块 | 功能清单 | 优先级 | 当前实现情况 |
| --- | --- | --- | --- |
| Agent Gateway | 工作台入口、事件流入口、task/session 装配、状态广播 | P0 | 部分实现，只有 runtime seam |
| Session Context Assembler | tenant / actor / business / policy / knowledge / execution context 装配 | P0 | 部分实现，当前只有 request context |
| Execution State Machine | `draft`、`planned`、`policy_checking`、`awaiting_approval`、`executing`、`succeeded` 等状态流转 | P0 | 未开始 |
| Workflow Orchestrator | 长事务、补偿、暂停/恢复、人工接管 | P0 | 未开始 |
| Tool Runtime | 工具目录、调用执行、schema 校验、权限和风险控制 | P1 | 未开始 |
| Knowledge Retrieval Layer | 检索、过滤、证据、上下文注入 | P1 | 未开始 |

### 4.2 Control Plane

| 模块 | 功能清单 | 优先级 | 当前实现情况 |
| --- | --- | --- | --- |
| Agent Governance | agent identity、risk class、allowed models、allowed tools、execution bounds | P0 | 未开始 |
| Policy Engine | 动态策略、审批门槛、上下文化决策 | P0 | 仅有静态 evaluator |
| Audit Governance | 区分人类发起与 agent 发起动作、保留 policy result 和 evidence | P0 | 仅有基础 recorder 模型 |
| Tool / Model Catalog | tool registry、model catalog、tenant enablement | P1 | 未开始 |

### 4.3 Interface Plane

| 模块 | 功能清单 | 优先级 | 当前实现情况 |
| --- | --- | --- | --- |
| Workspace Stream Protocol | `/ws/workspace`、`/ws/agent-events`、任务状态流、心跳和重连约束 | P0 | 未开始，只有 gateway skeleton |
| Workspace Query / Command API | 会话查询、任务查询、人工动作入口 | P1 | 未开始 |
| Monitoring / Timeline View | 执行状态时间线、policy decision、tool evidence | P1 | 未开始 |

### 4.4 Data Plane

| 模块 | 功能清单 | 优先级 | 当前实现情况 |
| --- | --- | --- | --- |
| Agent Session / Task Storage | session、task、tool execution、policy decision、evidence | P0 | 仅 session/task 基础表占位 |
| Workspace Read Models | 工作台视图、监控视图、任务概览 | P1 | 未开始 |
| Retrieval Store | 检索块、embedding、权限过滤索引 | P2 | 未开始 |

## 5. Phase 3 优先级清单

### 5.1 P0：必须先做

- Agent session / task 的真实仓储与状态机
- Session Context Assembler
- Policy / approval / audit 的真实执行链
- Workflow Orchestrator
- Workspace 实时事件协议
- Human handoff / escalation

### 5.2 P1：在主线成形后补齐

- Tool catalog / tool runtime
- Workspace 读模型与任务时间线
- 知识检索与证据注入
- 监控与执行分析

### 5.3 P2：可作为后续增强

- 更复杂的多代理协作
- 更高级的知识索引
- 更复杂的工具编排 DSL

## 6. 当前准备度判断

Phase 3 的若干前置骨架已经存在：

- `internal/interfaces/ws/workspace_gateway.go` 提供了单 runtime seam
- `internal/platform/runtime/workspace_event.go` 提供了 typed workspace event
- `internal/platform/policy/*`、`internal/platform/audit/*`、`internal/application/shared/pipeline.go` 提供了最小执行链骨架
- `agent_session`、`agent_task` 表已作为后续运行时数据的第一批占位

但总体上仍然不能说 Phase 3 已经开始，原因是：

- 还没有真实业务命令可供 Agent 执行
- 还没有 session/task 仓储与状态流
- 还没有工作台流式协议
- 还没有工具目录与知识检索

结论：

- `Phase 3 依赖已经开始准备`
- `但正式实施必须晚于 Phase 2 主线`

## 7. 推荐的分波次落地方式

### Wave 1：执行状态和工作台最小闭环

- session/task 仓储
- 执行状态机
- workspace 流式协议
- 基础 timeline

### Wave 2：审批暂停与人工接管

- policy gating
- approval pause/resume
- human handoff
- escalation path

### Wave 3：工具运行时与知识检索

- tool catalog
- query/command/action/knowledge tools
- retrieval context
- 证据归档与可视化

## 8. Phase 3 完成判定建议

只有当下面条件成立时，Phase 3 才能认为进入完成态：

- Agent Workspace 可以承载真实业务任务
- 会话、任务、状态流、证据、策略结果都可追踪
- 高风险动作会进入审批或人工接管
- 工具目录和权限边界清晰可控
- 失败、部分成功、重试、取消等状态都能被正确建模

## 9. 与 Phase 4 的衔接关系

Phase 4 的企业级扩展会把执行治理、租户隔离和分析能力继续放大。

因此，Phase 3 应该交付：

- 可规模化的执行模型
- 可治理的 agent runtime
- 可观测的任务证据链
- 可扩展的 workspace 协议与工具注册边界

否则 Phase 4 的“企业扩展”会建立在不稳定的执行层之上。
