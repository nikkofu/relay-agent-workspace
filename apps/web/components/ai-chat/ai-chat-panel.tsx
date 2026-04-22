"use client"

import { useAIChat } from "@/hooks/use-ai-chat"
import type { AIMessage } from "@/types"
import { AIMessageItem } from "./ai-message"
import { ScrollArea } from "@/components/ui/scroll-area"
import { MessageComposer } from "@/components/message/message-composer"
import { X, Sparkles, Wand2, History, Plus, MessageSquare } from "lucide-react"
import { Button } from "@/components/ui/button"
import { useUIStore } from "@/stores/ui-store"
import { useAIStore } from "@/stores/ai-store"
import { useChannelStore } from "@/stores/channel-store"
import { useMessageStore } from "@/stores/message-store"
import { useUserStore } from "@/stores/user-store"
import { useEffect, useRef, useState } from "react"
import { AISettings } from "./ai-settings"
import { formatDistanceToNow } from "date-fns"
import { cn } from "@/lib/utils"
import { toast } from "sonner"

export function AIChatPanel() {
  const { 
    messages, 
    append, 
    regenerate,
    copyToClipboard,
    submitFeedback,
    startNewChat,
    currentProvider, 
    setProvider,
    currentMode,
    setMode,
    currentModel,
    setModel,
    availableProviders,
    currentConversationId
  } = useAIChat()
  const { conversations, fetchConversations, setCurrentConversation } = useAIStore()
  const { isAIPanelOpen, closeAIPanel, openCanvas } = useUIStore()
  const { currentChannel } = useChannelStore()
  const [showHistory, setShowHistory] = useState(false)
  const scrollRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (isAIPanelOpen) {
      fetchConversations()
    }
  }, [isAIPanelOpen, fetchConversations])

  // Auto-scroll to bottom on new AI messages
  useEffect(() => {
    if (scrollRef.current) {
      const scrollContainer = scrollRef.current.querySelector('[data-radix-scroll-area-viewport]')
      if (scrollContainer) {
        scrollContainer.scrollTop = scrollContainer.scrollHeight
      }
    }
  }, [messages, showHistory])

  const handleSend = (content: string) => {
    // Intercept /canvas to open assistant canvas immediately
    if (content.replace(/<[^>]*>/g, '').trim() === "/canvas") {
      openCanvas('ai-assistant')
    }
    append(content)
  }

  const handleSuggestionClick = (suggestion: string) => {
    if (suggestion === "Summarize this channel") {
      if (!currentChannel) {
        toast.info("Please select a channel to summarize first.")
        return
      }

      // Pull recent channel history from the store and inject into the prompt
      // so the AI has actual context instead of just a channel name.
      const allMessages = useMessageStore.getState().messages
      const users = useUserStore.getState().users
      const channelMessages = allMessages
        .filter(m => m.channelId === currentChannel.id && !m.threadId)
        .slice(-50)

      if (channelMessages.length === 0) {
        handleSend(`Summarize the #${currentChannel.name} channel. No recent messages are available yet.`)
        return
      }

      const userMap = new Map(users.map(u => [u.id, u.name]))
      const history = channelMessages.map(m => {
        const sender = userMap.get(m.senderId) || m.senderId
        const plain = (m.content || "").replace(/<[^>]*>/g, '').replace(/\s+/g, ' ').trim()
        return `- [${sender}] ${plain}`
      }).join('\n')

      const displayContent = `Summarize the #${currentChannel.name} channel`
      const apiPrompt = [
        `Summarize the #${currentChannel.name} channel for a busy teammate.`,
        `Focus on key topics, decisions, action items with owners, blockers, and overall tone.`,
        `Return a concise markdown summary with bold section headings.`,
        ``,
        `Recent messages (${channelMessages.length}):`,
        history,
      ].join('\n')

      append(displayContent, apiPrompt)
      return
    }
    handleSend(suggestion)
  }

  if (!isAIPanelOpen) return null

  return (
    <div className="w-full h-full flex flex-col bg-white dark:bg-[#1a1d21] border-l shadow-2xl relative overflow-hidden animate-in slide-in-from-right duration-300">
      {/* Header */}
      <header className="h-14 px-4 flex items-center justify-between border-b shrink-0 bg-[#3f0e40] dark:bg-[#1a1d21] text-white z-10">
        <div className="flex items-center gap-2.5">
          <div className="w-7 h-7 rounded-lg bg-purple-500 flex items-center justify-center shadow-lg shadow-purple-500/20">
            <Sparkles className="w-4 h-4 text-white" />
          </div>
          <div className="flex flex-col">
            <h3 className="font-bold text-sm leading-tight">AI Assistant</h3>
            <span className="text-[10px] text-purple-200/70 font-medium capitalize">
              {currentProvider === 'openrouter' ? 'Open Router' : currentProvider} • {currentModel || 'No Model'} • {currentMode}
            </span>
          </div>
        </div>
        <div className="flex items-center gap-1">
          <Button 
            variant="ghost" 
            size="icon" 
            className={cn("h-8 w-8 rounded-full transition-colors", showHistory ? "text-purple-400 bg-white/10" : "text-white/70 hover:text-white hover:bg-white/10")}
            onClick={() => setShowHistory(!showHistory)}
          >
            <History className="w-4 h-4" />
          </Button>
          <Button 
            variant="ghost" 
            size="icon" 
            className="h-8 w-8 text-white/70 hover:text-white hover:bg-white/10 rounded-full transition-colors"
            onClick={() => {
              startNewChat()
              setShowHistory(false)
            }}
          >
            <Plus className="w-4 h-4" />
          </Button>
          <AISettings 
            provider={currentProvider}
            setProvider={setProvider}
            mode={currentMode}
            setMode={setMode}
            model={currentModel}
            setModel={setModel}
            availableProviders={availableProviders}
          />
          <Button 
            variant="ghost" 
            size="icon" 
            className="h-8 w-8 text-white/70 hover:text-white hover:bg-white/10 rounded-full transition-colors" 
            onClick={closeAIPanel}
          >
            <X className="w-4 h-4" />
          </Button>
        </div>
      </header>

      {/* Main Area */}
      <div className="flex-1 min-h-0 relative bg-slate-50/30 dark:bg-transparent overflow-hidden">
        {showHistory ? (
          <ScrollArea className="h-full w-full">
            <div className="p-4 flex flex-col gap-2">
              <h4 className="text-xs font-bold text-muted-foreground uppercase tracking-wider mb-2">Previous Chats</h4>
              {conversations.length === 0 ? (
                <div className="py-12 text-center text-muted-foreground text-sm italic">
                  No conversation history yet.
                </div>
              ) : (
                conversations.map((conv) => (
                  <button
                    key={conv.id}
                    onClick={() => {
                      setCurrentConversation(conv.id)
                      setShowHistory(false)
                    }}
                    className={cn(
                      "w-full text-left p-3 rounded-lg border transition-all group flex flex-col gap-1.5",
                      currentConversationId === conv.id 
                        ? "bg-purple-50 dark:bg-purple-900/10 border-purple-200 dark:border-purple-800/50" 
                        : "bg-white dark:bg-transparent border-transparent hover:bg-slate-50 dark:hover:bg-slate-900/30"
                    )}
                  >
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <MessageSquare className={cn("w-3.5 h-3.5", currentConversationId === conv.id ? "text-purple-500" : "text-muted-foreground")} />
                        <span className={cn("text-sm font-bold truncate max-w-[180px]", currentConversationId === conv.id ? "text-purple-700 dark:text-purple-300" : "text-foreground")}>
                          {conv.title || "Untitled Chat"}
                        </span>
                      </div>
                      <span className="text-[10px] text-muted-foreground whitespace-nowrap">
                        {formatDistanceToNow(new Date(conv.updatedAt), { addSuffix: true })}
                      </span>
                    </div>
                    {conv.lastMessage && (
                      <p className="text-xs text-muted-foreground line-clamp-1 italic">
                        {conv.lastMessage}
                      </p>
                    )}
                  </button>
                ))
              )}
            </div>
          </ScrollArea>
        ) : (
          /* Messages Area */
          <ScrollArea ref={scrollRef} className="h-full w-full">
            <div className="flex flex-col min-h-full pb-4">
              {messages.length === 0 ? (
                <div className="flex-1 flex flex-col items-center justify-center p-8 text-center gap-6 mt-12">
                  <div className="relative">
                    <div className="w-20 h-20 rounded-full bg-purple-100 dark:bg-purple-900/20 flex items-center justify-center animate-pulse">
                      <Wand2 className="w-10 h-10 text-purple-600 dark:text-purple-400" />
                    </div>
                    <div className="absolute -top-1 -right-1 w-6 h-6 rounded-full bg-purple-600 flex items-center justify-center border-2 border-white dark:border-[#1a1d21]">
                      <Sparkles className="w-3 h-3 text-white" />
                    </div>
                  </div>
                  <div className="max-w-[240px]">
                    <h4 className="font-bold text-lg mb-2">I&apos;m your AI sidekick</h4>
                    <p className="text-sm text-muted-foreground leading-relaxed">
                      Ask me to summarize channels, help with code, or draft messages. I&apos;m here to make you faster.
                    </p>
                  </div>
                  <div className="grid grid-cols-1 gap-2 w-full mt-4 px-2">
                    <SuggestionButton text="Summarize this channel" onClick={() => handleSuggestionClick("Summarize this channel")} />
                    <SuggestionButton text="Help me draft a professional reply" onClick={() => handleSuggestionClick("Help me draft a professional reply")} />
                    <SuggestionButton text="Explain current engineering trends" onClick={() => handleSuggestionClick("Explain current engineering trends")} />
                  </div>
                </div>
              ) : (
                <div className="flex flex-col gap-0">
                  {messages.map((msg: AIMessage, idx: number) => (
                    <AIMessageItem 
                      key={msg.id} 
                      message={msg} 
                      onCopy={copyToClipboard}
                      onRegenerate={regenerate}
                      onFeedback={(isGood) => submitFeedback(msg.id, isGood)}
                      isLast={idx === messages.length - 1 && msg.role === 'assistant'}
                    />
                  ))}
                </div>
              )}
            </div>
          </ScrollArea>
        )}
      </div>

      {/* Input Area */}
      {!showHistory && (
        <div className="p-4 border-t bg-white dark:bg-[#1a1d21] shrink-0 shadow-[0_-4px_12px_rgba(0,0,0,0.05)] z-10">
          <MessageComposer 
            placeholder="Message AI Assistant..." 
            onSend={handleSend} 
          />
        </div>
      )}
    </div>
  )
}

function SuggestionButton({ text, onClick }: { text: string, onClick: () => void }) {
  return (
    <Button 
      variant="outline" 
      className="justify-start h-auto py-2.5 px-3 text-left text-xs bg-white dark:bg-[#1a1d21] border-purple-200/50 dark:border-purple-800/30 hover:border-purple-50 hover:bg-purple-50 transition-all group"
      onClick={onClick}
    >
      <Sparkles className="w-3 h-3 mr-2 text-purple-500 opacity-50 group-hover:opacity-100" />
      <span className="truncate text-foreground">{text}</span>
    </Button>
  )
}
