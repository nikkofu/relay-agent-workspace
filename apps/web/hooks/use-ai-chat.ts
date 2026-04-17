"use client"

import { useState, useCallback } from "react"
import { AIMessage } from "@/types"
import { API_BASE_URL } from "@/lib/constants"
import { toast } from "sonner"

export function useAIChat() {
  const [messages, setMessages] = useState<AIMessage[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [currentProvider, setCurrentProvider] = useState<string>("gemini")
  const [currentModel, setCurrentModel] = useState<string>("")
  const [currentMode, setCurrentMode] = useState<"fast" | "planning">("fast")
  const [availableProviders, setAvailableProviders] = useState<string[]>(["gemini", "openrouter"])

  const append = useCallback(async (content: string) => {
    const userMsg: AIMessage = {
      id: Date.now().toString(),
      role: "user",
      content,
      createdAt: new Date().toISOString()
    }
    
    setMessages(prev => [...prev, userMsg])
    setIsLoading(true)

    const assistantId = (Date.now() + 1).toString()
    const initialAssistantMsg: AIMessage = {
      id: assistantId,
      role: "assistant",
      content: "",
      isStreaming: true,
      createdAt: new Date().toISOString()
    }
    
    setMessages(prev => [...prev, initialAssistantMsg])

    try {
      const response = await fetch(`${API_BASE_URL}/ai/execute`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ 
          prompt: content, 
          channel_id: "ai-chat", // Default or specific channel
          provider: currentProvider,
          model: currentModel,
          mode: currentMode
        })
      })

      if (!response.ok) {
        if (response.status === 429) {
          toast.error("Rate limit reached. Please try again later.")
        } else {
          toast.error("AI service error. Please try again.")
        }
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      const reader = response.body?.getReader()
      const decoder = new TextDecoder()
      let fullContent = ""
      let currentEvent = ""

      if (!reader) return

      while (true) {
        const { done, value } = await reader.read()
        if (done) break

        const chunk = decoder.decode(value, { stream: true })
        const lines = chunk.split("\n")

        for (const line of lines) {
          const trimmed = line.trim()
          if (trimmed.startsWith("event: ")) {
            currentEvent = trimmed.slice(7).trim()
          } else if (trimmed.startsWith("data: ")) {
            try {
              const data = JSON.parse(trimmed.slice(6))
              
              if (currentEvent === "chunk") {
                fullContent += data.text || ""
                setMessages(prev => prev.map(m => 
                  m.id === assistantId ? { ...m, content: fullContent } : m
                ))
              } else if (currentEvent === "done") {
                setMessages(prev => prev.map(m => 
                  m.id === assistantId ? { ...m, isStreaming: false } : m
                ))
              } else if (currentEvent === "error") {
                toast.error(data.message || "Streaming error")
              }
            } catch (e) {
              // Ignore invalid JSON or fragments
            }
          }
        }
      }
    } catch (error) {
      console.error("AI Chat failed:", error)
      setMessages(prev => prev.map(m => 
        m.id === assistantId ? { ...m, content: "Error communicating with AI service.", isStreaming: false } : m
      ))
    } finally {
      setIsLoading(false)
    }
  }, [currentProvider, currentModel, currentMode])

  return { 
    messages, 
    append, 
    isLoading, 
    currentProvider, 
    setCurrentProvider,
    currentModel,
    setCurrentModel,
    currentMode,
    setCurrentMode,
    availableProviders,
    setAvailableProviders
  }
}
