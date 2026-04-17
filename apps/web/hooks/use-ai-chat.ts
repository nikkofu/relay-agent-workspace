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
          model: currentModel
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

      if (!reader) return

      while (true) {
        const { done, value } = await reader.read()
        if (done) break

        const chunk = decoder.decode(value, { stream: true })
        const lines = chunk.split("\n")

        for (const line of lines) {
          if (line.startsWith("data: ")) {
            try {
              const data = JSON.parse(line.slice(6))
              
              if (data.event === "start") {
                // Could set model name or reasoning if provided
              } else if (data.event === "chunk") {
                fullContent += data.text || ""
                setMessages(prev => prev.map(m => 
                  m.id === assistantId ? { ...m, content: fullContent } : m
                ))
              } else if (data.event === "done") {
                setMessages(prev => prev.map(m => 
                  m.id === assistantId ? { ...m, isStreaming: false } : m
                ))
              } else if (data.event === "error") {
                toast.error(data.message || "Streaming error")
              }
            } catch (e) {
              // Not a valid JSON or other line, ignore
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
  }, [currentProvider, currentModel])

  return { 
    messages, 
    append, 
    isLoading, 
    currentProvider, 
    setCurrentProvider,
    currentModel,
    setCurrentModel
  }
}
