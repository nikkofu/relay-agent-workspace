# Phase 3: AI-Native 能力集成实施计划

> 历史说明：本计划编写于项目早期原型阶段，当时整体仓库仍以 Slack 风格协作 UI 为切入点，当前产品名称已更新为 `Relay Agent Workspace`。

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在早期 Slack 风格协作原型中深度集成 AI 能力，包括模拟流式对话、思考过程展示、AI Slash 命令以及 @AI 助手交互。

**Architecture:** 
- 使用 `use-ai-chat.ts` Hook 模拟 Vercel AI SDK 的流式行为。
- 手动实现 `AI Message` 组件组，支持多种 AI 内容块（文本、思考、工具、引用）。
- 扩展消息编辑器以支持 `@AI` 唤醒和 `/ai` 系列命令。
- 在右侧 Resizable 面板中集成完整的 AI 对话窗口。

**Tech Stack:** React 19, Next.js 15, Zustand, Lucide React, TailwindCSS v4.

---

### Task 1: AI 对话 Hook (模拟流式响应)

**Files:**
- Create: `hooks/use-ai-chat.ts`
- Modify: `types/index.ts`

- [ ] **Step 1: 完善 AI 消息类型定义**
在 `types/index.ts` 中添加更多 AI 相关的状态类型。

```typescript
// types/index.ts
export interface AIMessage {
  id: string
  role: "user" | "assistant" | "system"
  content: string
  createdAt: string
  isStreaming?: boolean
  reasoning?: string // 思考过程
  tools?: {
    name: string
    args: any
    state: "calling" | "result"
    result?: any
  }[]
  sources?: { title: string; url: string }[]
}
```

- [ ] **Step 2: 实现模拟流式的 Hook**
创建一个 Hook，能够模拟逐字输出的效果，并管理对话状态。

```typescript
// hooks/use-ai-chat.ts
import { useState, useCallback } from "react"
import { AIMessage } from "@/types"

export function useAIChat() {
  const [messages, setMessages] = useState<AIMessage[]>([])
  const [isLoading, setIsLoading] = useState(false)

  const append = useCallback(async (content: string) => {
    const userMsg: AIMessage = {
      id: Date.now().toString(),
      role: "user",
      content,
      createdAt: new Date().toISOString()
    }
    setMessages(prev => [...prev, userMsg])
    setIsLoading(true)

    // 模拟思考
    const assistantId = (Date.now() + 1).toString()
    const initialAssistantMsg: AIMessage = {
      id: assistantId,
      role: "assistant",
      content: "",
      reasoning: "Thinking about your request...",
      isStreaming: true,
      createdAt: new Date().toISOString()
    }
    setMessages(prev => [...prev, initialAssistantMsg])

    await new Promise(resolve => setTimeout(resolve, 1000))

    // 模拟流式回复
    const responseText = "I can help you with that! As an AI-Native Slack assistant, I can summarize channels, generate code, or help you coordinate with your team."
    let currentText = ""
    const words = responseText.split(" ")

    for (const word of words) {
      currentText += word + " "
      setMessages(prev => prev.map(m => 
        m.id === assistantId ? { ...m, content: currentText, reasoning: undefined } : m
      ))
      await new Promise(resolve => setTimeout(resolve, 50))
    }

    setMessages(prev => prev.map(m => 
      m.id === assistantId ? { ...m, isStreaming: false } : m
    ))
    setIsLoading(false)
  }, [])

  return { messages, append, isLoading }
}
```

- [ ] **Step 3: 提交**
```bash
git add types/index.ts hooks/use-ai-chat.ts
git commit -m "feat: add simulated AI chat hook with streaming support"
```

---

### Task 2: AI 消息原子组件 (手动实现)

**Files:**
- Create: `components/ai-chat/ai-message.tsx`
- Create: `components/ai-chat/ai-reasoning.tsx`

- [ ] **Step 1: 实现思考过程组件 (Reasoning)**
模仿 shadcn.io/ai 的风格。

