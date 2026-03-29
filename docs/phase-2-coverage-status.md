# Phase 2 覆盖领域、模块、功能清单、优先级与实现情况

更新时间：2026-03-29

## 1. Phase 2 定位

Phase 2 的名称是：

- `Supply-Chain Transaction Loop`
- 中文可表述为：`供应链交易闭环`

它是整个项目第一个真正落地业务价值的阶段，目标不是“做一个大而全 ERP”，而是做成一条从主数据到采购、库存、审批，再到应付/应收基础的最小可用闭环。

## 2. Phase 2 目标

### 2.1 核心目标

- 建立可被真实使用的供应链最小业务链路
- 让业务命令、审批和库存变化进入真正的事务域
- 为 Phase 3 的 Agent 执行面提供真实可调用的业务对象

### 2.2 推荐的最小闭环

建议 Phase 2 的最小闭环按下面口径定义：

1. 主数据建立
2. 发起采购需求或采购订单
3. 走审批
4. 收货入库
5. 形成库存可用状态
6. 形成应付基础记录

如果范围允许，再补一条轻量销售出库链作为第二条闭环。

## 3. Phase 2 领域覆盖矩阵

| 有界上下文 | 功能范围 | 优先级 | 当前实现情况 | 说明 |
| --- | --- | --- | --- | --- |
| Master Data | 客户、供应商、商品、仓库、库位、税码、币种、计量单位、价目表 | P0 | Wave 1 已完成（供应商 / 商品 / 仓库） | 已具备最小主数据建档能力，可支撑第一条采购闭环。 |
| Procurement | 请购、采购订单、供应商事务状态、收货计划 | P0 | Wave 1 已完成（采购订单建单 / 提交 / 审批） | 已形成采购订单草稿、提交审批、审批通过/拒绝的最小交易流。 |
| Inventory | 入库、出库、预留、调拨、库存台账、可用/预留状态 | P0 | Wave 2 已启动（收货入库 / inbound ledger / on-hand 查询） | 已把入库事实和 on-hand 库存建立起来，出库、预留和更完整库存视图仍待补齐。 |
| Approval and Workflow | 审批定义、审批实例、人工任务、流程推进 | P0 | Wave 1 已完成（审批请求实例、通过 / 拒绝） | 当前先落地审批实例状态机，模板、人工任务和更复杂流程留到后续波次。 |
| Receivable and Payable Basics | 应付单、应收单、付款计划、开票基础引用 | P1 | 未开始 | Phase 2 只需要基础账务引用，不需要完整总账。 |
| Sales Minimal Loop | 报价、销售订单、发运计划、订单生命周期 | P1 | 未开始 | 可以作为第二条交易闭环补上，但不应压过采购/库存主线。 |
| Agent Task and Automation | Agent 通过命令与查询工具调用业务动作 | P2 | 底座已就绪，业务命令已出现 | 已有真实主数据 / 采购 / 审批命令和详情查询边界，可在 Phase 3 继续工具化接入。 |

## 4. Phase 2 模块覆盖矩阵

### 4.1 Domain Plane

| 模块 | 功能清单 | 优先级 | 当前实现情况 |
| --- | --- | --- | --- |
| Master Data Domain | 供应商、商品、仓库、单位、币种、税码等实体和值对象 | P0 | Wave 1 已完成（供应商 / 商品 / 仓库实体） |
| Procurement Domain | 请购单、采购订单、订单状态、收货计划、取消/提交/审批规则 | P0 | Wave 1 已完成（采购订单状态机） |
| Inventory Domain | 库存流水、库存预留、入出库事务、调拨、台账与可用量计算 | P0 | Wave 2 部分完成（receipt / inbound ledger / balance） |
| Approval Domain | 审批模板、审批实例、任务分派、通过/拒绝/撤回规则 | P0 | Wave 1 已完成（审批请求状态机） |
| Payable Basics Domain | 应付单、应付状态、付款计划引用 | P1 | 未开始 |
| Sales Domain | 销售订单、发运、扣减库存、订单生命周期 | P1 | 未开始 |

### 4.2 Application Plane

| 模块 | 功能清单 | 优先级 | 当前实现情况 |
| --- | --- | --- | --- |
| Command Handlers | 创建主数据、创建采购单、提交流程、审批通过/拒绝、收货入库 | P0 | Wave 2 已扩展到收货入库 |
| Query Services | 列表、详情、状态视图、库存余额、审批待办 | P0 | Wave 2 已扩展到采购详情 + on-hand 库存余额 |
| Transaction Coordination | 领域事务、策略校验、审计记录、outbox 协调 | P0 | 底座已接入 supply-chain service |
| Read Model Builders | Backoffice 读模型、Workspace 读模型、Monitoring 读模型 | P1 | 未开始 |

### 4.3 Interface Plane

| 模块 | 功能清单 | 优先级 | 当前实现情况 |
| --- | --- | --- | --- |
| Admin API | 主数据维护、采购单操作、审批动作、库存查询 | P0 | Wave 2 已扩展到收货动作和库存余额查询 |
| Workspace API | Agent Workspace 的查询与动作入口 | P1 | 路由占位已存在，业务接口未开始 |
| Integration API | 外部系统推送、同步与回调入口 | P2 | 路由占位已存在，业务接口未开始 |

### 4.4 Data Plane

| 模块 | 功能清单 | 优先级 | 当前实现情况 |
| --- | --- | --- | --- |
| OLTP Schema | 主数据、采购、库存、审批、应付应收基础表 | P0 | Wave 2 迁移已扩展到 receipt / receipt_line / inventory_ledger |
| Outbox / Event Publication | 领域事件可靠发布、消费确认、失败重试 | P0 | 仅有 outbox 表和 worker 骨架 |
| Query Projections | 列表视图、聚合报表、工作台视图 | P1 | 未开始 |

