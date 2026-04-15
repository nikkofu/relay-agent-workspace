# ACIM-UI 任务清单

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
- [ ] 消息项组件 (message-item.tsx)
- [ ] 消息列表 (message-list.tsx)
- [ ] 消息编辑器 (message-composer.tsx)
- [ ] 消息操作栏 (message-actions.tsx)
- [ ] Emoji 反应 (emoji-reaction.tsx)
- [ ] 频道头部 (channel-header.tsx)

## Phase 3: AI-Native 能力集成
- [ ] AI 对话面板 (ai-chat-panel.tsx)
- [ ] AI Slash Commands (ai-slash-command.tsx)
- [ ] @Mention 弹出 (mention-popover.tsx)
- [ ] AI 对话 Hook (use-ai-chat.ts)
- [ ] AI 线程摘要 (ai-thread-summary.tsx)

## Phase 4: 高级功能
- [ ] 全局搜索 (search-dialog.tsx)
- [ ] Huddle 状态栏 (huddle-bar.tsx)
- [ ] DMs / Activity / Later 视图页面

## Phase 5: 主题 & 打磨
- [ ] Slack 深色主题定制
- [ ] 过渡动画
- [ ] 键盘快捷键
- [ ] 构建验证 & 浏览器测试
