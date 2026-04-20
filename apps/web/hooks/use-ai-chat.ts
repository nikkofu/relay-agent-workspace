"use client"

import { useState, useCallback, useEffect, useRef } from "react"
import { AIMessage } from "@/types"
import { API_BASE_URL } from "@/lib/constants"
import { toast } from "sonner"
import { useUserStore } from "@/stores/user-store"
import { useAIStore } from "@/stores/ai-store"
import { useArtifactStore } from "@/stores/artifact-store"
import { useUIStore } from "@/stores/ui-store"

interface AIProviderConfig {
  id: string
  models: string[]
}

interface AIConfigResponse {
  default_provider: string
  providers: AIProviderConfig[]
}

export function useAIChat() {
  const { currentUser } = useUserStore()
  const { 
    messages, currentConversationId, isLoading,
    addMessage, updateMessage, setLoading, fetchConversations, setCurrentConversation
  } = useAIStore()
  const { generateAIArtifact } = useArtifactStore()
  const { openCanvas } = useUIStore()
  
  // State initialized from user preferences
  const [currentProvider, setCurrentProvider] = useState<string>("gemini")
  const [currentModel, setCurrentModel] = useState<string>("")
  const [currentMode, setCurrentMode] = useState<"fast" | "planning">("fast")
  const [availableProviders, setAvailableProviders] = useState<AIProviderConfig[]>([])

  const configFetched = useRef(false)

  // Hydrate from currentUser
  useEffect(() => {
    if (currentUser) {
      if (currentUser.aiProvider) setCurrentProvider(currentUser.aiProvider)
      if (currentUser.aiModel) setCurrentModel(currentUser.aiModel)
      if (currentUser.aiMode) setCurrentMode(currentUser.aiMode)
    }
  }, [currentUser])

  // Fetch dynamic config
  const fetchConfig = useCallback(async () => {
    if (configFetched.current) return
    configFetched.current = true

    try {
      const response = await fetch(`${API_BASE_URL}/ai/config`)
      const data: AIConfigResponse = await response.json()
      setAvailableProviders(data.providers)
      
      // Set default provider if none set or if current not in available
      let provider = currentProvider
      const isCurrentValid = data.providers.some(p => p.id === provider)
      
      if ((!provider || !isCurrentValid) && data.default_provider) {
        provider = data.default_provider
        setCurrentProvider(provider)
      }

      // If no model set or invalid for provider, use first one from provider config
      if (provider) {
        const pConfig = data.providers.find(p => p.id === provider)
        if (pConfig && pConfig.models.length > 0) {
          if (!currentModel || !pConfig.models.includes(currentModel)) {
            setCurrentModel(pConfig.models[0])
          }
        }
      }
    } catch (error) {
      configFetched.current = false // Allow retry on error
      console.error("Failed to fetch AI config:", error)
    }
  }, [currentProvider, currentModel])

  useEffect(() => {
    fetchConfig()
  }, [fetchConfig])

  // Persist settings
  const saveSettings = useCallback(async (updates: { provider?: string, model?: string, mode?: string }) => {
    try {
      await fetch(`${API_BASE_URL}/me/settings`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(updates)
      })
    } catch (error) {
      console.error("Failed to save AI settings:", error)
    }
  }, [])

  const updateProvider = (provider: string) => {
    setCurrentProvider(provider)
    saveSettings({ provider })
    
    // Automatically pick the first model for the new provider
    const config = availableProviders.find(p => p.id === provider)
    if (config && config.models.length > 0) {
      const model = config.models[0]
      setCurrentModel(model)
      saveSettings({ provider, model })
    }
  }

  const updateModel = (model: string) => {
    setCurrentModel(model)
    saveSettings({ model })
  }

  const updateMode = (mode: "fast" | "planning") => {
    setCurrentMode(mode)
    saveSettings({ mode })
  }

  const append = useCallback(async (content: string) => {
    const plainText = content.replace(/<[^>]*>/g, '').trim()
    const slashCommand = plainText.startsWith('/') ? plainText.slice(1).split(/\s+/)[0] : undefined

    const userMsg: AIMessage = {
      id: Date.now().toString(),
      role: "user",
      content,
      createdAt: new Date().toISOString()
    }
    
    addMessage(userMsg)
    setLoading(true)

    const assistantId = (Date.now() + 1).toString()
    const initialAssistantMsg: AIMessage = {
      id: assistantId,
      role: "assistant",
      content: "",
      isStreaming: true,
      createdAt: new Date().toISOString()
    }
    
    addMessage(initialAssistantMsg)

    try {
      const response = await fetch(`${API_BASE_URL}/ai/execute`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ 
          prompt: content, 
          channel_id: "ai-chat", 
          conversation_id: currentConversationId,
          command: slashCommand,
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
      let reasoningContent = ""
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
              
              if (currentEvent === "conversation") {
                // If it's a new conversation, update the store and fetch list
                if (data.id && data.id !== currentConversationId) {
                  useAIStore.setState({ currentConversationId: data.id })
                  fetchConversations()
                }
              } else if (currentEvent === "canvas.generate") {
                if (data.prompt) {
                  generateAIArtifact(data.prompt, "ch-1").then((art: any) => {
                    if (art && art.id) {
                      openCanvas(art.id)
                    }
                  })
                }
              } else if (currentEvent === "start") {
                if (data.model) setCurrentModel(data.model)
                if (data.conversation_id && data.conversation_id !== currentConversationId) {
                  useAIStore.setState({ currentConversationId: data.conversation_id })
                  fetchConversations()
                }
              } else if (currentEvent === "chunk") {
                fullContent += data.text || ""
                updateMessage(assistantId, { content: fullContent })
              } else if (currentEvent === "reasoning") {
                reasoningContent += data.text || ""
                updateMessage(assistantId, { reasoning: reasoningContent })
              } else if (currentEvent === "done") {
                updateMessage(assistantId, { isStreaming: false })
                fetchConversations() // Refresh titles
              } else if (currentEvent === "error") {
                toast.error(data.message || "Streaming error")
              }
            } catch {
              // Ignore invalid JSON or fragments
            }
          }
        }
      }
    } catch (error) {
      console.error("AI Chat failed:", error)
      updateMessage(assistantId, { content: "Error communicating with AI service.", isStreaming: false })
    } finally {
      setLoading(false)
    }
  }, [currentProvider, currentModel, currentMode, currentConversationId, addMessage, setLoading, updateMessage, fetchConversations])

  const regenerate = useCallback(async () => {
    const lastUserMsg = [...messages].reverse().find(m => m.role === 'user')
    if (lastUserMsg) {
      // Remove last assistant message if it exists
      const lastMsg = messages[messages.length - 1]
      if (lastMsg.role === 'assistant') {
        useAIStore.setState((state) => ({ messages: state.messages.slice(0, -1) }))
      }
      append(lastUserMsg.content)
    }
  }, [messages, append])

  const copyToClipboard = useCallback((text: string) => {
    navigator.clipboard.writeText(text)
    toast.success("Copied to clipboard")
  }, [])

  const submitFeedback = useCallback(async (messageId: string, isGood: boolean) => {
    try {
      await fetch(`${API_BASE_URL}/ai/feedback`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ message_id: messageId, is_good: isGood })
      })
      toast.success(isGood ? "Thanks for your feedback!" : "Sorry to hear that. We'll improve.")
    } catch (error) {
      console.error("Failed to submit feedback:", error)
      toast.success(isGood ? "Thanks for your feedback!" : "Sorry to hear that. We'll improve.")
    }
  }, [])

  const startNewChat = useCallback(() => {
    setCurrentConversation(null)
  }, [setCurrentConversation])

  return { 
    messages, 
    append, 
    regenerate,
    copyToClipboard,
    submitFeedback,
    startNewChat,
    isLoading, 
    currentProvider, 
    setProvider: updateProvider,
    currentModel,
    setModel: updateModel,
    currentMode,
    setMode: updateMode,
    availableProviders,
    currentConversationId
  }
}