## 5. Phase 2 优先级清单

### 5.1 P0：必须先做的主线

- Master Data 基础实体
- Procurement 基础交易流
- Approval 基础审批流
- Inventory 入库与库存可用状态
- Command Handlers 与 Admin API
- Outbox 可靠发布主路径

### 5.2 P1：在主线跑通后补齐

- Payable Basics
- Sales Minimal Loop
- Query Read Models
- Workspace 查询入口

### 5.3 P2：可以作为收尾或 Phase 3 的衔接项

- 集成同步入口
- Agent 对业务命令的受控调用
- 较完整的报表与监控投影

## 6. 当前准备度判断

Phase 2 Wave 1 已经完成最小可执行切片，当前状态是“业务闭环已经起步，但库存与账务还没有接上”：

- 已有命令管道接入 supply-chain service：主数据、采购、审批命令已走统一管道
- 已有供应商 / 商品 / 仓库主数据实体和内存仓储
- 已有采购订单状态机与审批请求状态机
- 已有 Admin API：可建档、建单、提交审批、审批通过/拒绝、收货入库、查询详情和 on-hand 库存
- 已有 inbound receipt 和 inventory ledger：库存事实开始从业务事务产生
- 已有 Phase 2 Wave 1 / Wave 2 迁移：为 PostgreSQL 仓储预留主数据、采购、审批、库存表结构

但同时仍有几个明显缺口：

- 没有真实 Tenant / IAM 数据模型
- 还没有出库、预留、调拨和 reserved 维度
- 没有应付基础记录生成
- 没有独立 read model / projection
- PostgreSQL repository 仍未落地，当前运行时仍使用内存仓储

结论是：

- `Phase 2 已正式开工，Wave 1 已落地`
- `Wave 2 已开始，下一步应继续把库存从 on-hand 扩到更完整的可用状态与后续账务衔接`

## 7. 推荐的分波次落地方式

### Wave 1：主数据 + 采购 + 审批骨架

- 供应商
- 商品
- 仓库
- 采购订单
- 审批模板与审批实例
- 基础 Admin API

当前状态：

- `已完成（审批模板仍未引入，先落地审批实例状态机）`
- 已有集成测试覆盖主数据建档、采购建单、提交流程、详情查询、审批通过、未知供应商 404

### Wave 2：收货入库 + 库存真相

- 收货单
- 入库事务
- 库存台账
- 可用库存/预留库存计算
- 采购完成态推进

下一步 backlog：

- `已完成：收货单与采购订单关联`
- `已完成：入库后驱动 inventory ledger 和 on-hand 余额查询`
- `已完成：将采购流程从“已审批”推进到“已收货”`
- 待补：reserved / available 维度
- 待补：出库、预留、调拨等非 inbound movement

### Wave 3：应付基础 + 轻量销售

- 应付基础记录
- 付款计划引用
- 可选的销售出库最小链路
- 读模型与列表页支持

后续 backlog：

- 通过已批准且已收货的采购对象生成应付基础记录
- 评估是否在 Wave 3 同时引入最小销售闭环
- 补齐列表、待办和台账类读模型

## 8. Phase 2 完成判定建议

只有当下面几件事成立时，才能认为 Phase 2 进入完成态：

- 主数据、采购、审批、库存形成一条可测试闭环
- 至少一条真实采购到收货入库链路可跑通
- 库存状态来自真实领域事务，而非临时脚本或工作流假写
- 应付基础引用已经建立
- Admin API 可以操作这条闭环
- 集成测试可以覆盖关键命令、审批与库存变化

## 9. 与 Phase 3 的衔接关系

Phase 3 要强化 Agent 执行面，但它必须建立在 Phase 2 真实业务对象之上。

因此，Phase 2 的交付应该为 Phase 3 提供：

- 可查询的业务对象
- 可执行的命令处理器
- 可审计的审批节点
- 可观测的库存与采购状态
- 可被工具化暴露的查询/命令边界

如果做不到这一点，Phase 3 的 Agent 只能操作空壳流程。

## 10. Phase 2 Strict Release Gate

> 说明：以下清单用于建立严格发布门禁基线。当前为本轮基线记录，均保持未勾选；最终严格 sign-off 仅在后续任务完成并补齐证据后进行。

- [ ] Wave 2 plan task list completed
- [ ] supplychain unit tests pass
- [ ] integration inventory flow passes
- [ ] compose migration contract passes
- [ ] full repo tests pass

### Baseline command outcomes

1. `go test ./internal/application/admin/supplychain ./internal/infrastructure/persistence/memory -v`
   - Local baseline evidence (2026-03-29, package scope, not final sign-off): PASS
2. `go test ./test/integration -run 'TestAdminSupplyChainFlow|TestAdminInventoryReceiptFlow|TestAdminInventoryReceiptRequiresApprovedOrder|TestPhase2Wave2MigrationContract' -v`
   - Local baseline evidence (2026-03-29, focused integration scope, not final sign-off): PASS
3. `go test ./...`
   - Local baseline evidence (2026-03-29, full repository scope, not final sign-off): PASS

### Wave 2 baseline evidence rule

- Wave 2 唯一基线为：`docs/superpowers/plans/2026-03-26-phase-2-wave-2-inbound-inventory-implementation-plan.md`。
- 最终严格 sign-off 必须提供该基线计划中任务清单已完成（全部 completed）的证据；在证据补齐前，本节仅作为“测试证据基线”而非最终发布结论。
- 补充说明：在本计划内 strict negative-path inventory validations 实现后，strict integration gate 将扩展纳入 additional negative-path tests。
