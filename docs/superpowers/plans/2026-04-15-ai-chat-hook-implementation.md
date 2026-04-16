# AI Chat Hook (Simulated Streaming) Implementation Plan

> Historical note: this implementation plan was drafted during the project's early `acim-ui` phase. The current product name is `Relay Agent Workspace`.

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement a `useAIChat` React Hook that simulates AI dialogue with reasoning, tool-calling states, and word-by-word streaming output.

**Architecture:** The hook will manage local `messages` state using `useState`. The `append` function will trigger a sequence of simulated async events (reasoning, tool calling, and streaming) to mimic a real AI backend.

**Tech Stack:** React (Hooks), TypeScript, Vitest/Jest (for TDD).

---

### Task 1: Update AIMessage Type

**Files:**
- Modify: `types/index.ts`

- [ ] **Step 1: Update the `AIMessage` interface**

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

- [ ] **Step 2: Commit types update**

```bash
git add types/index.ts
git commit -m "chore: update AIMessage type to include streaming and reasoning fields"
```

---

### Task 2: Create Failing Tests for `useAIChat`

**Files:**
- Create: `hooks/__tests__/use-ai-chat.test.ts`

- [ ] **Step 1: Write the failing tests**

```typescript
import { renderHook, act } from "@testing-library/react"
import { useAIChat } from "../use-ai-chat"
import { vi, describe, it, expect, beforeEach, afterEach } from "vitest"

describe("useAIChat", () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it("should initialize with an empty message list", () => {
    const { result } = renderHook(() => useAIChat())
    expect(result.current.messages).toEqual([])
  })

  it("should add a user message and a streaming assistant message on append", async () => {
    const { result } = renderHook(() => useAIChat())
    
    await act(async () => {
      result.current.append("Hello AI")
    })

    expect(result.current.messages).toHaveLength(2)
    expect(result.current.messages[0].role).toBe("user")
    expect(result.current.messages[0].content).toBe("Hello AI")
    expect(result.current.messages[1].role).toBe("assistant")
    expect(result.current.messages[1].isStreaming).toBe(true)
  })

  it("should simulate reasoning then streaming content", async () => {
    const { result } = renderHook(() => useAIChat())
    
    await act(async () => {
      result.current.append("Tell me a joke")
    })

    // Advance time for reasoning
    await act(async () => {
      vi.advanceTimersByTime(1000)
    })
    expect(result.current.messages[1].reasoning).toBeDefined()

    // Advance time for first word
    await act(async () => {
      vi.advanceTimersByTime(500)
    })
    expect(result.current.messages[1].content).not.toBe("")

    // Advance time to completion
    await act(async () => {
      vi.advanceTimersByTime(5000)
    })
    expect(result.current.messages[1].isStreaming).toBe(false)
  })
})
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `npm test hooks/__tests__/use-ai-chat.test.ts`
Expected: FAIL (module not found or function not defined)

---

### Task 3: Implement `useAIChat` Hook

**Files:**
- Create: `hooks/use-ai-chat.ts`

- [ ] **Step 1: Implement the minimal `useAIChat` hook**

```typescript
import { useState, useCallback } from "react"
import { AIMessage } from "@/types"

export function useAIChat() {
  const [messages, setMessages] = useState<AIMessage[]>([])

  const append = useCallback((content: string) => {
    const userMessage: AIMessage = {
      id: Math.random().toString(36).substring(7),
      role: "user",
      content,
      createdAt: new Date().toISOString(),
    }

    const assistantId = Math.random().toString(36).substring(7)
    const assistantMessage: AIMessage = {
      id: assistantId,
      role: "assistant",
      content: "",
      createdAt: new Date().toISOString(),
      isStreaming: true,
    }

    setMessages((prev) => [...prev, userMessage, assistantMessage])

    // Simulate reasoning
    setTimeout(() => {
      setMessages((prev) =>
        prev.map((m) =>
          m.id === assistantId
            ? { ...m, reasoning: "I am thinking about your request..." }
            : m
        )
      )

      // Simulate streaming
      const response = "This is a simulated AI response to: " + content
      const words = response.split(" ")
      let currentContent = ""
      
      words.forEach((word, index) => {
        setTimeout(() => {
          currentContent += (index === 0 ? "" : " ") + word
          setMessages((prev) =>
            prev.map((m) =>
              m.id === assistantId
                ? {
                    ...m,
                    content: currentContent,
                    isStreaming: index < words.length - 1,
                  }
                : m
            )
          )
        }, (index + 1) * 100)
      })
    , 1000)
  }, [])

  return { messages, append }
}
```

- [ ] **Step 2: Run tests to verify they pass**

Run: `npm test hooks/__tests__/use-ai-chat.test.ts`
Expected: PASS

- [ ] **Step 3: Refine implementation (optional cleanup)**

Ensure types and dates are consistent.

- [ ] **Step 4: Final commit for hook implementation**

```bash
git add hooks/use-ai-chat.ts
git commit -m "feat: implement simulated AI chat hook with streaming and reasoning"
```

---

### Task 4: Final Verification

- [ ] **Step 1: Verify type safety**

Run: `npx tsc --noEmit`

- [ ] **Step 2: Final commit and cleanup**
