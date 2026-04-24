"use client"

// ── Canvas AI Dock ───────────────────────────────────────────────────────────
//
// A ChatGPT/Gemini Canvas-style AI chat rail that lives at the bottom of the
// Canvas panel while a document-type artifact is being edited. It replaces
// the previous modal AI-edit dialog so the user never loses cursor / selection
// context while iterating with the model.
//
// Features:
//   • Persistent composer with streaming SSE preview (reuses `/api/v1/ai/execute`
//     contract: event:"chunk" / {text:"…"}).
//   • Slash-command menu (/expand /shorten /rephrase /fix /formal /casual /bullets
//     /outline /summary /continue /translate-en /translate-zh) — arrow-key
//     navigation, Tab/Enter to insert.
//   • Target awareness — a live chip shows whether the next run will rewrite
//     the current selection or the whole document, with char counts.
//   • Chat-bubble history with per-message Apply / Insert at cursor / Copy /
//     Retry / Stop actions.
//   • Keyboard: ⌘K focuses the composer, Enter sends, Shift+Enter newline,
//     Esc closes the slash menu.
//
// The dock talks to the editor through a `CanvasEditorHandle` ref (see
// canvas-tiptap-editor.tsx) — no need to lift editor state up.

import {
  forwardRef, useCallback, useEffect, useImperativeHandle, useMemo, useRef, useState,
} from "react"
import {
  Sparkles, Send, X, Loader2, Check, RefreshCw, ChevronDown, ChevronUp,
  Copy, Wand2, Trash2,
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"
import { API_BASE_URL } from "@/lib/constants"
import { toast } from "sonner"
import type { CanvasEditorHandle } from "./canvas-tiptap-editor"

// ── Public handle ────────────────────────────────────────────────────────────

export interface CanvasAIDockHandle {
  /** Focus the composer input and expand the dock. */
  focusInput: () => void
}

// ── Slash commands ───────────────────────────────────────────────────────────

interface SlashCommand {
  cmd: string
  label: string
  instruction: string
  /** Emoji / glyph shown in the menu. */
  glyph: string
}

const SLASH_COMMANDS: SlashCommand[] = [
  { cmd: "/expand",       glyph: "➕", label: "Expand",
    instruction: "Expand the text below with additional detail, examples, and context, while preserving the original meaning and tone." },
  { cmd: "/shorten",      glyph: "✂︎", label: "Shorten",
    instruction: "Shorten the text below to its essential points while preserving meaning and tone." },
  { cmd: "/rephrase",     glyph: "↻",  label: "Rephrase",
    instruction: "Rephrase the text below with clearer wording while preserving meaning." },
  { cmd: "/fix",          glyph: "✓",  label: "Fix grammar & spelling",
    instruction: "Correct spelling, grammar, and punctuation in the text below without changing its meaning." },
  { cmd: "/formal",       glyph: "🎩", label: "Formal tone",
    instruction: "Rewrite the text below in a more professional, formal tone." },
  { cmd: "/casual",       glyph: "😊", label: "Casual tone",
    instruction: "Rewrite the text below in a more relaxed, conversational tone." },
  { cmd: "/bullets",      glyph: "•",  label: "Convert to bullet list",
    instruction: "Convert the text below into a clear bulleted list. Preserve the original meaning." },
  { cmd: "/outline",      glyph: "☰",  label: "Generate outline",
    instruction: "Produce a structured outline (headings + sub-bullets) that captures the key ideas of the text below." },
  { cmd: "/summary",      glyph: "∑",  label: "Summarize",
    instruction: "Summarize the text below in 2-3 concise paragraphs. Preserve the key facts and arguments." },
  { cmd: "/continue",     glyph: "→",  label: "Continue writing",
    instruction: "Continue writing naturally from where the text below ends. Match the existing voice and style. Return ONLY the continuation, not the original text." },
  { cmd: "/translate-en", glyph: "🇬🇧", label: "Translate → English",
    instruction: "Translate the text below into natural, fluent English." },
  { cmd: "/translate-zh", glyph: "🇨🇳", label: "翻译 → 中文",
    instruction: "将下面的文字翻译成自然流畅的中文。" },
]

// ── Message shape ────────────────────────────────────────────────────────────

interface DockMessage {
  id: string
  role: "user" | "assistant"
  /** For user messages: the label shown in the bubble (the prompt). For
   *  assistant messages: the streaming accumulator. */
  text: string
  /** Only on assistant messages — the full instruction sent to the LLM
   *  (used by Retry). */
  instruction?: string
  /** Target scope captured at send-time so Retry/Apply honor the original
   *  intent even if the user changes their selection mid-stream. */
  targetRange?: "selection" | "document"
  targetPreview?: string
  isStreaming?: boolean
  applied?: boolean
  errored?: boolean
}

// ── Props ────────────────────────────────────────────────────────────────────

interface CanvasAIDockProps {
  editorRef: React.RefObject<CanvasEditorHandle | null>
  channelId: string
  /** Shown at top of the dock when no chat yet — e.g. "Drafting: My launch plan". */
  artifactTitle?: string
}

// ── Component ────────────────────────────────────────────────────────────────

export const CanvasAIDock = forwardRef<CanvasAIDockHandle, CanvasAIDockProps>(
  function CanvasAIDock({ editorRef, channelId, artifactTitle }, ref) {
    const [input, setInput] = useState("")
    const [expanded, setExpanded] = useState(false)
    const [messages, setMessages] = useState<DockMessage[]>([])

    // Slash-menu state — derived from `input`.
    const [showSlashMenu, setShowSlashMenu] = useState(false)
    const [slashFilter, setSlashFilter] = useState("")
    const [slashSelected, setSlashSelected] = useState(0)

    const [streamingId, setStreamingId] = useState<string | null>(null)
    const abortCtrlRef = useRef<AbortController | null>(null)
    const inputRef = useRef<HTMLTextAreaElement>(null)
    const chatEndRef = useRef<HTMLDivElement>(null)

    // Expose `focusInput` to parent.
    useImperativeHandle(ref, () => ({
      focusInput: () => {
        setExpanded(true)
        setTimeout(() => inputRef.current?.focus(), 30)
      },
    }), [])

    const filteredCommands = useMemo(() => {
      if (!slashFilter) return SLASH_COMMANDS
      const f = slashFilter.toLowerCase()
      return SLASH_COMMANDS.filter(c =>
        c.cmd.toLowerCase().includes(f) || c.label.toLowerCase().includes(f),
      )
    }, [slashFilter])

    // Slash menu visibility — re-derive whenever the input changes.
    useEffect(() => {
      if (input.startsWith("/")) {
        const filter = input.slice(1).split(/\s/)[0]
        setSlashFilter(filter)
        setShowSlashMenu(true)
        setSlashSelected(0)
      } else {
        setShowSlashMenu(false)
      }
    }, [input])

    // Auto-scroll chat to the newest bubble.
    useEffect(() => {
      chatEndRef.current?.scrollIntoView({ behavior: "smooth", block: "end" })
    }, [messages])

    // Global ⌘K to focus the composer.
    useEffect(() => {
      const handler = (e: KeyboardEvent) => {
        if ((e.metaKey || e.ctrlKey) && (e.key === "k" || e.key === "K")) {
          // Only hijack when the canvas panel is actually in the document — the
          // dock is only rendered in that case, so this is safe.
          e.preventDefault()
          setExpanded(true)
          setTimeout(() => inputRef.current?.focus(), 30)
        }
      }
      window.addEventListener("keydown", handler)
      return () => window.removeEventListener("keydown", handler)
    }, [])

    // Compute target (selection vs document) at the moment we send.
    const snapshotTarget = useCallback((): {
      range: "selection" | "document"
      text: string
      preview: string
    } => {
      const h = editorRef.current
      if (!h) return { range: "document", text: "", preview: "" }
      const hasSel = h.hasSelection()
      const text = hasSel ? h.getSelectionText() : h.getDocText()
      const preview = text.length > 140 ? text.slice(0, 140) + "…" : text
      return { range: hasSel ? "selection" : "document", text, preview }
    }, [editorRef])

    // ── Stream from /ai/execute, appending tokens to `aiMsgId`. ──────────────
    const runAI = useCallback(async (
      instruction: string,
      userLabel: string,
      target: ReturnType<typeof snapshotTarget>,
    ) => {
      if (!channelId) {
        toast.error("Missing channel context for AI request")
        return
      }

      const userMsgId = `u-${Date.now()}`
      const aiMsgId = `a-${Date.now()}`

      setMessages(prev => [
        ...prev,
        {
          id: userMsgId,
          role: "user",
          text: userLabel,
          targetRange: target.range,
          targetPreview: target.preview,
        },
        {
          id: aiMsgId,
          role: "assistant",
          text: "",
          instruction,
          targetRange: target.range,
          targetPreview: target.preview,
          isStreaming: true,
        },
      ])
      setStreamingId(aiMsgId)
      setExpanded(true)

      abortCtrlRef.current?.abort()
      const ctrl = new AbortController()
      abortCtrlRef.current = ctrl

      const prompt = target.text
        ? `${instruction}\n\n---\n\n${target.text}\n\n---\n\n` +
          `Return ONLY the transformed text. Do not include explanations, preambles, or markdown fencing.`
        : `${instruction}\n\nReturn ONLY the requested text. Do not include explanations, preambles, or markdown fencing.`

      try {
        const res = await fetch(`${API_BASE_URL}/ai/execute`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ prompt, channel_id: channelId }),
          signal: ctrl.signal,
        })
        if (!res.ok || !res.body) throw new Error(`AI request failed (${res.status})`)

        const reader = res.body.getReader()
        const decoder = new TextDecoder()
        let buffer = ""
        let acc = ""

        while (true) {
          const { value, done } = await reader.read()
          if (done) break
          buffer += decoder.decode(value, { stream: true })
          const frames = buffer.split("\n\n")
          buffer = frames.pop() ?? ""
          for (const frame of frames) {
            const lines = frame.split("\n")
            let evt = "message"
            let data = ""
            for (const line of lines) {
              if (line.startsWith("event:")) evt = line.slice(6).trim()
              else if (line.startsWith("data:")) data += line.slice(5).trim()
            }
            if (!data) continue
            try {
              const parsed = JSON.parse(data)
              if (evt === "chunk" || evt === "message" || evt === "token") {
                const chunk =
                  (typeof parsed.text === "string" && parsed.text) ||
                  (typeof parsed.content === "string" && parsed.content) ||
                  (typeof parsed.delta === "string" && parsed.delta) ||
                  ""
                if (chunk) {
                  acc += chunk
                  setMessages(prev =>
                    prev.map(m => m.id === aiMsgId ? { ...m, text: acc } : m),
                  )
                }
              } else if (evt === "error") {
                throw new Error(parsed.message || "AI streaming error")
              }
              // "start", "reasoning", "done", "conversation" are intentionally ignored.
            } catch {
              acc += data
              setMessages(prev =>
                prev.map(m => m.id === aiMsgId ? { ...m, text: acc } : m),
              )
            }
          }
        }
      } catch (err: unknown) {
        const isAbort = err instanceof DOMException && err.name === "AbortError"
        if (!isAbort) {
          console.error("Canvas AI dock failed:", err)
          toast.error(err instanceof Error ? err.message : "AI request failed")
          setMessages(prev =>
            prev.map(m =>
              m.id === aiMsgId
                ? { ...m, text: m.text || "⚠︎ AI request failed. Try Retry.", isStreaming: false, errored: true }
                : m,
            ),
          )
        }
      } finally {
        setStreamingId(null)
        abortCtrlRef.current = null
        setMessages(prev =>
          prev.map(m => m.id === aiMsgId ? { ...m, isStreaming: false } : m),
        )
      }
    }, [channelId])

    // ── Send the current input. ──────────────────────────────────────────────
    const handleSend = useCallback(() => {
      const trimmed = input.trim()
      if (!trimmed) return

      // Slash-command expansion: `/cmd optional trailing text` → that command's
      // instruction, with the trailing text appended as extra guidance.
      const slashMatch = SLASH_COMMANDS.find(c =>
        trimmed === c.cmd || trimmed.startsWith(c.cmd + " "),
      )
      let instruction: string
      let label: string
      if (slashMatch) {
        const arg = trimmed.slice(slashMatch.cmd.length).trim()
        instruction = arg
          ? `${slashMatch.instruction}\n\nAdditional user instructions: ${arg}`
          : slashMatch.instruction
        label = arg ? `${slashMatch.label} — ${arg}` : slashMatch.label
      } else {
        instruction = trimmed
        label = trimmed
      }

      const target = snapshotTarget()
      runAI(instruction, label, target)

      setInput("")
      setShowSlashMenu(false)
    }, [input, runAI, snapshotTarget])

    // ── Composer keydown. ────────────────────────────────────────────────────
    const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
      // Slash menu navigation takes priority.
      if (showSlashMenu && filteredCommands.length > 0) {
        if (e.key === "ArrowDown") {
          e.preventDefault()
          setSlashSelected(s => (s + 1) % filteredCommands.length)
          return
        }
        if (e.key === "ArrowUp") {
          e.preventDefault()
          setSlashSelected(s => (s - 1 + filteredCommands.length) % filteredCommands.length)
          return
        }
        if (e.key === "Tab") {
          e.preventDefault()
          const pick = filteredCommands[slashSelected]
          setInput(pick.cmd + " ")
          setShowSlashMenu(false)
          return
        }
        if (e.key === "Enter" && !e.shiftKey) {
          // Enter with a partial slash — if the input *is* exactly a slash command
          // (or "/cmd <arg>"), fall through to send. Otherwise pick the highlighted one.
          const exactMatch = SLASH_COMMANDS.some(c =>
            input.trim() === c.cmd || input.trim().startsWith(c.cmd + " "),
          )
          if (!exactMatch) {
            e.preventDefault()
            const pick = filteredCommands[slashSelected]
            setInput(pick.cmd + " ")
            setShowSlashMenu(false)
            return
          }
        }
        if (e.key === "Escape") {
          e.preventDefault()
          setShowSlashMenu(false)
          return
        }
      }

      // Send shortcuts.
      if (e.key === "Enter" && !e.shiftKey) {
        e.preventDefault()
        handleSend()
      }
    }

    // ── Per-message actions. ─────────────────────────────────────────────────
    const handleApply = (msgId: string) => {
      const msg = messages.find(m => m.id === msgId)
      if (!msg?.text.trim()) return
      const h = editorRef.current
      if (!h) return
      const html = plainTextToHtml(msg.text)
      if (msg.targetRange === "selection") {
        // We snapshotted the intent at send-time; apply to current selection.
        h.applyHtmlToSelection(html)
      } else {
        h.applyHtmlToDoc(html)
      }
      setMessages(prev => prev.map(m => m.id === msgId ? { ...m, applied: true } : m))
      toast.success(
        msg.targetRange === "selection"
          ? "Applied to selection"
          : "Applied to document",
      )
    }

    const handleInsert = (msgId: string) => {
      const msg = messages.find(m => m.id === msgId)
      if (!msg?.text.trim()) return
      const h = editorRef.current
      if (!h) return
      h.insertHtmlAtCursor(plainTextToHtml(msg.text))
      setMessages(prev => prev.map(m => m.id === msgId ? { ...m, applied: true } : m))
      toast.success("Inserted at cursor")
    }

    const handleCopy = async (text: string) => {
      try {
        await navigator.clipboard.writeText(text)
        toast.success("Copied to clipboard")
      } catch {
        toast.error("Copy failed")
      }
    }

    const handleRetry = (msgId: string) => {
      const msg = messages.find(m => m.id === msgId)
      if (!msg?.instruction) return
      const idx = messages.findIndex(m => m.id === msgId)
      const userMsg = idx > 0 ? messages[idx - 1] : null
      // Fresh snapshot — selection may have changed since the original send.
      const target = snapshotTarget()
      runAI(msg.instruction, userMsg?.text ?? "Retry", target)
    }

    const handleStop = () => abortCtrlRef.current?.abort()

    const handleClear = () => {
      setMessages([])
      setExpanded(false)
    }

    // ── Render ───────────────────────────────────────────────────────────────
    const showHistory = expanded && messages.length > 0

    return (
      <div className="border-t bg-gradient-to-b from-white to-purple-50/30 dark:from-[#1a1d21] dark:to-purple-950/10 shrink-0">
        {/* Chat history — only while expanded & we have something to show. */}
        {showHistory && (
          <div className="max-h-[280px] overflow-y-auto p-3 space-y-3 border-b border-purple-500/10">
            {messages.map(m => (
              <ChatBubble
                key={m.id}
                message={m}
                onApply={() => handleApply(m.id)}
                onInsert={() => handleInsert(m.id)}
                onCopy={() => handleCopy(m.text)}
                onRetry={() => handleRetry(m.id)}
                onStop={handleStop}
                isStreaming={streamingId === m.id}
              />
            ))}
            <div ref={chatEndRef} />
          </div>
        )}

        {/* Composer. */}
        <div className="p-2 relative">
          {showSlashMenu && filteredCommands.length > 0 && (
            <div className="absolute bottom-full left-2 right-2 mb-1.5 z-20 bg-popover border rounded-lg shadow-lg max-h-72 overflow-y-auto">
              <div className="px-3 py-1.5 text-[9px] font-black uppercase tracking-widest text-muted-foreground border-b bg-muted/30">
                Slash commands · ↑↓ navigate · Tab select
              </div>
              {filteredCommands.map((c, i) => (
                <button
                  key={c.cmd}
                  onMouseEnter={() => setSlashSelected(i)}
                  onClick={() => {
                    setInput(c.cmd + " ")
                    setShowSlashMenu(false)
                    inputRef.current?.focus()
                  }}
                  className={cn(
                    "w-full text-left px-3 py-1.5 text-xs flex items-center gap-3 transition-colors",
                    i === slashSelected && "bg-purple-500/10",
                  )}
                >
                  <span className="text-sm w-4 text-center">{c.glyph}</span>
                  <code className="text-[10px] font-bold text-purple-600 dark:text-purple-300 bg-purple-500/10 px-1.5 py-0.5 rounded min-w-[90px]">
                    {c.cmd}
                  </code>
                  <span className="truncate font-medium">{c.label}</span>
                </button>
              ))}
            </div>
          )}

          <div className="flex items-end gap-2 bg-white dark:bg-[#19171d] border rounded-lg px-2.5 py-1.5 focus-within:border-purple-500/50 focus-within:ring-2 focus-within:ring-purple-500/15 transition-shadow">
            <Sparkles className="w-4 h-4 text-purple-500 shrink-0 mt-1" />
            <textarea
              ref={inputRef}
              value={input}
              onChange={e => setInput(e.target.value)}
              onKeyDown={handleKeyDown}
              onFocus={() => messages.length > 0 && setExpanded(true)}
              placeholder={
                streamingId
                  ? "AI is thinking…"
                  : messages.length === 0
                    ? `Ask AI to edit${artifactTitle ? ` "${artifactTitle}"` : ""}… try / for commands`
                    : "Follow up, or try / for commands"
              }
              rows={1}
              className="flex-1 text-xs resize-none bg-transparent focus:outline-none max-h-32 min-h-[22px] leading-relaxed py-1"
              disabled={!!streamingId}
            />
            <div className="flex items-center gap-1 shrink-0">
              {messages.length > 0 && (
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-7 w-7 text-muted-foreground"
                  onClick={() => setExpanded(v => !v)}
                  title={expanded ? "Collapse chat" : "Expand chat"}
                >
                  {expanded
                    ? <ChevronDown className="w-3.5 h-3.5" />
                    : <ChevronUp className="w-3.5 h-3.5" />}
                </Button>
              )}
              {messages.length > 0 && !streamingId && (
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-7 w-7 text-muted-foreground hover:text-red-600"
                  onClick={handleClear}
                  title="Clear chat history"
                >
                  <Trash2 className="w-3.5 h-3.5" />
                </Button>
              )}
              {streamingId ? (
                <Button
                  size="icon"
                  variant="outline"
                  className="h-7 w-7"
                  onClick={handleStop}
                  title="Stop"
                >
                  <X className="w-3.5 h-3.5" />
                </Button>
              ) : (
                <Button
                  size="icon"
                  onClick={handleSend}
                  disabled={!input.trim()}
                  className="h-7 w-7 bg-purple-600 hover:bg-purple-700 disabled:bg-muted disabled:text-muted-foreground text-white"
                  title="Send (Enter)"
                >
                  <Send className="w-3.5 h-3.5" />
                </Button>
              )}
            </div>
          </div>

          {/* Target chip + keyboard hint row. */}
          <div className="flex items-center justify-between mt-1 px-1 text-[10px] text-muted-foreground">
            <TargetChip editorRef={editorRef} />
            <div className="hidden sm:flex items-center gap-2">
              <span>
                <kbd className="px-1 py-0.5 bg-muted rounded text-[9px] font-mono">⌘K</kbd> focus
              </span>
              <span>
                <kbd className="px-1 py-0.5 bg-muted rounded text-[9px] font-mono">/</kbd> commands
              </span>
              <span>
                <kbd className="px-1 py-0.5 bg-muted rounded text-[9px] font-mono">↵</kbd> send
              </span>
            </div>
          </div>
        </div>
      </div>
    )
  },
)

