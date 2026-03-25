# Phase 4 覆盖领域、模块、功能清单、优先级与实现情况

更新时间：2026-03-25

## 1. Phase 4 定位

Phase 4 的名称是：

- `Enterprise Expansion`
- 中文可表述为：`企业级扩展`

这个阶段的目标不是“再加几个模块”，而是把平台从“具备平台与业务主线”推进到“能够支撑更复杂租户、更复杂集成、更高治理要求”的企业级系统。

## 2. Phase 4 目标

### 2.1 核心目标

- 开放集成平台
- 专属租户 cell / 更强隔离模型
- 更高级的搜索、分析与监控
- 更深的 workflow 和 model governance

### 2.2 进入条件

Phase 4 必须建立在下面两个前提之上：

- 至少一个真实业务闭环已经稳定运行
- Agent 执行面已经具备可治理能力

如果没有这两个前提，Phase 4 会退化成“提前建设过重平台”。

## 3. Phase 4 覆盖矩阵

| 方向 | 功能范围 | 优先级 | 当前实现情况 | 说明 |
| --- | --- | --- | --- | --- |
| Open Integration Platform | 连接器注册、外部系统同步、插件化适配、租户启用控制 | P0 | 未开始 | 这是平台化扩张的第一能力。 |
| Dedicated Tenant Cells | 专属 schema / database / execution isolation / cache namespace / object prefix | P0 | 部分底座已准备 | `CellRoute` 已有基础结构，但没有真正的多 cell 路由与部署模型。 |
| Advanced Search | 全文检索、过滤搜索、跨对象搜索 | P1 | 未开始 | 设计已明确 Phase 1 可先保持简单，Phase 4 再做强化。 |
| Advanced Analytics | 监控指标、分析快照、执行统计、业务分析 | P1 | 未开始 | 当前只有 OTEL / Prometheus / Grafana compose 合同，没有分析层。 |
| Deeper Workflow Governance | 更复杂流程编排、策略治理、审批编排、模型治理 | P1 | 未开始 | 这是企业复杂流程的扩展层。 |
| Multi-Cell Operations | shared cell A/B、dedicated cell X、路由演进与运维治理 | P1 | 未开始 | 设计已给出演进路径，但实现尚未开始。 |
| Compliance and Audit Expansion | 更细的证据链、审计检索、跨租户运维审计、保留策略 | P2 | 未开始 | 当前只有基础审计骨架。 |

## 4. Phase 4 模块覆盖矩阵

### 4.1 Control Plane

| 模块 | 功能清单 | 优先级 | 当前实现情况 |
| --- | --- | --- | --- |
| Tenant Cell Governance | 租户路由、cell 策略、隔离等级、迁移策略 | P0 | 部分底座已准备，真实治理未开始 |
| Feature Flag / Quota | 租户能力分级、资源配额、模型与工具额度 | P1 | 未开始 |
| Plugin / Connector Registry | 插件、连接器、租户启用、版本治理 | P0 | 未开始 |
| Model Governance | 模型目录、风险级别、租户允许列表、审批门槛 | P1 | 未开始 |

### 4.2 Data and Integration Plane

| 模块 | 功能清单 | 优先级 | 当前实现情况 |
| --- | --- | --- | --- |
| Multi-Cell Data Routing | schema / db / cache / object storage / event topic prefix 路由 | P0 | 仅有 `CellRoute` 基础结构 |
| Search Platform | PostgreSQL FTS 强化、trigram、后续 OpenSearch 演进 | P1 | 未开始 |
| Analytics Store | 监控、业务分析、执行分析、快照报表 | P1 | 未开始 |
| Connector Runtime | webhook、ERP/SCM/WMS 连接器、同步与回调 | P0 | 未开始 |

### 4.3 Execution Plane

| 模块 | 功能清单 | 优先级 | 当前实现情况 |
| --- | --- | --- | --- |
| Advanced Workflow Orchestration | 更复杂的长事务编排、补偿、回放、跨系统协调 | P1 | 未开始 |
| Model and Tool Governance | 工具风险治理、模型审批、审计增强、租户差异化能力 | P1 | 未开始 |
| Enterprise Monitoring | queue depth、task timeline、policy decision metrics、异常告警 | P1 | 部分 observability 合同已准备 |

## 5. Phase 4 优先级清单

### 5.1 P0：企业扩展的第一层

- Connector / integration platform
- Dedicated tenant cell routing
- Multi-cell operations
- 租户级插件/连接器治理

### 5.2 P1：规模化能力增强

- Advanced search
- Advanced analytics
- Workflow / model governance
- 统一可观测性与运营视图

### 5.3 P2：治理深挖和平台化优化

- 更深的审计与合规策略
- 更细的计量、配额与租户差异化控制
- 更复杂的跨系统流程编排

## 6. 当前准备度判断

Phase 4 虽然尚未启动，但并不是从零开始，因为底座已经提供了若干前置能力：

- `CellRoute` 已经把 tenant routing 作为一等对象引入
- 配置、runtime、event bus、command pipeline 已为多运行时扩展打底
- Compose 中已经有 OTEL / Prometheus / Grafana 合同
- Integration API 路由已经预留分组入口

但这些都只是准备度，不代表 Phase 4 已有实体实现。

结论：

- `Phase 4 当前处于“底座有准备、能力未落地”状态`

## 7. 推荐的分波次落地方式

### Wave 1：集成与租户扩展

- connector registry
- integration runtime
- dedicated tenant cell routing
- 多 cell 运维基础能力

### Wave 2：搜索与分析

- query/read model 增强
- advanced search
- analytics snapshot
- 平台运营与业务监控视图

### Wave 3：深度治理

- model governance
- tool governance
- workflow governance
- 更强的审计与合规策略

## 8. Phase 4 完成判定建议

可以把 Phase 4 完成态定义为：

- 平台能够支撑多类租户隔离模式
- 集成平台可以承载真实企业外部系统对接
- 搜索、分析、监控可以支撑运营与治理
- 模型、工具、workflow 都具备更深的企业级控制能力

## 9. 对实施节奏的提醒

Phase 4 很容易被误做成“提前上很多重平台能力”，这是要避免的。

正确节奏应该是：

- 先让真实业务闭环和 Agent 执行面稳定
- 再把扩展能力建立在真实瓶颈之上

因此，Phase 4 更适合在前两三个阶段已经跑出真实业务压力之后进入实施。
