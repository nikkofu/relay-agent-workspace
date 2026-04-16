# Phase 8: 动态化运行所需 API 扩充清单

## 1. 结论

当前 `Relay Agent Workspace` 仍然是 **纯前端 Mock 驱动**：

- 工作区、频道、消息、用户都来自 [lib/mock-data.ts](/Users/admin/Documents/WORK/ai/acim-ui/lib/mock-data.ts:1)
- 页面状态主要保存在 Zustand store 中：
  - [stores/workspace-store.ts](/Users/admin/Documents/WORK/ai/acim-ui/stores/workspace-store.ts:1)
  - [stores/channel-store.ts](/Users/admin/Documents/WORK/ai/acim-ui/stores/channel-store.ts:1)
  - [stores/message-store.ts](/Users/admin/Documents/WORK/ai/acim-ui/stores/message-store.ts:1)
  - [stores/ui-store.ts](/Users/admin/Documents/WORK/ai/acim-ui/stores/ui-store.ts:1)
- AI 对话也是本地模拟流式：
  - [hooks/use-ai-chat.ts](/Users/admin/Documents/WORK/ai/acim-ui/hooks/use-ai-chat.ts:1)

这意味着，如果项目要真正“动态运行起来”，后端至少要覆盖以下 3 类能力：

1. 基础协作数据：工作区、频道、消息、线程、DM、用户资料。
2. 生产力能力：搜索、Later、Activity、AI 对话、AI 摘要、Canvas。
3. 实时能力：消息流、未读数、在线状态、Typing/AI 流式输出。

## 2. 分析依据

本清单不是凭空设计，而是从当前 UI 和交互直接反推：

- 工作区容器与右侧面板来源：
  - [app/workspace/layout.tsx](/Users/admin/Documents/WORK/ai/acim-ui/app/workspace/layout.tsx:1)
- 主频道页与消息区：
  - [app/workspace/page.tsx](/Users/admin/Documents/WORK/ai/acim-ui/app/workspace/page.tsx:1)
  - [components/layout/message-area.tsx](/Users/admin/Documents/WORK/ai/acim-ui/components/layout/message-area.tsx:1)
- 频道侧边栏、搜索入口、DM 入口：
  - [components/layout/channel-sidebar.tsx](/Users/admin/Documents/WORK/ai/acim-ui/components/layout/channel-sidebar.tsx:1)
- 消息发送、AI Slash Command、Canvas 唤起：
  - [components/message/message-composer.tsx](/Users/admin/Documents/WORK/ai/acim-ui/components/message/message-composer.tsx:1)
- 消息悬浮操作、Reaction、Thread：
  - [components/message/message-item.tsx](/Users/admin/Documents/WORK/ai/acim-ui/components/message/message-item.tsx:1)
  - [components/message/message-actions.tsx](/Users/admin/Documents/WORK/ai/acim-ui/components/message/message-actions.tsx:1)
  - [components/message/emoji-reaction.tsx](/Users/admin/Documents/WORK/ai/acim-ui/components/message/emoji-reaction.tsx:1)
  - [components/layout/thread-panel.tsx](/Users/admin/Documents/WORK/ai/acim-ui/components/layout/thread-panel.tsx:1)
- 搜索、Activity、Later：
  - [components/search/search-dialog.tsx](/Users/admin/Documents/WORK/ai/acim-ui/components/search/search-dialog.tsx:1)
  - [app/workspace/activity/page.tsx](/Users/admin/Documents/WORK/ai/acim-ui/app/workspace/activity/page.tsx:1)
  - [app/workspace/later/page.tsx](/Users/admin/Documents/WORK/ai/acim-ui/app/workspace/later/page.tsx:1)
- AI Chat、Thread Summary、Canvas：
  - [components/ai-chat/ai-chat-panel.tsx](/Users/admin/Documents/WORK/ai/acim-ui/components/ai-chat/ai-chat-panel.tsx:1)
  - [components/ai-chat/ai-thread-summary.tsx](/Users/admin/Documents/WORK/ai/acim-ui/components/ai-chat/ai-thread-summary.tsx:1)
  - [components/layout/canvas-panel.tsx](/Users/admin/Documents/WORK/ai/acim-ui/components/layout/canvas-panel.tsx:1)

## 3. 最小可运行 API

这是“页面不再依赖 Mock，能真实切换工作区、浏览频道、收发消息”的最小集合。

### 3.1 工作区与当前用户

- `GET /api/me`
  - 返回当前登录用户、默认工作区、权限、头像、状态。
- `GET /api/workspaces`
  - 返回用户可访问的工作区列表。
- `GET /api/workspaces/:workspaceId`
  - 返回工作区详情。

### 3.2 频道侧边栏