// ── Chat bubble ──────────────────────────────────────────────────────────────

interface ChatBubbleProps {
  message: DockMessage
  onApply: () => void
  onInsert: () => void
  onCopy: () => void
  onRetry: () => void
  onStop: () => void
  isStreaming: boolean
}

function ChatBubble({ message, onApply, onInsert, onCopy, onRetry, onStop, isStreaming }: ChatBubbleProps) {
  if (message.role === "user") {
    return (
      <div className="flex justify-end">
        <div className="max-w-[75%] bg-purple-600 text-white rounded-lg rounded-tr-sm px-3 py-2 text-xs leading-relaxed shadow-sm">
          <p className="whitespace-pre-wrap break-words">{message.text}</p>
          {message.targetPreview && (
            <div className="mt-1.5 pt-1.5 border-t border-white/20 text-[10px] opacity-80 italic line-clamp-2">
              → {message.targetRange === "selection" ? "Selection" : "Whole document"}
              {message.targetPreview && `: "${message.targetPreview}"`}
            </div>
          )}
        </div>
      </div>
    )
  }

  // Assistant
  return (
    <div className="flex justify-start">
      <div className="max-w-[88%] w-full">
        <div className="flex items-center gap-1.5 mb-1">
          <Wand2 className="w-3 h-3 text-purple-500" />
          <span className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">
            AI
          </span>
          {isStreaming && <Loader2 className="w-3 h-3 animate-spin text-purple-500" />}
          {message.applied && (
            <span className="text-[9px] font-bold uppercase tracking-widest text-emerald-600 dark:text-emerald-400 flex items-center gap-0.5">
              <Check className="w-2.5 h-2.5" />
              Applied
            </span>
          )}
          {message.errored && !isStreaming && (
            <span className="text-[9px] font-bold uppercase tracking-widest text-red-500">
              Failed
            </span>
          )}
        </div>
        <div className={cn(
          "rounded-lg rounded-tl-sm border px-3 py-2 text-xs leading-relaxed bg-white dark:bg-[#19171d] shadow-sm",
          message.applied && "border-emerald-500/30 bg-emerald-50/40 dark:bg-emerald-900/10",
          message.errored && "border-red-500/30 bg-red-50/40 dark:bg-red-900/10",
          !message.applied && !message.errored && "border-purple-500/20",
        )}>
          {message.text ? (
            <p className="whitespace-pre-wrap break-words">{message.text}</p>
          ) : (
            <p className="italic text-muted-foreground">Thinking…</p>
          )}
        </div>
        {!isStreaming && message.text && !message.errored && (
          <div className="flex items-center flex-wrap gap-1 mt-1.5">
            {!message.applied && (
              <>
                <Button
                  size="sm"
                  variant="outline"
                  className="h-6 text-[10px] font-bold uppercase tracking-widest gap-1 border-emerald-500/40 text-emerald-700 dark:text-emerald-300 hover:bg-emerald-500/10"
                  onClick={onApply}
                >
                  <Check className="w-3 h-3" />
                  Apply to {message.targetRange === "selection" ? "selection" : "document"}
                </Button>
                <Button
                  size="sm"
                  variant="ghost"
                  className="h-6 text-[10px] font-bold uppercase tracking-widest gap-1"
                  onClick={onInsert}
                  title="Insert at cursor instead of replacing"
                >
                  Insert at cursor
                </Button>
              </>
            )}
            <Button
              size="sm"
              variant="ghost"
              className="h-6 text-[10px] gap-1 text-muted-foreground hover:text-foreground"
              onClick={onCopy}
              title="Copy to clipboard"
            >
              <Copy className="w-3 h-3" />
            </Button>
            <Button
              size="sm"
              variant="ghost"
              className="h-6 text-[10px] gap-1 text-muted-foreground hover:text-foreground"
              onClick={onRetry}
              title="Retry with the same instruction on current selection"
            >
              <RefreshCw className="w-3 h-3" />
            </Button>
          </div>
        )}
        {message.errored && !isStreaming && (
          <div className="flex items-center gap-1 mt-1.5">
            <Button
              size="sm"
              variant="outline"
              className="h-6 text-[10px] font-bold uppercase tracking-widest gap-1"
              onClick={onRetry}
            >
              <RefreshCw className="w-3 h-3" />
              Retry
            </Button>
          </div>
        )}
        {isStreaming && (
          <div className="flex items-center gap-1 mt-1.5">
            <Button
              size="sm"
              variant="outline"
              className="h-6 text-[10px] font-bold uppercase tracking-widest gap-1"
              onClick={onStop}
            >
              <X className="w-3 h-3" />
              Stop
            </Button>
          </div>
        )}
      </div>
    </div>
  )
}