```tsx
// components/ai-chat/ai-reasoning.tsx
import { Brain, ChevronDown } from "lucide-react"
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible"

export function AIReasoning({ content }: { content: string }) {
  return (
    <Collapsible className="my-2">
      <CollapsibleTrigger className="flex items-center gap-2 text-[11px] font-medium text-muted-foreground hover:text-foreground transition-colors group">
        <Brain className="w-3 h-3 text-purple-400" />
        <span>Thought process</span>
        <ChevronDown className="w-3 h-3 group-data-[state=open]:rotate-180 transition-transform" />
      </CollapsibleTrigger>
      <CollapsibleContent className="mt-2 pl-5 border-l-2 border-muted text-xs text-muted-foreground italic leading-relaxed">
        {content}
      </CollapsibleContent>
    </Collapsible>
  )
}
```

- [ ] **Step 2: 实现 AI 消息气泡组件**

```tsx
// components/ai-chat/ai-message.tsx
import { AIMessage } from "@/types"
import { AIReasoning } from "./ai-reasoning"
import { Sparkles, User } from "lucide-react"
import { cn } from "@/lib/utils"

export function AIMessageItem({ message }: { message: AIMessage }) {
  const isAI = message.role === "assistant"

  return (
    <div className={cn("flex gap-3 px-4 py-3 group", isAI && "bg-purple-500/5")}>
      <div className={cn(
        "w-8 h-8 rounded flex items-center justify-center shrink-0",
        isAI ? "bg-purple-600 text-white" : "bg-muted text-muted-foreground"
      )}>
        {isAI ? <Sparkles className="w-4 h-4" /> : <User className="w-4 h-4" />}
      </div>
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2 mb-1">
          <span className="font-bold text-sm">{isAI ? "AI Assistant" : "You"}</span>
          <span className="text-[10px] text-muted-foreground">
            {new Date(message.createdAt).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
          </span>
        </div>
        {message.reasoning && <AIReasoning content={message.reasoning} />}
        <div className="text-sm leading-relaxed whitespace-pre-wrap">
          {message.content}
          {message.isStreaming && <span className="inline-block w-1.5 h-4 bg-purple-500 ml-1 animate-pulse align-middle" />}
        </div>
      </div>
    </div>
  )
}
```

- [ ] **Step 3: 提交**
```bash
git add components/ai-chat/ai-reasoning.tsx components/ai-chat/ai-message.tsx
git commit -m "feat: implement AI message components (reasoning and bubble)"
```

---

### Task 3: AI 对话面板 (AI Chat Panel)

**Files:**
- Create: `components/ai-chat/ai-chat-panel.tsx`
- Modify: `components/layout/primary-nav.tsx`
- Modify: `stores/ui-store.ts`

- [ ] **Step 1: 扩展 UI Store 支持 AI 面板**

```typescript
// stores/ui-store.ts
interface UIState {
  // ... existing
  isAIPanelOpen: boolean
  toggleAIPanel: () => void
  openAIPanel: () => void
  closeAIPanel: () => void
}

export const useUIStore = create<UIState>((set) => ({
  // ... existing
  isAIPanelOpen: false,
  toggleAIPanel: () => set((state) => ({ isAIPanelOpen: !state.isAIPanelOpen, isThreadOpen: false })),
  openAIPanel: () => set({ isAIPanelOpen: true, isThreadOpen: false }),
  closeAIPanel: () => set({ isAIPanelOpen: false }),
}))
```

- [ ] **Step 2: 实现 AI 对话面板组件**

```tsx
// components/ai-chat/ai-chat-panel.tsx
"use client"

import { useAIChat } from "@/hooks/use-ai-chat"
import { AIMessageItem } from "./ai-message"
import { ScrollArea } from "@/components/ui/scroll-area"
import { MessageComposer } from "@/components/message/message-composer"
import { X, Sparkles } from "lucide-react"
import { Button } from "@/components/ui/button"
import { useUIStore } from "@/stores/ui-store"

export function AIChatPanel() {
  const { messages, append, isLoading } = useAIChat()
  const { isAIPanelOpen, closeAIPanel } = useUIStore()

  if (!isAIPanelOpen) return null

  return (
    <div className="w-full h-full flex flex-col bg-white dark:bg-[#1a1d21] border-l shadow-2xl">
      <header className="h-14 px-4 flex items-center justify-between border-b shrink-0 bg-purple-600 text-white">
        <div className="flex items-center gap-2">
          <Sparkles className="w-4 h-4" />
          <h3 className="font-bold">AI Assistant</h3>
        </div>
        <Button variant="ghost" size="icon" className="h-8 w-8 text-white hover:bg-white/20" onClick={closeAIPanel}>
          <X className="w-4 h-4" />
        </Button>
      </header>

      <ScrollArea className="flex-1">
        <div className="flex flex-col min-h-full">
          {messages.length === 0 ? (
            <div className="flex-1 flex flex-col items-center justify-center p-8 text-center gap-4">
              <div className="w-12 h-12 rounded-full bg-purple-100 dark:bg-purple-900/30 flex items-center justify-center">
                <Sparkles className="w-6 h-6 text-purple-600" />
              </div>
              <div>
                <h4 className="font-bold">How can I help you today?</h4>
                <p className="text-sm text-muted-foreground mt-1">Ask me to summarize a channel, draft a message, or explain a concept.</p>
              </div>
            </div>
          ) : (
            messages.map(msg => <AIMessageItem key={msg.id} message={msg} />)
          )}
        </div>
      </ScrollArea>

      <div className="p-4 border-t">
        <MessageComposer 
          placeholder="Ask AI..." 
          onSend={append} 
        />
      </div>
    </div>
  )
}
```

