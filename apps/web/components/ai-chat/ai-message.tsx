"use client"

import { AIMessage } from "@/types"
import { AIReasoning } from "./ai-reasoning"
import { Sparkles, User, RefreshCcw, ThumbsUp, ThumbsDown, Copy } from "lucide-react"
import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"

interface AIMessageItemProps {
  message: AIMessage
  onCopy: (text: string) => void
  onRegenerate: () => void
  onFeedback: (isGood: boolean) => void
  isLast: boolean
}

export function AIMessageItem({ 
  message, 
  onCopy, 
  onRegenerate, 
  onFeedback,
  isLast 
}: AIMessageItemProps) {
  const isAI = message.role === "assistant"

  return (
    <div className={cn(
      "flex gap-3 px-4 py-4 group transition-colors", 
      isAI ? "bg-purple-500/[0.03] dark:bg-purple-500/[0.05]" : "bg-transparent"
    )}>
      <div className={cn(
        "w-8 h-8 rounded flex items-center justify-center shrink-0 border shadow-sm",
        isAI ? "bg-purple-600 text-white border-purple-500" : "bg-muted text-muted-foreground border-border"
      )}>
        {isAI ? <Sparkles className="w-4 h-4" /> : <User className="w-4 h-4" />}
      </div>
      
      <div className="flex-1 min-w-0 flex flex-col gap-1">
        <div className="flex items-center gap-2">
          <span className="font-bold text-sm tracking-tight">{isAI ? "AI Assistant" : "You"}</span>
          <span className="text-[10px] text-muted-foreground font-medium opacity-70">
            {new Date(message.createdAt).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
          </span>
        </div>

        {/* Reasoning Section */}
        {message.reasoning && <AIReasoning content={message.reasoning} />}

        {/* Content Section */}
        <div className={cn(
          "text-sm leading-relaxed whitespace-pre-wrap break-words",
          isAI ? "text-foreground font-normal" : "text-foreground/90 font-normal"
        )}>
          {message.content}
          {message.isStreaming && (
            <span className="inline-block w-1 h-4 bg-purple-500 ml-1 animate-pulse rounded-full align-middle" />
          )}
        </div>

        {/* Actions for AI Response */}
        {isAI && !message.isStreaming && message.content && (
          <div className="flex items-center gap-1 mt-3 opacity-0 group-hover:opacity-100 transition-opacity">
            <ActionButton icon={Copy} tooltip="Copy" onClick={() => onCopy(message.content)} />
            {isLast && <ActionButton icon={RefreshCcw} tooltip="Regenerate" onClick={onRegenerate} />}
            <div className="w-[1px] h-3 bg-border mx-1" />
            <ActionButton icon={ThumbsUp} tooltip="Good response" onClick={() => onFeedback(true)} />
            <ActionButton icon={ThumbsDown} tooltip="Bad response" onClick={() => onFeedback(false)} />
          </div>
        )}
      </div>
    </div>
  )
}

function ActionButton({ icon: Icon, tooltip, onClick }: { icon: React.ElementType, tooltip: string, onClick?: () => void }) {
  return (
    <Button 
      variant="ghost" 
      size="icon" 
      className="h-7 w-7 text-muted-foreground hover:text-foreground hover:bg-muted"
      title={tooltip}
      onClick={onClick}
    >
      <Icon className="w-3.5 h-3.5" />
    </Button>
  )
}