// ── Target chip ──────────────────────────────────────────────────────────────
//
// Live indicator of what "Send" will rewrite — selection or whole doc. Polls
// the editor handle cheaply (TipTap doesn't expose a selection-change event
// to arbitrary siblings without extra plumbing, and a 400ms tick is fine for
// a static indicator).

function TargetChip({ editorRef }: { editorRef: React.RefObject<CanvasEditorHandle | null> }) {
  const [info, setInfo] = useState<{ hasSel: boolean; len: number }>({ hasSel: false, len: 0 })

  useEffect(() => {
    const poll = () => {
      const h = editorRef.current
      if (!h) return
      const hasSel = h.hasSelection()
      const text = hasSel ? h.getSelectionText() : h.getDocText()
      setInfo(prev => {
        if (prev.hasSel === hasSel && prev.len === text.length) return prev
        return { hasSel, len: text.length }
      })
    }
    poll()
    const interval = setInterval(poll, 400)
    return () => clearInterval(interval)
  }, [editorRef])

  return (
    <span className="flex items-center gap-1.5">
      <span className={cn(
        "inline-block w-1.5 h-1.5 rounded-full",
        info.hasSel ? "bg-purple-500" : "bg-blue-400",
      )} />
      <span>
        Applies to{" "}
        <strong className="text-foreground/70">
          {info.hasSel
            ? `selection (${info.len.toLocaleString()} chars)`
            : `whole document (${info.len.toLocaleString()} chars)`}
        </strong>
      </span>
    </span>
  )
}

// ── Helpers ──────────────────────────────────────────────────────────────────

function plainTextToHtml(text: string): string {
  return text
    .split(/\n{2,}/)
    .map(block => `<p>${escapeHtml(block).replace(/\n/g, "<br/>")}</p>`)
    .join("")
}

function escapeHtml(s: string): string {
  return s
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
}
