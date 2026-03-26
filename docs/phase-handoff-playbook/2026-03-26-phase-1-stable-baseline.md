# Phase 1 Stable Baseline

更新时间：2026-03-26

## 1. 文档目的

本文用于记录当前 `Phase 1` 的稳定交付基线，作为推送远程后的协作契约。

它回答四个问题：

- 当前稳定基线覆盖了哪些能力
- 哪些能力已经可以被 `Phase 2` 或其他并行分支依赖
- 当前仍有哪些明确缺口，不应被误判为“已完成”
- 这次稳定基线实际对应哪些 git 提交

## 2. 当前稳定快照

- 目标分支：`feature/phase1-control-plane`
- 目标工作树：`.worktrees/phase1-control-plane`
- 根仓库状态：`main` 工作区已清理干净，不再残留误落的未跟踪 `Phase 1` 文件
- 版本号：`0.2.6`
- bootstrap 合同：运行时 catalog 在非测试路径下失败即中止，不再静默回退为内存存储
- Fresh 验证命令：`GOCACHE=$(pwd)/.cache/go-build go test ./... -count=1`

## 3. 本次稳定基线覆盖范围

### 3.1 平台与控制面

- tenant / user / agent profile catalog baseline
- role / department catalog
- user-role / user-department binding
- Admin API create/list/bind 最小闭环
- repository-backed policy rule lifecycle
- audit event store / query baseline

### 3.2 Agent Runtime 与 Workspace

- session / task domain model 与 Postgres repository
- runtime service 的 open / close / fail / cancel 状态流转
- workspace event publish / replay seam
- workspace sessions / tasks / events 最小写入/查询 HTTP surface
- 以 `session_key` 为统一契约的 task query / event replay / SSE stream read side

### 3.3 Capability / Approval / Reliability

- tenant-scoped model catalog baseline
- tenant-scoped tool catalog baseline
- approval definition / instance / task baseline
- `REQUIRE_APPROVAL` 到 approval instance/task 的最小 pipeline 接线
- outbox dispatch / retry / failed recovery / stale `publishing` reclaim / poll instrumentation baseline

## 4. Phase 2 可依赖契约

在 `Phase 2` 或其他并行分支中，可以默认依赖以下稳定边界：

- `Admin API` 已存在控制面目录型接口，不需要重新从零搭 tenant/iam catalog
- `Workspace API` 已存在 sessions/tasks/events/stream 最小写入/查询/SSE 面，可作为后续更深 workspace protocol 的依赖底座
- `policy` / `audit` / `approval starter seam` 已存在，可在业务命令中继续复用
- `model catalog` / `tool catalog` 已存在，不需要重新设计租户级 capability 主键和仓储边界
- `outbox reliability baseline` 已存在，后续业务事件发布应沿用该路径，而不是绕开它
- tenant-scoped admin mutations 已要求 tenant root 先存在，后续业务建模应沿用这一边界

## 5. 当前明确未完成项

以下内容仍然不能被误判为已完成：

- Phase 2 供应链业务闭环，包括 master data、procurement、inventory、sales 与 receivable/payable
- 更完整的 capability policy binding、tenant enablement、plugin registry、quota / feature flag
- 更完整的 workflow orchestration、approval admin surface、多级审批与 approver resolution
- 真实 WebSocket / streaming protocol、跨进程 event replay 与更完整的 stream durability
- live DB round-trip、consumer-side idempotency、DLQ 与更完整的 reliability hardening

## 6. 协作约束

为避免并行协作时互相覆盖，后续分支默认遵守以下约束：

- `Phase 2` 优先消费当前稳定接口，不反向重写 `Phase 1` 已稳定的仓储主键、HTTP 路由和最小 pipeline 契约
- 若确实需要改动 `Phase 1` 契约，必须单独形成兼容性提交，而不是在业务分支里直接揉碎重做
- 新一轮并行开发建议从当前分支或其稳定提交新开 worktree，而不是在脏工作区继续叠加

## 7. 实际提交清单

当前稳定基线至少覆盖以下稳定切片：

- `6c99a48` `feat: add phase1 approval admin surface`
- `619d40b` `feat: add phase1 capability admin surface`
- `13e5f1f` `feat: add phase1 workspace write surface`
- `a366ac1` `feat: add phase1 governance admin surface`
- `0281578` `feat: add phase1 outbox operator surface`
- 当前这次 workspace streaming slice 与其配套文档/版本同步提交

## 8. 建议的下一步

1. 以当前稳定基线为起点邀请 Claude 在独立分支推进 `Phase 2`
2. 当前分支继续只处理 `Phase 1` 剩余项，不与 `Phase 2` 业务建模混写
3. Phase 1 后续优先顺序保持为：
   - capability policy binding / tenant enablement
   - approval / workflow orchestration
   - workspace WebSocket / cross-process streaming protocol
   - live DB reliability hardening
