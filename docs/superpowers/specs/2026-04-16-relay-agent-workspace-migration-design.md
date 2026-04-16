# Relay Agent Workspace Migration Design

## Goal

将当前 `acim-ui` 仓库整体重定位为 `Relay Agent Workspace`，保留现有代码和历史演进脉络，同时完成品牌、仓库元信息、README、核心文档和迁移说明的统一更新，并最终迁移到新仓库：

- `https://github.com/nikkofu/relay-agent-workspace`

本次工作不是“新开一个产品”，而是对现有仓库做一次完整的产品重命名与定位升级。

## Scope

本次迁移包含：

1. 产品命名与定位统一
2. GitHub repo metadata 文案准备
3. README 重写
4. 核心文档同步更新
5. 迁移说明文档落档
6. 旧命名残留扫描与替换
7. 为后续切换 remote 做准备

本次迁移不包含：

1. 后端代码实现
2. 数据模型重构
3. API 实现
4. Agent runtime 实现
5. 立即切换 remote 并 push 到新仓库

最后一步会在用户确认后单独执行。

## Current State

当前仓库状态：

- 仓库名与项目名仍然是 `acim-ui`
- 没有 `README.md`
- 已存在的主要文档包括：
  - `docs/implementation_plan.md`
  - `docs/task.md`
  - `docs/phase8-api-expansion.md`
- 当前 git remote 仍指向：
  - `https://github.com/nikkofu/acim-ui`
- 当前工作树已有未提交改动，因此迁移不能假设仓库干净

## Positioning

### Product Name

- Full name: `Relay Agent Workspace`
- Short name: `Relay`

### Positioning Statement

`Relay Agent Workspace` 是一个面向 humans + agents 的 AI-native collaboration workspace。它以即时通讯与工作区协作为入口，将 messaging、threads、AI chat、artifacts、agent execution 和 orchestration 统一在一个 workspace 中。

### Messaging

推荐对外统一采用以下表述：

- AI-native collaboration workspace for humans and agents
- Instant messaging, artifacts, and orchestration for AI teams
- A shared workspace where people and agents think, chat, and execute together

### Why This Positioning

该定位兼顾了当前项目的两个现实：

1. 当前 UI 的核心入口仍然是 Slack 风格 workspace + messaging
2. 项目真正的差异化目标是 agents collaboration，而不只是 chat with AI

因此不应继续沿用 `acim-ui` 这类内部代号，也不应把项目简单表述成 Slack clone。

## Repository Metadata

需要产出并统一一套 repo 级别文案：

### Repo Name

- `relay-agent-workspace`

### GitHub Description

主推荐：

- `AI-native collaboration workspace for humans and agents, combining messaging, artifacts, and orchestration.`

备选：

- `Instant messaging, AI chat, artifacts, and agent orchestration in one shared workspace.`
- `A shared workspace for human-agent collaboration with realtime messaging and AI-native workflows.`

### Topics

建议 GitHub topics：

- `ai`
- `agents`
- `multi-agent`
- `collaboration`
- `workspace`
- `messaging`
- `realtime`
- `nextjs`
- `typescript`
- `ui`

## README Strategy

README 需要从零开始创建，而不是拼接旧文档。

结构建议：

1. 标题与一句话定位
2. 核心能力概览
3. 为什么是 Relay
4. 当前技术栈
5. 当前项目状态
6. 本地开发方式
7. 近期 roadmap
8. 仓库结构概览

README 不应再使用“Slack 1:1 复刻实施计划”作为主叙事。该内容可以保留在实现文档里，但不能继续作为仓库首页定位。

## Documentation Strategy

以下文档应同步调整：

### Update

- `package.json`
  - `name` 从 `acim-ui` 改为 `relay-agent-workspace`
- `docs/implementation_plan.md`
  - 将项目名与对外定位同步为新品牌
  - 保留历史技术计划内容，但弱化 “Slack 1:1 复刻” 的主标题权重
- `docs/task.md`
  - 将项目标题同步到新名称
- `docs/phase8-api-expansion.md`
  - 将上下文名从 `acim-ui` / `ACIM-UI` 同步为 `Relay Agent Workspace`

### Create

- `README.md`
- `docs/project-positioning.md`
- `docs/repo-migration-plan.md`

其中：

- `project-positioning.md` 用于沉淀品牌、目标用户、能力边界、产品定位
- `repo-migration-plan.md` 用于记录从旧仓库迁移到新仓库的具体步骤和注意事项

## Rename Strategy

本次迁移采用“整体重定位”策略：

1. 在当前仓库内完成所有文本和品牌更新
2. 保留历史提交记录
3. 不创建第二份平行代码库
4. 待用户确认后，再切换 remote 到新仓库

### Residual Name Handling

需要扫描以下旧命名：

- `acim-ui`
- `ACIM-UI`
- `ACIM`

处理原则：

1. 对外命名全部替换为 `Relay Agent Workspace` 或 `Relay`
2. 历史上下文处，如果内容是“曾用代号/原始实现名”，可保留一次说明
3. 避免机械全量替换导致历史语境失真

## Git Migration Strategy

迁移分两阶段：

### Stage 1: Prepare In Place

在当前仓库完成：

- 文案与文档更新
- README 创建
- 新品牌定位落档
- 迁移计划落档
- 旧命名扫描

此阶段不切 remote。

### Stage 2: Switch Remote

待用户确认后：

1. 校验当前 worktree
2. 明确哪些改动属于本次迁移
3. 处理当前未提交改动的边界
4. 将 `origin` 从旧 repo 切换到新 repo
5. 再执行 push

## Risks

### Dirty Worktree

当前工作树已经存在与迁移无关的修改。迁移实施时必须避免：

- 覆盖用户现有改动
- 把无关改动错误归类为迁移内容

### Naming Drift

如果只更新 README 而不更新实施文档，仓库会出现“新品牌 + 旧项目名”并存的混乱状态。

### Over-Renaming

如果将所有历史文档中的 `Slack`、`ACIM`、`clone` 表述机械替换，可能损失项目真实演进背景。需要进行语义级改写，而不是字符串替换。

## Recommended Execution Order

1. 创建迁移设计 spec
2. 创建 repo positioning 文档
3. 创建 README
4. 创建 migration plan 文档
5. 更新 `package.json`
6. 更新核心文档中的项目命名与定位
7. 扫描残留旧命名
8. 用户 review
9. 确认后再切换 remote

## Deliverables

本次应交付：

1. `README.md`
2. `docs/project-positioning.md`
3. `docs/repo-migration-plan.md`
4. 更新后的 `package.json`
5. 更新后的核心文档
6. 一套可直接用于 GitHub 的 repo description / tagline / topics

## Approval Gate

在用户 review 并确认该 spec 之前，不执行实际迁移和大规模文本改写。
