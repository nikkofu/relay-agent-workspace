"use client"

import { useState, useRef, useEffect } from "react"
import {
  Bold,
  Italic,
  Strikethrough,
  Link,
  List,
  ListOrdered,
  Quote,
  Code,
  FileCode,
  Type,
  Smile,
  Paperclip,
  Send,
  Mic,
  Sparkles,
  Plus,
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip"
import { cn } from "@/lib/utils"
import { AISlashCommand } from "@/components/ai-chat/ai-slash-command"
import { MentionPopover } from "./mention-popover"
import { useUIStore } from "@/stores/ui-store"

interface MessageComposerProps {
  placeholder?: string
  onSend?: (content: string) => void
}

// Static demo version of the composer kept for reference during UI exploration.
export function MessageComposer({ placeholder, onSend }: MessageComposerProps) {
  const [content, setContent] = useState("")
  const [showSlashCommands, setShowSlashCommands] = useState(false)
  const [showMentions, setShowMentions] = useState(false)
  const { openCanvas } = useUIStore()
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  useEffect(() => {
    if (textareaRef.current) {
      textareaRef.current.style.height = "auto"
      textareaRef.current.style.height = `${Math.min(textareaRef.current.scrollHeight, 200)}px`
    }
  }, [content])

  const handleSend = () => {
    if (content.trim()) {
      onSend?.(content)
      setContent("")
      setShowSlashCommands(false)
      setShowMentions(false)
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
    if (e.key === "ArrowUp" && content === "") {
      console.log("Slack feature: Edit last message triggered")
    }
    if (e.key === "Escape") {
      setShowSlashCommands(false)
      setShowMentions(false)
    }
  }

  const handleContentChange = (val: string) => {
    setContent(val)

    if (val === "/") {
      setShowSlashCommands(true)
      setShowMentions(false)
    } else if (val.endsWith("@")) {
      setShowMentions(true)
      setShowSlashCommands(false)
    } else {
      if (!val.includes("/")) setShowSlashCommands(false)
      if (!val.includes("@")) setShowMentions(false)
    }
  }

  const handleSlashCommandSelect = (cmd: string) => {
    if (cmd === "/canvas") {
      openCanvas("new-doc")
      setContent("")
      setShowSlashCommands(false)
      return
    }
    setContent(cmd + " ")
    setShowSlashCommands(false)
    textareaRef.current?.focus()
  }

  const handleMentionSelect = (name: string) => {
    const newContent = content.substring(0, content.lastIndexOf("@")) + "@" + name + " "
    setContent(newContent)
    setShowMentions(false)
    textareaRef.current?.focus()
  }

  return (
    <div className="p-4 pt-0 relative flex flex-col">
      <div className="relative w-full">
        {showSlashCommands && <AISlashCommand onSelect={handleSlashCommandSelect} />}
        {showMentions && <MentionPopover onSelect={handleMentionSelect} />}
      </div>

      <div className="border rounded-lg bg-white dark:bg-[#1a1d21] focus-within:ring-1 focus-within:ring-ring transition-shadow group shrink-0">
        <div className="flex items-center gap-0.5 p-1 border-b bg-muted/30">
          <ToolbarButton icon={Bold} tooltip="Bold (⌘B)" />
          <ToolbarButton icon={Italic} tooltip="Italic (⌘I)" />
          <ToolbarButton icon={Strikethrough} tooltip="Strikethrough" />
          <div className="w-[1px] h-4 bg-border mx-1" />
          <ToolbarButton icon={Link} tooltip="Link" />
          <div className="w-[1px] h-4 bg-border mx-1" />
          <ToolbarButton icon={ListOrdered} tooltip="Ordered list" />
          <ToolbarButton icon={List} tooltip="Bulleted list" />
          <div className="w-[1px] h-4 bg-border mx-1" />
          <ToolbarButton icon={Quote} tooltip="Blockquote" />
          <div className="w-[1px] h-4 bg-border mx-1" />
          <ToolbarButton icon={Code} tooltip="Code" />
          <ToolbarButton icon={FileCode} tooltip="Code block" />
        </div>

        <textarea
          ref={textareaRef}
          value={content}
          onChange={(e) => handleContentChange(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder={placeholder}
          className="w-full resize-none bg-transparent border-none focus:ring-0 px-3 py-2 text-sm min-h-[44px] max-h-[200px] outline-none placeholder:text-muted-foreground/50 block"
        />

        <div className="flex items-center justify-between px-2 pb-2">
          <div className="flex items-center gap-0.5">
            <Tooltip>
              <TooltipTrigger asChild>
                <Button variant="ghost" size="icon" className="h-8 w-8 text-muted-foreground hover:bg-muted">
                  <Plus className="h-4 w-4" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>Shortcuts</TooltipContent>
            </Tooltip>
            <div className="w-[1px] h-4 bg-border mx-1" />
            <ToolbarButton icon={Type} tooltip="Formatting" />
            <ToolbarButton icon={Smile} tooltip="Emoji" />
            <ToolbarButton icon={Paperclip} tooltip="Attach file" />
            <ToolbarButton icon={Mic} tooltip="Video/Audio clip" />
            <div className="w-[1px] h-4 bg-border mx-1" />
            <Tooltip>
              <TooltipTrigger asChild>
                <Button variant="ghost" size="icon" className="h-8 w-8 text-purple-500 hover:bg-purple-500/10">
                  <Sparkles className="h-4 w-4" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>AI Assistant</TooltipContent>
            </Tooltip>
          </div>

          <Button
            size="sm"
            disabled={!content.trim()}
            onClick={handleSend}
            className={cn(
              "h-7 w-7 p-0 transition-colors",
              content.trim() ? "bg-[#007a5a] hover:bg-[#007a5a]/90 text-white" : "bg-muted text-muted-foreground"
            )}
          >
            <Send className="h-4 w-4" />
          </Button>
        </div>
      </div>
      <div className="flex justify-end mt-1 px-1">
        <p className="text-[10px] text-muted-foreground font-medium italic">
          <span className="font-bold not-italic">Return</span> to send · <span className="font-bold not-italic">Shift + Return</span> to add a new line
        </p>
      </div>
    </div>
  )
}

function ToolbarButton({ icon: Icon, tooltip, active }: { icon: React.ElementType; tooltip: string; active?: boolean }) {
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Button
          variant="ghost"
          size="icon"
          className={cn("h-7 w-7 hover:bg-muted", active ? "bg-muted text-foreground" : "text-muted-foreground")}
        >
          <Icon className="h-3.5 w-3.5" />
        </Button>
      </TooltipTrigger>
      <TooltipContent side="top" className="text-[10px] px-2 py-1">
        {tooltip}
      </TooltipContent>
    </Tooltip>
  )
}
