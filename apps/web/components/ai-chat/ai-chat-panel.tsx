"use client"

import { useAIChat } from "@/hooks/use-ai-chat"
import { AIMessageItem } from "./ai-message"
import { ScrollArea } from "@/components/ui/scroll-area"
import { MessageComposer } from "@/components/message/message-composer"
import { X, Sparkles, Wand2 } from "lucide-react"
import { Button } from "@/components/ui/button"
import { useUIStore } from "@/stores/ui-store"
import { useEffect, useRef } from "react"
import { AISettings } from "./ai-settings"

export function AIChatPanel() {
  const { 
    messages, 
    append, 
    currentProvider, 
    setCurrentProvider,
    currentMode,
    setCurrentMode,
    currentModel,
    setCurrentModel,
    availableProviders
  } = useAIChat()
  const { isAIPanelOpen, closeAIPanel } = useUIStore()
  const scrollRef = useRef<HTMLDivElement>(null)

  // Auto-scroll to bottom on new AI messages
  useEffect(() => {
    if (scrollRef.current) {
      const scrollContainer = scrollRef.current.querySelector('[data-radix-scroll-area-viewport]')
      if (scrollContainer) {
        scrollContainer.scrollTop = scrollContainer.scrollHeight
      }
    }
  }, [messages])

  if (!isAIPanelOpen) return null

  return (
    <div className="w-full h-full flex flex-col bg-white dark:bg-[#1a1d21] border-l shadow-2xl relative overflow-hidden animate-in slide-in-from-right duration-300">
      {/* Header */}
      <header className="h-14 px-4 flex items-center justify-between border-b shrink-0 bg-[#3f0e40] dark:bg-[#1a1d21] text-white">
        <div className="flex items-center gap-2.5">
          <div className="w-7 h-7 rounded-lg bg-purple-500 flex items-center justify-center shadow-lg shadow-purple-500/20">
            <Sparkles className="w-4 h-4 text-white" />
          </div>
          <div className="flex flex-col">
            <h3 className="font-bold text-sm leading-tight">AI Assistant</h3>
            <span className="text-[10px] text-purple-200/70 font-medium capitalize">
              {currentProvider} • {currentMode}
            </span>
          </div>
        </div>
        <div className="flex items-center gap-1">
          <AISettings 
            provider={currentProvider}
            setProvider={setCurrentProvider}
            mode={currentMode}
            setMode={setCurrentMode}
            model={currentModel}
            setModel={setCurrentModel}
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

      {/* Messages Area - Ensure this fills space and scrolls */}
      <div className="flex-1 min-h-0 relative">
        <ScrollArea ref={scrollRef} className="h-full bg-slate-50/30 dark:bg-transparent">
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
                <div className="grid grid-cols-1 gap-2 w-full mt-4">
                  <SuggestionButton text="Summarize this channel" onClick={() => append("Summarize this channel")} />
                  <SuggestionButton text="Help me draft a professional reply" onClick={() => append("Help me draft a professional reply")} />
                  <SuggestionButton text="Explain current engineering trends" onClick={() => append("Explain current engineering trends")} />
                </div>
              </div>
            ) : (
              messages.map(msg => <AIMessageItem key={msg.id} message={msg} />)
            )}
          </div>
        </ScrollArea>
      </div>

      {/* Input Area - Stay at bottom */}
      <div className="p-4 border-t bg-white dark:bg-[#1a1d21] shrink-0 shadow-[0_-4px_12px_rgba(0,0,0,0.02)]">
        <MessageComposer 
          placeholder="Message AI Assistant..." 
          onSend={append} 
        />
      </div>
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
      <span className="truncate">{text}</span>
    </Button>
  )
}