- [ ] **Step 3: 集成到主布局和导航**
修改 `app/workspace/layout.tsx` 以包含 `AIChatPanel`。
修改 `components/layout/primary-nav.tsx` 的 AI 按钮触发 `toggleAIPanel`。

- [ ] **Step 4: 提交**
```bash
git add components/ai-chat/ai-chat-panel.tsx stores/ui-store.ts components/layout/primary-nav.tsx app/workspace/layout.tsx
git commit -m "feat: integrate AI Chat Panel into the workspace layout"
```

---

### Task 4: AI 输入增强 (Slash Commands & @Mention)

**Files:**
- Create: `components/ai-chat/ai-slash-command.tsx`
- Modify: `components/message/message-composer.tsx`

- [ ] **Step 1: 实现 AI Slash 命令弹出组件**

```tsx
// components/ai-chat/ai-slash-command.tsx
import { Command, CommandGroup, CommandItem, CommandList } from "@/components/ui/command"
import { Sparkles, FileText, Languages, Code } from "lucide-react"

export function AISlashCommand({ onSelect }: { onSelect: (cmd: string) => void }) {
  const commands = [
    { icon: Sparkles, label: "Ask AI", value: "/ai", desc: "Start a conversation" },
    { icon: FileText, label: "Summarize", value: "/summarize", desc: "Summarize this channel" },
    { icon: Languages, label: "Translate", value: "/translate", desc: "Translate last message" },
    { icon: Code, label: "Explain Code", value: "/explain", desc: "Explain code snippets" },
  ]

  return (
    <Command className="border shadow-md rounded-lg max-w-[300px]">
      <CommandList>
        <CommandGroup heading="AI Commands">
          {commands.map(cmd => (
            <CommandItem key={cmd.value} onSelect={() => onSelect(cmd.value)} className="flex items-center gap-2 p-2 cursor-pointer">
              <cmd.icon className="w-4 h-4 text-purple-500" />
              <div className="flex flex-col">
                <span className="text-sm font-medium">{cmd.label}</span>
                <span className="text-[10px] text-muted-foreground">{cmd.desc}</span>
              </div>
              <span className="ml-auto text-xs opacity-50">{cmd.value}</span>
            </CommandItem>
          ))}
        </CommandGroup>
      </CommandList>
    </Command>
  )
}
```

- [ ] **Step 2: 在编辑器中集成触发逻辑**
修改 `message-composer.tsx`，当输入 `/` 时弹出命令列表（模拟）。

- [ ] **Step 3: 提交**
```bash
git add components/ai-chat/ai-slash-command.tsx components/message/message-composer.tsx
git commit -m "feat: add AI slash commands integration to message composer"
```

---

### Task 5: 验证 Phase 3

- [ ] **Step 1: 验证流程**
1. 点击左侧窄栏的 ✨ 图标，右侧滑出 AI 面板。
2. 在 AI 面板输入内容，验证“Thinking”状态和逐字流式回复效果。
3. 验证 AI 消息中的“Thought Process”可折叠。
4. 在主频道输入框输入 `/`，弹出 AI 命令列表。
5. 提交所有更改并推送到仓库。

- [ ] **Step 2: 提交最终 Phase 3 更改**
```bash
git push origin main
```
