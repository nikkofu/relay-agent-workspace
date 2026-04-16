# AI 对话 Hook 设计文档 (2026-04-15)

> 历史说明：本设计文档形成于项目早期，当时仓库仍使用 `acim-ui / ACIM-UI` 作为代号。当前产品名称为 `Relay Agent Workspace`。

## 目标
实现 `useAIChat` Hook，用于在 `Relay Agent Workspace` 的早期前端中模拟 AI 对话逻辑。该 Hook 需要支持流式输出模拟、思考（reasoning）阶段，以及对话历史管理。

## 核心接口变更

### `types/index.ts`
更新 `AIMessage` 接口，包含流式输出所需的元数据：
```typescript
export interface AIMessage {
  id: string
  role: "user" | "assistant" | "system"
  content: string
  createdAt: string
  isStreaming?: boolean
  reasoning?: string
  tools?: {
    name: string
    args: any
    state: "calling" | "result"
    result?: any
  }[]
  sources?: { title: string; url: string }[]
}
```

## `useAIChat` Hook 设计

### 状态管理
- `messages: AIMessage[]`: 对话消息数组。
- `input: string`: 当前输入框内容（可选，取决于是否需要集中管理）。
- `isLoading: boolean`: 是否正在生成响应。

### 核心方法
- `append(content: string)`: 
  1. 添加用户消息到 `messages`。
  2. 初始化一条带有 `isStreaming: true` 的助手消息。
  3. 模拟“思考”阶段：更新消息的 `reasoning` 字段。
  4. 模拟“流式”输出阶段：使用 `setTimeout` 逐词更新消息的 `content` 字段。
  5. 完成后设置 `isStreaming: false`。

### 模拟逻辑
- 思考耗时：1-2 秒。
- 流式速度：每 50-100 毫秒一个词。
- 随机响应模版：预设一些简单的 AI 回复。

## 测试策略 (TDD)
- 编写 `hooks/__tests__/use-ai-chat.test.ts`。
- 验证 `append` 能正确增加消息。
- 验证助手消息的状态流转：思考 -> 正在流式输出 -> 完成。
- 使用 `jest.useFakeTimers()` 处理时间相关的模拟。

## 关联组件
- `AI Chat` 侧边栏/主界面将使用该 Hook。
- `MessageItem` 需要根据 `isStreaming` 和 `reasoning` 显示不同的 UI 状态（如 Skeleton 或正在输入的指示器）。