- `GET /api/workspaces/:workspaceId/channels`
  - 返回频道列表，至少包含：
  - `id`
  - `name`
  - `type`
  - `description`
  - `memberCount`
  - `isStarred`
  - `unreadCount`
- `POST /api/workspaces/:workspaceId/channels`
  - 新建频道。
- `PATCH /api/channels/:channelId`
  - 编辑频道信息、星标状态、描述等。

### 3.3 主消息区

- `GET /api/channels/:channelId/messages`
  - 分页拉取消息，支持 `cursor`、`limit`。
- `POST /api/channels/:channelId/messages`
  - 发送频道消息。
- `PATCH /api/messages/:messageId`
  - 编辑消息。
- `DELETE /api/messages/:messageId`
  - 删除消息。

### 3.4 基础线程

- `GET /api/messages/:messageId/thread`
  - 返回 thread 根消息、回复列表、replyCount、lastReplyAt。
- `POST /api/messages/:messageId/thread/replies`
  - 发送 thread reply。

## 4. 结合当前 UI 必须补齐的扩展 API

这部分是当前界面已经做出来，但没有真实数据源时会直接“空转”的能力。

### 4.1 Reaction / Pin / Save / Share / Unread

这些能力直接来自消息悬浮操作和消息状态展示。

- `POST /api/messages/:messageId/reactions`
  - 请求体：`{ emoji: "👋" }`
- `DELETE /api/messages/:messageId/reactions/:emoji`
- `POST /api/messages/:messageId/pin`
- `DELETE /api/messages/:messageId/pin`
- `POST /api/messages/:messageId/save`
- `DELETE /api/messages/:messageId/save`
- `POST /api/messages/:messageId/mark-unread`
- `GET /api/messages/:messageId/permalink`

### 4.2 DM 页面

当前 [app/workspace/dms/page.tsx](/Users/admin/Documents/WORK/ai/acim-ui/app/workspace/dms/page.tsx:1) 已经存在独立列表页，所以至少需要：

- `GET /api/dms`
  - 返回 DM 会话列表、最后一条消息、未读数、对方用户资料。
- `GET /api/dms/:dmId/messages`
- `POST /api/dms/:dmId/messages`
- `POST /api/dms`
  - 新建或打开一个 DM 会话。

### 4.3 Activity 页面

- `GET /api/activity`
  - 支持 `cursor`、`type`、`workspaceId`。
  - 至少覆盖：
  - mention
  - reply
  - reaction
  - assignment
  - AI 结果通知

### 4.4 Later 页面

- `GET /api/later`
  - 返回用户保存的消息、文件、文档。
- `POST /api/later/items`
  - 保存某个对象到 Later。
- `DELETE /api/later/items/:itemId`

### 4.5 全局搜索

当前搜索框已经明确表达“Ask AI to search across workspace”，因此普通搜索和语义搜索都要有。

- `GET /api/search`
  - 关键词搜索。
  - 参数建议：`q`, `workspaceId`, `type=message|channel|user|file`, `cursor`
- `POST /api/search/semantic`
  - 语义搜索。
  - 请求体建议：`{ query, workspaceId, scopes }`
- `GET /api/search/suggestions`
  - 返回 recent / suggested / channel / people 混合建议。

## 5. AI 能力 API

这是 Phase 8 最需要单独规划的一组，因为当前 UI 已经有 AI Chat、AI Summary、AI Slash Command、Canvas，但实现都还是本地假数据。

### 5.1 AI Chat 主接口

- `POST /api/ai/chat`
  - 建议使用 SSE 或 chunked streaming。
  - 输入至少包含：
  - `workspaceId`
  - `channelId`
  - `threadId?`
  - `messages`
  - `mode`：`chat | summarize | draft | translate | explain | optimize`
  - `contextScope`
- `GET /api/ai/conversations/:conversationId`
  - 拉历史对话。
- `GET /api/ai/conversations`
  - 返回会话列表。

### 5.2 AI Slash Commands

对应当前 [components/ai-chat/ai-slash-command.tsx](/Users/admin/Documents/WORK/ai/acim-ui/components/ai-chat/ai-slash-command.tsx:1) 的命令集：

- `/ai`
- `/summarize`
- `/translate`
- `/explain`
- `/draft`
- `/optimize`
- `/canvas`

推荐不要拆成 7 个后端接口，而是统一走：

- `POST /api/ai/commands/execute`
  - 请求体：`{ command, workspaceId, channelId, threadId?, input, selection?, messageId? }`

### 5.3 Thread Summary

- `POST /api/ai/threads/:threadId/summary`
  - 返回摘要、关键决策、action items、participants。

### 5.4 Channel Summary

因为首页 suggestion 和 slash command 都有 “Summarize this channel”：

- `POST /api/ai/channels/:channelId/summary`
  - 支持 `timeRange`、`includeFiles`、`includeActionItems`。

