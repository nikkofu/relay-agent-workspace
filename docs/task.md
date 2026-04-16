# Relay Agent Workspace 任务清单

## Phase 0: 项目脚手架 & 基础设施
- [x] 使用 shadcn CLI 初始化 Next.js 项目
- [x] 安装额外依赖（zustand, ai-sdk 等）
- [x] 批量安装 shadcn/ui 核心组件
- [ ] 安装 shadcn.io/ai 组件 (因 Auth 限制，将在后续阶段手动实现)
- [x] 创建类型定义 (types/index.ts)
- [x] 创建 Mock 数据 (lib/mock-data.ts)
- [x] 创建 Zustand stores

## Phase 1: Slack 核心三栏布局
- [x] 主导航栏 (primary-nav.tsx)
- [x] 频道侧边栏 (channel-sidebar.tsx)
- [x] 消息主区域 (message-area.tsx)
- [x] 线程面板 (thread-panel.tsx)
- [x] 工作区布局组装 (workspace/layout.tsx)
- [x] 工作区切换器 (workspace-switcher.tsx) (集成在侧边栏中)


## Phase 2: 消息系统 & 交互
- [x] 消息项组件 (message-item.tsx)
- [x] 消息列表 (message-list.tsx)
- [x] 消息编辑器 (message-composer.tsx)
- [x] 消息操作栏 (message-actions.tsx)
- [x] Emoji 反应 (emoji-reaction.tsx)
- [x] 频道头部 (channel-header.tsx) (集成在 message-area.tsx)


## Phase 3: AI-Native 能力集成
- [x] AI 对话面板 (ai-chat-panel.tsx)
- [x] AI Slash Commands (ai-slash-command.tsx)
- [x] @Mention 弹出 (mention-popover.tsx) (集成在 AI 命令中)
- [x] AI 对话 Hook (use-ai-chat.ts)
- [x] AI 思考展示 (ai-reasoning.tsx)


## Phase 4: 高级功能
- [x] 全局搜索 (search-dialog.tsx)
- [x] Huddle 状态栏 (huddle-bar.tsx)
- [x] DMs / Activity / Later 视图页面
- [x] 路由导航集成


## Phase 5: 主题 & 打磨
- [x] Slack 深色主题定制
- [x] 过渡动画
- [x] 键盘快捷键 (Esc 关闭面板, ArrowUp 模拟编辑)
- [x] 构建验证 & 浏览器测试 (pnpm build 通过)

## Phase 6: AI 生产力增强
- [x] AI 线程摘要 (ai-thread-summary.tsx)
- [x] 智能协作资料卡 (user-profile.tsx)
- [x] 智能画布 (AI Canvas / Artifacts)
- [x] 布局动态适配 (Thread/AI/Canvas 互斥与宽度切换)

## Phase 7: Tiptap 高级编辑器集成
- [x] 安装 Tiptap 相关依赖
- [x] 使用 Tiptap 重构 MessageComposer (保留所有 AI 亮点功能)
- [x] 实现富文本格式化工具栏 (Bold, Italic, Code, Link 等)
- [x] 深度定制编辑器配色与 Slack 风格一致
- [x] 集成 @ 提及与 / 快捷命令逻辑

**当前前端实现任务已完成，仓库正在重定位为 Relay Agent Workspace。**


