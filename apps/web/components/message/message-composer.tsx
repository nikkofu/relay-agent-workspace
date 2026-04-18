"use client"

import { useState, useRef } from "react"
import { 
  Bold, Italic, Strikethrough, Link as LinkIcon, List, ListOrdered, 
  Quote, Code, FileCode, Type, Smile, Paperclip, Send, Mic, Sparkles 
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip"
import { cn } from "@/lib/utils"
import { AISlashCommand } from "@/components/ai-chat/ai-slash-command"
import { MentionPopover } from "./mention-popover"
import { useUIStore } from "@/stores/ui-store"
import { toast } from "sonner"
import { EmojiPicker } from "./emoji-picker"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"

import { useEditor, EditorContent } from '@tiptap/react'
import StarterKit from '@tiptap/starter-kit'
import Placeholder from '@tiptap/extension-placeholder'
import TiptapLink from '@tiptap/extension-link'
import { useMemo, useEffect, useCallback } from "react"
import { useDraftStore } from "@/stores/draft-store"
import { usePresenceStore } from "@/stores/presence-store"

interface MessageComposerProps {
  placeholder?: string
  onSend?: (content: string) => void
  scope?: string
}

type TypingScope = {
  channelId?: string
  dmId?: string
  threadId?: string
}

export function MessageComposer({ placeholder, onSend, scope }: MessageComposerProps) {
  const [showSlashCommands, setShowSlashCommands] = useState(false)
  const [showMentions, setShowMentions] = useState(false)
  const [showFormatting, setShowFormatting] = useState(false)
  const { openCanvas } = useUIStore()
  const { saveDraft, drafts, fetchDrafts } = useDraftStore()
  const { sendTyping } = usePresenceStore()
  const fileInputRef = useRef<HTMLInputElement>(null)
  const typingTimeoutRef = useRef<NodeJS.Timeout | null>(null)

  const broadcastTyping = useCallback((isTyping: boolean) => {
    if (!scope) return
    const parts = scope.split(':')
    const scopeType = parts[0]
    const scopeId = parts[1]

    const scopeObj: TypingScope = {}
    if (scopeType === 'channel') scopeObj.channelId = scopeId
    else if (scopeType === 'dm') scopeObj.dmId = scopeId
    else if (scopeType === 'thread') scopeObj.threadId = scopeId

    sendTyping(scopeObj, isTyping)
  }, [scope, sendTyping])

  const handleTyping = useCallback(() => {
    if (typingTimeoutRef.current) {
      clearTimeout(typingTimeoutRef.current)
    } else {
      broadcastTyping(true)
    }

    typingTimeoutRef.current = setTimeout(() => {
      broadcastTyping(false)
      typingTimeoutRef.current = null
    }, 2000)
  }, [broadcastTyping])

  const extensions = useMemo(() => [
    StarterKit.configure({}),
    Placeholder.configure({
      placeholder: placeholder || 'Type a message...',
    }),
    TiptapLink.configure({
      openOnClick: false,
      HTMLAttributes: {
        class: 'text-blue-500 underline cursor-pointer',
      },
    }),
  ], [placeholder])

  const editor = useEditor({
    extensions,
    content: '',
    immediatelyRender: false,
    editorProps: {
      attributes: {
        class: 'min-h-[72px] max-h-[200px] overflow-y-auto px-3 py-2 text-sm block w-full',
      },
      handleKeyDown: (view, event) => {
        const keyboardEvent = event as KeyboardEvent & { isComposing?: boolean }

        if (event.key === 'Enter' && !event.shiftKey && !keyboardEvent.isComposing) {
          event.preventDefault()
          handleSend()
          return true
        }
        if (event.key === 'ArrowUp' && editor?.getText() === "") {
          console.log("Slack feature: Edit last message triggered")
        }
        if (event.key === 'Escape') {
          setShowSlashCommands(false)
          setShowMentions(false)
        }
        return false
      },
    },
    onUpdate: ({ editor }) => {
      const text = editor.getText()
      const lastChar = text.slice(-1)
      
      if (lastChar === "/") {
        setShowSlashCommands(true)
        setShowMentions(false)
      } else if (lastChar === "@") {
        setShowMentions(true)
        setShowSlashCommands(false)
      } else {
        if (!text.includes("/")) setShowSlashCommands(false)
        if (!text.includes("@")) setShowMentions(false)
      }

      // Autosave draft
      if (scope) {
        const content = editor.getHTML()
        // If empty content, save as empty string to backend
        if (content === "<p></p>" || editor.isEmpty) {
          saveDraft(scope, "")
        } else {
          saveDraft(scope, content)
        }
      }

      // Broadcast typing
      handleTyping()
    },
  })

  // Restore draft when scope changes
  useEffect(() => {
    if (editor && scope) {
      const draftContent = drafts[scope]?.content
      const currentContent = editor.getHTML()
      
      if (draftContent && draftContent !== currentContent) {
        editor.commands.setContent(draftContent)
      } else if (!draftContent && currentContent !== "<p></p>") {
        editor.commands.clearContent()
      }
    }
  }, [scope, editor, drafts])

  // Initial fetch of drafts
  useEffect(() => {
    fetchDrafts()
  }, [fetchDrafts])

  // We need to use a ref for handleSend to avoid dependency cycles in useEditor
  // or just use the editor instance inside handleSend directly which is fine since it's defined in the same scope.
  function handleSend() {
    if (editor && !editor.isEmpty) {
      onSend?.(editor.getHTML())
      editor.commands.clearContent()
      if (scope) saveDraft(scope, "") // Clear draft on send
      setShowSlashCommands(false)
      setShowMentions(false)
    }
  }

  const handleSlashCommandSelect = (cmd: string) => {
    if (cmd === "/canvas") {
      openCanvas("new-doc")
      editor?.commands.clearContent()
      setShowSlashCommands(false)
      return
    }
    // Replace the trailing slash with the command
    editor?.commands.insertContent(cmd.slice(1) + " ")
    setShowSlashCommands(false)
    editor?.commands.focus()
  }

  const handleMentionSelect = (name: string) => {
    // Replace trailing @ with mention (we can use a more advanced approach but this works for now)
    editor?.commands.insertContent(name + " ")
    setShowMentions(false)
    editor?.commands.focus()
  }

  const handleEmojiSelect = (emoji: string) => {
    editor?.commands.insertContent(emoji)
    editor?.commands.focus()
  }

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (file) {
      toast.success(`Attached file: ${file.name}`)
      e.target.value = ''
    }
  }

  if (!editor) {
    return null
  }

  return (
    <div className="p-4 pt-0 relative flex flex-col">
      <div className="relative w-full">
        {showSlashCommands && (
          <AISlashCommand onSelect={handleSlashCommandSelect} />
        )}
        {showMentions && (
          <MentionPopover onSelect={handleMentionSelect} />
        )}
      </div>
      
      <div className="border rounded-lg bg-white dark:bg-[#1a1d21] focus-within:ring-1 focus-within:ring-ring transition-shadow group shrink-0">
        {/* Toolbar */}
        {showFormatting && (
          <div className="flex items-center gap-0.5 p-1 border-b bg-muted/30">
            <ToolbarButton 
              icon={Bold} 
              tooltip="Bold (⌘B)" 
              onClick={() => editor.chain().focus().toggleBold().run()}
              active={editor.isActive('bold')}
            />
            <ToolbarButton 
              icon={Italic} 
              tooltip="Italic (⌘I)" 
              onClick={() => editor.chain().focus().toggleItalic().run()}
              active={editor.isActive('italic')}
            />
            <ToolbarButton 
              icon={Strikethrough} 
              tooltip="Strikethrough" 
              onClick={() => editor.chain().focus().toggleStrike().run()}
              active={editor.isActive('strike')}
            />
            <div className="w-[1px] h-4 bg-border mx-1" />
            <ToolbarButton 
              icon={LinkIcon} 
              tooltip="Link" 
              onClick={() => {
                const url = window.prompt('URL')
                if (url) editor.chain().focus().setLink({ href: url }).run()
              }}
              active={editor.isActive('link')}
            />
            <div className="w-[1px] h-4 bg-border mx-1" />
            <ToolbarButton 
              icon={ListOrdered} 
              tooltip="Ordered list" 
              onClick={() => editor.chain().focus().toggleOrderedList().run()}
              active={editor.isActive('orderedList')}
            />
            <ToolbarButton 
              icon={List} 
              tooltip="Bulleted list" 
              onClick={() => editor.chain().focus().toggleBulletList().run()}
              active={editor.isActive('bulletList')}
            />
            <div className="w-[1px] h-4 bg-border mx-1" />
            <ToolbarButton 
              icon={Quote} 
              tooltip="Blockquote" 
              onClick={() => editor.chain().focus().toggleBlockquote().run()}
              active={editor.isActive('blockquote')}
            />
            <div className="w-[1px] h-4 bg-border mx-1" />
            <ToolbarButton 
              icon={Code} 
              tooltip="Code" 
              onClick={() => editor.chain().focus().toggleCode().run()}
              active={editor.isActive('code')}
            />
            <ToolbarButton 
              icon={FileCode} 
              tooltip="Code block" 
              onClick={() => editor.chain().focus().toggleCodeBlock().run()}
              active={editor.isActive('codeBlock')}
            />
          </div>
        )}

        {/* Editor Content */}
        <EditorContent editor={editor} />
        <input 
          type="file" 
          ref={fileInputRef} 
          className="hidden" 
          onChange={handleFileSelect} 
        />

        {/* Bottom Actions */}
        <div className="flex items-center justify-between px-2 pb-2">
          <div className="flex items-center gap-0.5">
            <DropdownMenu>
              <Tooltip>
                <TooltipTrigger asChild>
                  <DropdownMenuTrigger asChild>
                    <Button variant="ghost" size="icon" className="h-8 w-8 text-muted-foreground hover:bg-muted">
                      <PlusIcon className="h-4 w-4" />
                    </Button>
                  </DropdownMenuTrigger>
                </TooltipTrigger>
                <TooltipContent>Shortcuts</TooltipContent>
              </Tooltip>
              <DropdownMenuContent align="start" className="w-56">
                <DropdownMenuLabel>Message Shortcuts</DropdownMenuLabel>
                <DropdownMenuSeparator />
                <DropdownMenuItem className="flex justify-between" onClick={() => toast.info("Post message clicked")}>
                  <span>Post message</span>
                  <span className="text-xs text-muted-foreground">Enter</span>
                </DropdownMenuItem>
                <DropdownMenuItem className="flex justify-between" onClick={() => toast.info("New line clicked")}>
                  <span>New line</span>
                  <span className="text-xs text-muted-foreground">Shift + Enter</span>
                </DropdownMenuItem>
                <DropdownMenuItem className="flex justify-between" onClick={() => toast.info("Bold clicked")}>
                  <span>Bold</span>
                  <span className="text-xs text-muted-foreground">⌘ B</span>
                </DropdownMenuItem>
                <DropdownMenuItem className="flex justify-between" onClick={() => toast.info("Italic clicked")}>
                  <span>Italic</span>
                  <span className="text-xs text-muted-foreground">⌘ I</span>
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>

            <div className="w-[1px] h-4 bg-border mx-1" />
            <ToolbarButton 
              icon={Type} 
              tooltip="Formatting" 
              active={showFormatting}
              onClick={() => setShowFormatting(prev => !prev)} 
            />
            <EmojiPicker onSelect={handleEmojiSelect}>
              <div className="flex">
                <ToolbarButton icon={Smile} tooltip="Emoji" />
              </div>
            </EmojiPicker>
            <ToolbarButton icon={Paperclip} tooltip="Attach file" onClick={() => fileInputRef.current?.click()} />
            <ToolbarButton icon={Mic} tooltip="Video/Audio clip" onClick={() => toast.info("Video/Audio clip coming soon")} />
            <div className="w-[1px] h-4 bg-border mx-1" />
            <Tooltip>
              <TooltipTrigger asChild>
                <Button variant="ghost" size="icon" className="h-8 w-8 text-purple-500 hover:bg-purple-500/10" onClick={() => openCanvas('ai-assistant')}>
                  <Sparkles className="h-4 w-4" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>AI Canvas</TooltipContent>
            </Tooltip>
          </div>
          
          <Button 
            size="sm" 
            disabled={editor.isEmpty}
            onClick={handleSend}
            className={cn(
              "h-7 w-7 p-0 transition-colors",
              !editor.isEmpty ? "bg-[#007a5a] hover:bg-[#007a5a]/90 text-white" : "bg-muted text-muted-foreground"
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

function ToolbarButton({ icon: Icon, tooltip, active, onClick }: { icon: React.ElementType, tooltip: string, active?: boolean, onClick?: () => void }) {
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Button 
          variant="ghost" 
          size="icon" 
          onClick={onClick}
          className={cn(
            "h-7 w-7 hover:bg-muted transition-colors",
            active ? "bg-muted text-blue-500" : "text-muted-foreground"
          )}
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

function PlusIcon({ className }: { className?: string }) {
  return (
    <div className={cn("w-4 h-4 rounded-full border-2 border-current flex items-center justify-center font-bold text-[10px]", className)}>
      +
    </div>
  )
}
