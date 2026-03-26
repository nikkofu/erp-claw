# Phase Handoff Playbook

本目录用于保存“可恢复执行”的项目级交接文档，避免会话切换后出现上下文丢失。

## 目标

- 明确当前分支快照
- 给出新鲜验证证据
- 把“已完成”和“未完成”拆开
- 给出明天第一条命令和第一件最小任务

## 文件命名

- `YYYY-MM-DD-<topic>-handoff.md`

示例：

- `2026-03-26-phase2-wave6-handoff.md`

## 交接前最低要求

- 写明 `worktree + branch + head sha`
- 写明 `origin/main` 与阶段分支远端 sha
- 写明新鲜验证命令和结果
- 写明 merge 安全状态（`safe_to_merge` / `blocked`）
- 写明明天第一任务和 acceptance criteria

## 推荐流程

1. 使用 `skills/phase-handoff-playbook/handoff-template.md` 起草
2. 对照 `skills/phase-handoff-playbook/handoff-quality-checklist.md` 自检
3. 在次日开工前执行 `skills/phase-handoff-playbook/resume-checklist.md`
