"use client"

import { useState, useRef, useCallback, useMemo, useEffect, useLayoutEffect } from "react"
import { 
  Bold, Italic, Strikethrough, Link as LinkIcon, List, ListOrdered, 
  Quote, Code, FileCode, Type, Smile, Globe, Loader2, Paperclip, Send, Mic, Sparkles, X,
  Wand2, RefreshCw, ThumbsUp, ThumbsDown,
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
import { useDraftStore } from "@/stores/draft-store"
import { usePresenceStore } from "@/stores/presence-store"
import { useFileStore } from "@/stores/file-store"
import { useKnowledgeStore } from "@/stores/knowledge-store"
import { useWorkspaceStore } from "@/stores/workspace-store"
import { useChannelStore } from "@/stores/channel-store"
import type { EntitySuggestResult, EntityTextMatch, ComposeSuggestion } from "@/types"

interface MessageComposerProps {
  placeholder?: string
  onSend?: (content: string, artifactIds?: string[], fileIds?: string[]) => void
  scope?: string
}

type TypingScope = {
  channelId?: string
  dmId?: string
  threadId?: string
}

export function MessageComposer({ placeholder, onSend, scope }: MessageComposerProps) {
  const [showSlashCommands, setShowSlashCommands] = useState(false)
  const [slashQuery, setSlashQuery] = useState("")
  const [showMentions, setShowMentions] = useState(false)
  const [showFormatting, setShowFormatting] = useState(false)
  const [uploadedFileIds, setUploadedFileIds] = useState<string[]>([])
  const [showEntityMentions, setShowEntityMentions] = useState(false)
  const [entityMentionQuery, setEntityMentionQuery] = useState("")
  const [entitySuggestions, setEntitySuggestions] = useState<EntitySuggestResult[]>([])
  const [isLoadingEntitySuggestions, setIsLoadingEntitySuggestions] = useState(false)
  const entityMentionStartRef = useRef<number>(0)
  const entitySuggestTimeoutRef = useRef<NodeJS.Timeout | null>(null)
  const { openCanvas } = useUIStore()
  const { saveDraft, deleteDraft, drafts, fetchDrafts } = useDraftStore()
  const { sendTyping } = usePresenceStore()
  const { uploadFile } = useFileStore()
  const {
    suggestEntities, matchEntitiesInText,
    clearComposeResult, composeResults, isComposing,
    suggestComposeStream, sendComposeFeedback, composeStreaming, composeFeedback,
  } = useKnowledgeStore()
  const { currentWorkspace } = useWorkspaceStore()
  const { currentChannel } = useChannelStore()

  // Phase 63B: AI Compose suggestion state
  const [showComposeSuggestions, setShowComposeSuggestions] = useState(false)

  // Derive channel/thread ids for /ai/compose (only for channel & thread scopes)
  const composeIds = useMemo(() => {
    if (!scope) return null
    if (scope.startsWith('channel:')) {
      return { channelId: scope.slice('channel:'.length), threadId: undefined as string | undefined }
    }
    if (scope.startsWith('thread:')) {
      const threadId = scope.slice('thread:'.length)
      if (currentChannel?.id) return { channelId: currentChannel.id, threadId }
    }
    return null
  }, [scope, currentChannel?.id])

  const composeKey = composeIds ? `${composeIds.channelId}:${composeIds.threadId || ''}` : null
  const composeResult = composeKey ? composeResults[composeKey] : undefined
  const composeBusy = composeKey ? !!isComposing[composeKey] : false
  const composeStream = composeKey ? composeStreaming[composeKey] : undefined
  const fileInputRef = useRef<HTMLInputElement>(null)
  const typingTimeoutRef = useRef<NodeJS.Timeout | null>(null)
  const draftsRef = useRef(drafts)

  // Phase 55: passive entity reverse-lookup from the composer text
  const [detectedMatches, setDetectedMatches] = useState<EntityTextMatch[]>([])
  const [dismissedMatchKeys, setDismissedMatchKeys] = useState<Set<string>>(new Set())
  const matchTimeoutRef = useRef<NodeJS.Timeout | null>(null)
  const lastMatchedTextRef = useRef<string>("")

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
    StarterKit.configure({
      // Disable the built-in extensions we add separately
      link: false,
    }),
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
      
      // Handle slash commands and query extraction
      const slashIndex = text.lastIndexOf('/')
      if (slashIndex !== -1 && !text.slice(slashIndex + 1).includes(' ')) {
        const query = text.slice(slashIndex + 1)
        setSlashQuery(query)
        setShowSlashCommands(true)
        setShowMentions(false)
      } else {
        setShowSlashCommands(false)
        setSlashQuery("")
      }

      if (lastChar === "@") {
        setShowMentions(true)
        setShowSlashCommands(false)
        setShowEntityMentions(false)
      } else {
        if (!text.includes("@")) setShowMentions(false)
      }

      // @entity: autocomplete
      const entityMatch = text.match(/@entity:([^\s]*)$/i)
      if (entityMatch) {
        setShowEntityMentions(true)
        setShowMentions(false)
        setShowSlashCommands(false)
        const q = entityMatch[1]
        setEntityMentionQuery(q)
        const cursorPos = editor.state.selection.to
        entityMentionStartRef.current = cursorPos - entityMatch[0].length
        if (entitySuggestTimeoutRef.current) clearTimeout(entitySuggestTimeoutRef.current)
        entitySuggestTimeoutRef.current = setTimeout(async () => {
          setIsLoadingEntitySuggestions(true)
          const channelId = scope?.startsWith('channel:') ? scope.split(':')[1] : undefined
          const results = await suggestEntities(q, channelId)
          setEntitySuggestions(results)
          setIsLoadingEntitySuggestions(false)
        }, 180)
      } else {
        setShowEntityMentions(false)
        setEntityMentionQuery("")
      }

      // Autosave draft
      if (scope) {
        const content = editor.getHTML()
        // If empty content, explicitly delete draft from backend
        if (content === "<p></p>" || editor.isEmpty) {
          deleteDraft(scope)
        } else {
          saveDraft(scope, content)
        }
      }

      // Phase 55: passive entity reverse-lookup (match-text)
      // Skip while the user is mid-@mention / @entity: / slash command
      if (matchTimeoutRef.current) clearTimeout(matchTimeoutRef.current)
      const trimmed = text.trim()
      const isMidMention = /@entity:[^\s]*$/i.test(text) || lastChar === "@" || slashIndex !== -1
      if (!trimmed || editor.isEmpty) {
        setDetectedMatches([])
        lastMatchedTextRef.current = ""
      } else if (!isMidMention && currentWorkspace?.id) {
        matchTimeoutRef.current = setTimeout(async () => {
          if (trimmed === lastMatchedTextRef.current) return
          lastMatchedTextRef.current = trimmed
          const matches = await matchEntitiesInText(currentWorkspace.id, trimmed, 10)
          // Hide matches that fall immediately after a leading `@` (already an explicit mention)
          const filtered = matches.filter(m => {
            const charBefore = m.start > 0 ? trimmed[m.start - 1] : ""
            return charBefore !== "@"
          })
          setDetectedMatches(filtered)
        }, 500)
      }

      // Broadcast typing
      handleTyping()
    },
  })

  // Keep draftsRef current so scope-change effect doesn't need drafts in its deps
  useLayoutEffect(() => {
    draftsRef.current = drafts
  })

  // Restore draft when scope changes (NOT when drafts change, to avoid re-populating after send)
  useEffect(() => {
    if (editor && scope) {
      const draftContent = draftsRef.current[scope]?.content
      const currentContent = editor.getHTML()

      if (draftContent && draftContent !== currentContent) {
        editor.commands.setContent(draftContent)
      } else if (!draftContent && currentContent !== "<p></p>") {
        editor.commands.clearContent()
      }
    }
  }, [scope, editor])

  // Initial fetch of drafts
  useEffect(() => {
    fetchDrafts()
  }, [fetchDrafts])

  // We need to use a ref for handleSend to avoid dependency cycles in useEditor
  // or just use the editor instance inside handleSend directly which is fine since it's defined in the same scope.
  function handleSend() {
    if (editor && !editor.isEmpty) {
      const htmlContent = editor.getHTML()
      const textContent = editor.getText().trim()

      // Intercept slash commands in plain text form
      if (textContent === "/canvas") {
        openCanvas("new-doc")
        editor.commands.clearContent()
        if (scope) deleteDraft(scope)
        setShowSlashCommands(false)
        return
      }

      onSend?.(htmlContent, [], uploadedFileIds)
      editor.commands.clearContent()
      setUploadedFileIds([])
      if (scope) deleteDraft(scope) // Clear draft on send
      setShowSlashCommands(false)
      setShowMentions(false)
      setDetectedMatches([])
      setDismissedMatchKeys(new Set())
      lastMatchedTextRef.current = ""
    }
  }

  // Phase 55: convert a detected entity text span into an explicit @mention.
  // Finds the match text inside the live editor doc (safer than trusting
  // plain-text offsets across HTML) and replaces it with "@<title> ".
  const convertMatchToMention = useCallback((match: EntityTextMatch) => {
    if (!editor) return
    const docText = editor.state.doc.textContent
    const idx = docText.indexOf(match.matched_text)
    if (idx < 0) {
      // Text no longer present (user likely edited it away); just drop it
      setDetectedMatches(prev => prev.filter(m => m.entity_id !== match.entity_id))
      return
    }
    // For typical single-paragraph drafts ProseMirror adds +1 for the opening
    // paragraph node boundary; this mapping is correct for the common case
    // and safely no-ops via the indexOf check above otherwise.
    const from = idx + 1
    const to = from + match.matched_text.length
    editor.chain().focus().deleteRange({ from, to }).insertContent(`@${match.entity_title} `).run()
    setDetectedMatches(prev => prev.filter(m => m.entity_id !== match.entity_id))
    // Ensure the next onUpdate re-evaluates against the new text
    lastMatchedTextRef.current = ""
  }, [editor])

  const dismissMatch = useCallback((match: EntityTextMatch) => {
    const key = `${match.entity_id}:${match.matched_text}`
    setDismissedMatchKeys(prev => {
      const next = new Set(prev)
      next.add(key)
      return next
    })
    setDetectedMatches(prev => prev.filter(m => `${m.entity_id}:${m.matched_text}` !== key))
  }, [])

  const visibleDetectedMatches = useMemo(
    () => detectedMatches.filter(m => !dismissedMatchKeys.has(`${m.entity_id}:${m.matched_text}`)),
    [detectedMatches, dismissedMatchKeys]
  )

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

  const handleEntityMentionSelect = useCallback((entity: EntitySuggestResult) => {
    if (!editor) return
    const from = entityMentionStartRef.current
    const to = editor.state.selection.to
    editor.chain().focus().deleteRange({ from, to }).insertContent(`@${entity.entity_title} `).run()
    setShowEntityMentions(false)
    setEntitySuggestions([])
  }, [editor])

  const handleEmojiSelect = (emoji: string) => {
    editor?.commands.insertContent(emoji)
    editor?.commands.focus()
  }

  // Phase 63B/C: Request grounded compose suggestions for the current scope.
  // Phase 63C upgrade: prefer SSE streaming, fall back to sync on failure (handled in store).
  const runSuggestCompose = useCallback(async () => {
    if (!composeIds || !editor) return
    const draftText = editor.getText().trim()
    setShowComposeSuggestions(true)
    await suggestComposeStream(composeIds.channelId, composeIds.threadId, draftText, 3)
  }, [composeIds, editor, suggestComposeStream])

  // Phase 63B/C: Insert a suggestion into the draft (no auto-send) + fire 'edited' feedback signal.
  const handleInsertSuggestion = useCallback((s: ComposeSuggestion) => {
    if (!editor || !composeIds) return
    const html = `<p>${s.text
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/\n{2,}/g, '</p><p>')
      .replace(/\n/g, '<br/>')}</p>`
    editor.chain().focus().clearContent().insertContent(html).run()
    setShowComposeSuggestions(false)
    toast.success('Suggestion inserted — edit and send when ready')
    // Fire 'edited' feedback in the background (non-blocking)
    if (s.id && !composeFeedback[s.id]) {
      sendComposeFeedback(s.id, {
        channelId: composeIds.channelId,
        threadId: composeIds.threadId,
        feedback: 'edited',
        suggestionText: s.text,
        provider: composeResult?.provider,
        model: composeResult?.model,
      }).catch(() => { /* best-effort */ })
    }
  }, [editor, composeIds, composeFeedback, sendComposeFeedback, composeResult])

  // Phase 63C: thumbs up/down feedback for a specific suggestion
  const handleSuggestionFeedback = useCallback(async (s: ComposeSuggestion, value: 'up' | 'down') => {
    if (!composeIds || !s.id) return
    const ok = await sendComposeFeedback(s.id, {
      channelId: composeIds.channelId,
      threadId: composeIds.threadId,
      feedback: value,
      suggestionText: s.text,
      provider: composeResult?.provider,
      model: composeResult?.model,
    })
    if (ok) {
      toast.success(value === 'up' ? 'Thanks — feedback recorded' : 'Noted — we\u2019ll learn from it')
    }
  }, [composeIds, composeResult, sendComposeFeedback])

  const handleFileSelect = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (file && editor) {
      let channelId: string | undefined
      if (scope?.startsWith('channel:')) {
        channelId = scope.split(':')[1]
      }

      const uploadedFile = await uploadFile(file, channelId)
      if (uploadedFile) {
        setUploadedFileIds(prev => [...prev, uploadedFile.id])
        // Insert a link to the file in the editor
        editor.chain().focus().insertContent([
          {
            type: 'text',
            text: `📎 ${uploadedFile.name}`,
            marks: [
              {
                type: 'link',
                attrs: {
                  href: uploadedFile.url,
                  target: '_blank',
                },
              },
            ],
          },
          {
            type: 'text',
            text: ' ',
          },
        ]).run()
      }
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
          <AISlashCommand onSelect={handleSlashCommandSelect} query={slashQuery} />
        )}
        {showMentions && (
          <MentionPopover onSelect={handleMentionSelect} />
        )}
        {showEntityMentions && (
          <div className="absolute bottom-0 left-0 right-0 bg-white dark:bg-[#1a1d21] border rounded-xl shadow-xl overflow-hidden z-50 max-h-64">
            <div className="px-3 py-2 border-b flex items-center gap-2">
              <Globe className="w-3 h-3 text-emerald-600" />
              <span className="text-[10px] font-black uppercase tracking-widest text-emerald-700">Entity Mention</span>
              {entityMentionQuery && <span className="text-[10px] text-muted-foreground ml-1">&ldquo;{entityMentionQuery}&rdquo;</span>}
              {isLoadingEntitySuggestions && <Loader2 className="w-3 h-3 animate-spin ml-auto text-muted-foreground" />}
            </div>
            {entitySuggestions.length === 0 && !isLoadingEntitySuggestions && (
              <div className="px-3 py-4 text-xs text-muted-foreground text-center italic">
                {entityMentionQuery ? 'No entities found' : 'Type to search entities...'}
              </div>
            )}
            {entitySuggestions.map((ent) => (
              <button
                key={ent.entity_id}
                className="w-full text-left px-3 py-2 hover:bg-muted/60 flex items-center gap-2 transition-colors"
                onMouseDown={e => { e.preventDefault(); handleEntityMentionSelect(ent) }}
              >
                <Globe className="w-3 h-3 text-emerald-600 shrink-0" />
                <span className="text-xs font-bold truncate">{ent.entity_title}</span>
                <span className="text-[10px] text-muted-foreground uppercase shrink-0">{ent.entity_kind}</span>
                {ent.ref_count !== undefined && ent.ref_count > 0 && (
                  <span className="ml-auto text-[9px] text-muted-foreground shrink-0">{ent.ref_count} refs</span>
                )}
              </button>
            ))}
            <div className="px-3 py-1.5 border-t">
              <p className="text-[9px] text-muted-foreground">Type <code className="font-mono">@entity:name</code> to mention a knowledge entity</p>
            </div>
          </div>
        )}
      </div>

      {/* Phase 63B: AI Compose suggestions */}
      {composeIds && showComposeSuggestions && (composeBusy || composeResult) && (
        <div className="mb-2 rounded-xl border border-sky-400/40 bg-gradient-to-br from-sky-500/5 to-cyan-500/5 overflow-hidden">
          <div className="px-3 py-1.5 border-b border-sky-400/20 flex items-center gap-2">
            <Wand2 className="w-3 h-3 text-sky-500" />
            <span className="text-[10px] font-black uppercase tracking-widest text-sky-700 dark:text-sky-400">
              AI Suggestions
            </span>
            {composeResult?.provider && (
              <span className="text-[9px] font-mono text-muted-foreground">
                {composeResult.provider}{composeResult.model ? ` / ${composeResult.model}` : ''}
              </span>
            )}
            <div className="ml-auto flex items-center gap-1">
              <Button
                variant="ghost"
                size="sm"
                className="h-6 text-[10px] gap-1 px-2 text-sky-700 dark:text-sky-400 hover:bg-sky-500/10"
                disabled={composeBusy}
                onClick={runSuggestCompose}
                title="Regenerate suggestions"
              >
                {composeBusy ? <Loader2 className="w-3 h-3 animate-spin" /> : <RefreshCw className="w-3 h-3" />}
                {composeBusy ? 'Thinking…' : 'Regenerate'}
              </Button>
              <Button
                variant="ghost"
                size="sm"
                className="h-6 w-6 p-0 text-muted-foreground"
                onClick={() => {
                  setShowComposeSuggestions(false)
                  if (composeIds) clearComposeResult(composeIds.channelId, composeIds.threadId)
                }}
                title="Dismiss"
              >
                <X className="w-3 h-3" />
              </Button>
            </div>
          </div>
          {composeBusy && !composeResult && !composeStream && (
            <div className="px-3 py-4 text-[11px] text-muted-foreground italic flex items-center gap-2">
              <Loader2 className="w-3 h-3 animate-spin" /> Drafting grounded reply suggestions…
            </div>
          )}
          {composeBusy && composeStream && !composeResult && (
            <div className="px-3 py-2">
              <div className="rounded-lg border border-sky-400/40 bg-background/60 p-2 space-y-1.5">
                <div className="flex items-center gap-1.5">
                  <Loader2 className="w-3 h-3 animate-spin text-sky-500" />
                  <span className="text-[9px] font-black uppercase tracking-widest text-sky-700 dark:text-sky-400">Streaming…</span>
                </div>
                <p className="text-xs text-foreground/85 leading-relaxed whitespace-pre-wrap">
                  {composeStream.text}
                  <span className="inline-block w-1.5 h-3 align-text-bottom bg-sky-500/70 animate-pulse ml-0.5" />
                </p>
              </div>
            </div>
          )}
          {composeResult && (
            <div className="px-3 py-2 space-y-2">
              {composeResult.suggestions.length === 0 && (
                <p className="text-[11px] text-muted-foreground italic">No suggestions returned.</p>
              )}
              {composeResult.suggestions.map((s) => {
                const fb = composeFeedback[s.id]
                return (
                  <div
                    key={s.id}
                    className="rounded-lg border bg-background/60 p-2 space-y-1.5 hover:border-sky-400/40 transition-colors"
                  >
                    <p className="text-xs text-foreground/90 leading-relaxed whitespace-pre-wrap">{s.text}</p>
                    <div className="flex items-center gap-1.5 flex-wrap">
                      {s.tone && (
                        <span className="text-[9px] font-bold uppercase tracking-wider px-1.5 py-0.5 rounded border bg-sky-500/5 text-sky-700 dark:text-sky-400 border-sky-400/30">
                          {s.tone}
                        </span>
                      )}
                      {s.kind && s.kind !== s.tone && (
                        <span className="text-[9px] font-bold uppercase tracking-wider px-1.5 py-0.5 rounded border bg-muted/40 text-muted-foreground">
                          {s.kind}
                        </span>
                      )}
                      {fb === 'edited' && (
                        <span className="text-[9px] font-bold uppercase tracking-wider px-1.5 py-0.5 rounded border bg-emerald-500/10 text-emerald-700 dark:text-emerald-400 border-emerald-400/30">
                          used
                        </span>
                      )}
                      <div className="ml-auto flex items-center gap-1">
                        <Button
                          variant="ghost"
                          size="sm"
                          className={cn(
                            "h-6 w-6 p-0 text-muted-foreground hover:bg-emerald-500/10",
                            fb === 'up' && "text-emerald-600 bg-emerald-500/10"
                          )}
                          disabled={!!fb && fb !== 'up'}
                          onClick={() => handleSuggestionFeedback(s, 'up')}
                          title="Good suggestion"
                        >
                          <ThumbsUp className="w-3 h-3" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          className={cn(
                            "h-6 w-6 p-0 text-muted-foreground hover:bg-rose-500/10",
                            fb === 'down' && "text-rose-600 bg-rose-500/10"
                          )}
                          disabled={!!fb && fb !== 'down'}
                          onClick={() => handleSuggestionFeedback(s, 'down')}
                          title="Not helpful"
                        >
                          <ThumbsDown className="w-3 h-3" />
                        </Button>
                        <Button
                          size="sm"
                          variant="outline"
                          className="h-6 text-[10px] gap-1 px-2 border-sky-400/40 text-sky-700 dark:text-sky-400 hover:bg-sky-500/10"
                          onClick={() => handleInsertSuggestion(s)}
                        >
                          <Sparkles className="w-3 h-3" /> Insert into draft
                        </Button>
                      </div>
                    </div>
                  </div>
                )
              })}
              {(composeResult.context_entities?.length || composeResult.citations?.length) ? (
                <div className="pt-1 border-t border-sky-400/15 space-y-1">
                  {composeResult.context_entities && composeResult.context_entities.length > 0 && (
                    <div className="flex items-center gap-1.5 flex-wrap">
                      <span className="text-[9px] font-black uppercase tracking-widest text-muted-foreground">Context</span>
                      {composeResult.context_entities.slice(0, 6).map(ent => (
                        <span
                          key={ent.id}
                          className="inline-flex items-center gap-1 text-[10px] rounded-md border bg-emerald-500/5 text-emerald-700 dark:text-emerald-400 border-emerald-400/30 px-1.5 py-0.5"
                        >
                          <Globe className="w-2.5 h-2.5" />
                          <span className="truncate max-w-[140px] font-bold">{ent.title}</span>
                          <span className="text-[8px] uppercase opacity-70">{ent.kind}</span>
                        </span>
                      ))}
                    </div>
                  )}
                  {composeResult.citations && composeResult.citations.length > 0 && (
                    <details className="group">
                      <summary className="text-[9px] font-black uppercase tracking-widest text-muted-foreground cursor-pointer select-none hover:text-foreground">
                        Citations ({composeResult.citations.length})
                      </summary>
                      <div className="mt-1 space-y-1 pl-1">
                        {composeResult.citations.slice(0, 6).map((c, i) => (
                          <div key={c.id || i} className="flex items-start gap-1.5 text-[10px]">
                            <span className="text-[9px] font-bold uppercase tracking-wider px-1.5 py-0.5 rounded border bg-muted/40 text-muted-foreground shrink-0">
                              {c.source_kind || c.ref_kind}
                            </span>
                            <span className="text-foreground/75 italic line-clamp-2">&ldquo;{c.snippet}&rdquo;</span>
                          </div>
                        ))}
                      </div>
                    </details>
                  )}
                </div>
              ) : null}
            </div>
          )}
        </div>
      )}

      {/* Phase 55: passive entity detection hint row */}
      {visibleDetectedMatches.length > 0 && !showEntityMentions && !showSlashCommands && !showMentions && (
        <div className="mb-1.5 flex items-center gap-1.5 flex-wrap px-1">
          <div className="flex items-center gap-1 text-[9px] font-black uppercase tracking-widest text-purple-700 dark:text-purple-400 shrink-0">
            <Sparkles className="w-2.5 h-2.5" />
            Knowledge detected
          </div>
          {visibleDetectedMatches.slice(0, 5).map((m) => (
            <div
              key={`${m.entity_id}:${m.start}`}
              className="inline-flex items-center gap-0.5 text-[10px] rounded-md border bg-purple-500/10 text-purple-700 dark:text-purple-400 border-purple-300 dark:border-purple-700 pl-1.5 pr-0.5 py-0.5"
            >
              <button
                onClick={() => convertMatchToMention(m)}
                className="inline-flex items-center gap-1 font-bold hover:underline"
                title={`Convert "${m.matched_text}" into @${m.entity_title} mention`}
              >
                <Globe className="w-2.5 h-2.5" />
                <span className="truncate max-w-[140px]">{m.entity_title}</span>
                <span className="text-[8px] uppercase text-purple-600/70 dark:text-purple-300/70 ml-0.5">
                  {m.entity_kind}
                </span>
              </button>
              <button
                onClick={() => dismissMatch(m)}
                className="ml-0.5 h-3.5 w-3.5 rounded hover:bg-purple-500/20 inline-flex items-center justify-center text-purple-700/70 dark:text-purple-300/70"
                title="Dismiss"
              >
                <X className="w-2 h-2" />
              </button>
            </div>
          ))}
          {visibleDetectedMatches.length > 5 && (
            <span className="text-[9px] text-muted-foreground">+{visibleDetectedMatches.length - 5} more</span>
          )}
        </div>
      )}

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
            {composeIds && (
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8 text-sky-500 hover:bg-sky-500/10"
                    disabled={composeBusy}
                    onClick={runSuggestCompose}
                  >
                    {composeBusy ? <Loader2 className="h-4 w-4 animate-spin" /> : <Wand2 className="h-4 w-4" />}
                  </Button>
                </TooltipTrigger>
                <TooltipContent>AI Suggest (grounded reply)</TooltipContent>
              </Tooltip>
            )}
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
