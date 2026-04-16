"use client"

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
      reasoning: "Analyzing the request and searching for relevant information...",
      isStreaming: true,
      createdAt: new Date().toISOString()
    }
    
    setMessages(prev => [...prev, initialAssistantMsg])

    // 等待 1.2 秒模拟思考
    await new Promise(resolve => setTimeout(resolve, 1200))

    // 模拟流式回复
    const responseText = "I can help you with that! As an AI-Native Slack assistant, I'm designed to supercharge your collaboration. I can summarize channel history, help you draft professional messages, or even coordinate tasks with your teammates."
    
    let currentText = ""
    const words = responseText.split(" ")

    for (const word of words) {
      currentText += word + " "
      setMessages(prev => prev.map(m => 
        m.id === assistantId ? { 
          ...m, 
          content: currentText, 
          reasoning: undefined // 思考完成后移除思考文本
        } : m
      ))
      // 模拟逐词输出速度
      await new Promise(resolve => setTimeout(resolve, 60))
    }

    // 完成流式
    setMessages(prev => prev.map(m => 
      m.id === assistantId ? { ...m, isStreaming: false } : m
    ))
    setIsLoading(false)
  }, [])

  return { messages, append, isLoading }
}