### 5.5 Draft / Rewrite / Translate / Explain

如果不想让 `POST /api/ai/chat` 过重，也可以分离成生产力接口：

- `POST /api/ai/draft`
- `POST /api/ai/rewrite`
- `POST /api/ai/translate`
- `POST /api/ai/explain`

但从维护成本看，仍建议统一归入 `POST /api/ai/commands/execute`。

## 6. Canvas / Artifact API

当前 Canvas 面板已经是一个独立文档工作流，不只是弹窗。

- `GET /api/canvas`
  - 返回当前用户可见的 Canvas 列表。
- `POST /api/canvas`
  - 创建新 Canvas。
- `GET /api/canvas/:canvasId`
  - 获取文档详情。
- `PATCH /api/canvas/:canvasId`
  - 更新标题、正文、状态、版本号。
- `POST /api/canvas/:canvasId/share`
  - 分享给用户或频道。
- `POST /api/ai/canvas/generate`
  - AI 生成文档初稿。

如果后端希望更规范，可把 Canvas 视为 docs/artifacts：

- `GET /api/artifacts`
- `POST /api/artifacts`
- `GET /api/artifacts/:artifactId`
- `PATCH /api/artifacts/:artifactId`

## 7. 用户资料与在线状态 API

用户头像、hover card、DM、成员数展示都依赖用户信息。

- `GET /api/users/:userId`
- `GET /api/workspaces/:workspaceId/members`
- `GET /api/users/:userId/presence`
- `PATCH /api/users/:userId/status`

如果要支撑资料卡里的 AI Insight：

- `GET /api/users/:userId/insights`

## 8. 文件与附件 API

虽然当前只在 Mock 里展示了 link attachment，但消息编辑器已留出附件入口。

- `POST /api/files/upload`
- `GET /api/files/:fileId`
- `GET /api/messages/:messageId/attachments`
- `POST /api/messages/:messageId/attachments`

## 9. 实时与流式 API

只做 REST 不够。这个项目如果要像 Slack 一样“动态”，至少还要一条实时通道。

### 9.1 建议的实时通道

- `GET /api/realtime/events`
  - SSE 方案，适合快速接前端。
- 或 `WS /api/realtime`
  - WebSocket 方案，适合消息、presence、typing、AI 状态统一推送。

### 9.2 实时事件最少集合

- `message.created`
- `message.updated`
- `message.deleted`
- `reaction.updated`
- `thread.updated`
- `channel.updated`
- `presence.updated`
- `activity.created`
- `ai.response.delta`
- `ai.response.completed`

## 10. 推荐的最小落地顺序

如果目标不是一次性做全，而是先把当前 UI 跑成“真动态”，推荐分三批：

### 第一批：必须先做

- `GET /api/me`
- `GET /api/workspaces`
- `GET /api/workspaces/:workspaceId/channels`
- `GET /api/channels/:channelId/messages`
- `POST /api/channels/:channelId/messages`
- `GET /api/messages/:messageId/thread`
- `POST /api/messages/:messageId/thread/replies`

### 第二批：让 Slack 核心交互可用

- Reaction / Pin / Save / Unread / Permalink
- `GET /api/dms`
- `GET /api/activity`
- `GET /api/later`
- `GET /api/search`

### 第三批：把 AI-Native 特色做实

- `POST /api/ai/chat`
- `POST /api/ai/commands/execute`
- `POST /api/ai/threads/:threadId/summary`
- `POST /api/ai/channels/:channelId/summary`
- Canvas 全套接口
- realtime 通道

## 11. 建议的接口分组总表

### 核心协作域

- `/api/me`
- `/api/workspaces`
- `/api/channels`
- `/api/messages`
- `/api/threads`
- `/api/dms`
- `/api/activity`
- `/api/later`
- `/api/search`
- `/api/users`
- `/api/files`

### AI 域

- `/api/ai/chat`
- `/api/ai/commands/execute`
- `/api/ai/threads/:threadId/summary`
- `/api/ai/channels/:channelId/summary`
- `/api/ai/canvas/generate`

### 实时域

- `/api/realtime/events`
- 或 `WS /api/realtime`

## 12. 当前阶段的判断

如果只以“让 `/workspace` 真正动态化”为目标，**最关键缺的不是单个接口，而是后端领域模型和实时机制**。  
当前前端已经把 Slack 风格的壳子搭完了，下一阶段后端至少要同时提供：

- 面向工作区的层级数据模型
- 面向消息/线程的分页与写入能力
- 面向 AI 的统一执行入口
- 面向 UI 刷新的实时事件流

否则即使补了几个 REST 接口，页面也只能从“本地 Mock”变成“远程 Mock”，还不能算真正跑起来。
