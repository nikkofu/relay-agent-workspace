"use client"

// ── Canvas AI Dock ───────────────────────────────────────────────────────────
//
// A ChatGPT / Gemini Canvas–style AI chat rail that lives inside the Canvas
// panel while a document-type artifact is being edited. Replaces the previous
// modal AI-edit dialog so the user never loses cursor / selection context.
//
// Features
//   • Persistent composer with streaming SSE preview (reuses `/api/v1/ai/execute`,
//     `event: chunk` / `{text}`) — plus `event: reasoning` is captured and
//     rendered as a collapsible "Thinking…" block with its own scroll so users
//     can follow the model's reasoning.
//   • Slash-command menu (`/expand`, `/shorten`, `/rephrase`, `/fix`,
//     `/formal`, `/casual`, `/bullets`, `/outline`, `/summary`, `/continue`,
//     `/translate-en`, `/translate-zh`).
//   • **Selection-reference pill** — when the user has text selected in the
//     editor, the composer shows a ChatGPT-style quote card above the input
//     with a preview + ✕ to clear. Makes it obvious that the next run will
//     operate on just the quoted slice.
//   • **Post-apply confirmation** — after Apply / Insert at cursor, the
//     editor auto-focuses, selects the newly-inserted range, scrolls it into
//     view, and a Sonner toast offers a one-click **Undo** (6s) so users can
//     see + safely revert the change even though their original cursor
//     position is gone.
//   • **Two layouts** — `bottom` (compact bar under the editor, default) and
//     `rail` (full-height ~400 px sidebar). Toggleable; the parent persists the
//     preference in localStorage.
//   • Chat-bubble history with per-message actions: Apply / Insert / Copy /
//     Retry / Stop.
//   • Keyboard: ⌘K focuses, Enter sends, Shift+Enter newline, Esc closes
//     slash menu.

import {
  forwardRef, useCallback, useEffect, useImperativeHandle, useMemo, useRef, useState,
} from "react"
import {
  Sparkles, Send, X, Loader2, Check, RefreshCw, ChevronDown, ChevronUp,
  Copy, Wand2, Trash2, Brain, Quote, PanelRight, PanelBottom,
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"
import { API_BASE_URL } from "@/lib/constants"
import { toast } from "sonner"
import type { CanvasEditorHandle, EditorRange } from "./canvas-tiptap-editor"
import { parseAIStreamEvent } from "@/lib/ai-sidecar"
import { ToolTimeline, UsageChip } from "@/components/ai/ai-sidecar-blocks"

// ── Public handle ────────────────────────────────────────────────────────────

export interface CanvasAIDockHandle {
  /** Focus the composer input and expand the dock. */
  focusInput: () => void
}

// ── Layout mode ──────────────────────────────────────────────────────────────

export type CanvasAIDockLayout = "bottom" | "rail"

// ── Slash commands ───────────────────────────────────────────────────────────

interface SlashCommand {
  cmd: string
  label: string
  instruction: string
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
  text: string
  /** Only on assistant messages — the instruction sent to the LLM (Retry). */
  instruction?: string
  /** Accumulated reasoning tokens, kept as a flat string for legacy renderers.
   *  The canonical structured form is on `sidecar.reasoning`. */
  reasoning?: string
  /** Unified AI Side-Channel Contract sidecar accumulated from the SSE
   *  stream (`reasoning` / `tool_call` / `usage` kinds). Drives the shared
   *  Reasoning panel + Tool timeline + Usage chip components. */
  sidecar?: import("@/lib/ai-sidecar").AISidecar
  /** Target scope captured at send-time. */
  targetRange?: "selection" | "document"
  targetPreview?: string
  /** Full target text captured at send-time (for Retry with identical scope). */
  capturedTargetText?: string
  /** Editor range where this message's content was applied (for the Undo
   *  toast to know what to re-select on undo). */
  appliedRange?: EditorRange
  isStreaming?: boolean
  applied?: boolean
  errored?: boolean
}

// ── Props ────────────────────────────────────────────────────────────────────

interface CanvasAIDockProps {
  editorRef: React.RefObject<CanvasEditorHandle | null>
  channelId: string
  artifactTitle?: string
  /** Current layout; defaults to "bottom". */
  layout?: CanvasAIDockLayout
  /** Called when the user toggles layout from the dock header. */
  onLayoutChange?: (layout: CanvasAIDockLayout) => void
}

// ── Component ────────────────────────────────────────────────────────────────

export const CanvasAIDock = forwardRef<CanvasAIDockHandle, CanvasAIDockProps>(
  function CanvasAIDock({ editorRef, channelId, artifactTitle, layout = "bottom", onLayoutChange }, ref) {
    const [input, setInput] = useState("")
    const [expanded, setExpanded] = useState(layout === "rail")
    const [messages, setMessages] = useState<DockMessage[]>([])

    const [showSlashMenu, setShowSlashMenu] = useState(false)
    const [slashFilter, setSlashFilter] = useState("")
    const [slashSelected, setSlashSelected] = useState(0)

    const [streamingId, setStreamingId] = useState<string | null>(null)
    const abortCtrlRef = useRef<AbortController | null>(null)
    const inputRef = useRef<HTMLTextAreaElement>(null)
    const chatEndRef = useRef<HTMLDivElement>(null)

    // Cached target info — polled from the editor so the selection pill and
    // the status chip can re-render as the user changes the selection.
    const [targetInfo, setTargetInfo] = useState<{
      hasSel: boolean
      selText: string
      selPreview: string
      docLen: number
    }>({ hasSel: false, selText: "", selPreview: "", docLen: 0 })

    useEffect(() => {
      const poll = () => {
        const h = editorRef.current
        if (!h) return
        const hasSel = h.hasSelection()
        const selText = hasSel ? h.getSelectionText() : ""
        const docText = h.getDocText()
        const selPreview = selText.length > 180 ? selText.slice(0, 180) + "…" : selText
        setTargetInfo(prev => {
          if (
            prev.hasSel === hasSel &&
            prev.selText.length === selText.length &&
            prev.docLen === docText.length &&
            prev.selPreview === selPreview
          ) return prev
          return { hasSel, selText, selPreview, docLen: docText.length }
        })
      }
      poll()
      const interval = setInterval(poll, 350)
      return () => clearInterval(interval)
    }, [editorRef])

    // Expose `focusInput` to parent.
    useImperativeHandle(ref, () => ({
      focusInput: () => {
        setExpanded(true)
        setTimeout(() => inputRef.current?.focus(), 30)
      },
    }), [])

    // Layout switching side-effect: ensure the dock is expanded in rail mode
    // so the conversation is actually visible as a full-height column.
    useEffect(() => {
      if (layout === "rail") setExpanded(true)
    }, [layout])

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
          e.preventDefault()
          setExpanded(true)
          setTimeout(() => inputRef.current?.focus(), 30)
        }
      }
      window.addEventListener("keydown", handler)
      return () => window.removeEventListener("keydown", handler)
    }, [])

    // Snapshot the editor target (selection vs document) at send-time.
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
          reasoning: "",
          instruction,
          targetRange: target.range,
          targetPreview: target.preview,
          capturedTargetText: target.text,
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
        // On non-OK (502/503/etc) the server replies with a JSON body like
        // `{"error":"upstream returned 401: ..."}`. Surface that real message
        // to the user instead of a generic status code so they can actually
        // diagnose the problem (bad API key, model unavailable, rate limit…).
        if (!res.ok) {
          const detail = await readErrorBody(res)
          throw new Error(formatBackendError(res.status, detail))
        }
        if (!res.body) throw new Error(`AI request failed (${res.status}) — empty body`)

        const reader = res.body.getReader()
        const decoder = new TextDecoder()
        let buffer = ""
        let accText = ""
        let accReasoning = ""
        // Unified contract sidecar accumulator. We mutate this in-place as
        // `reasoning` / `tool_call` / `usage` events arrive and re-render
        // the assistant bubble with a fresh shallow copy each time so React
        // sees the change.
        const sidecar: import("@/lib/ai-sidecar").AISidecar = {}
        const toolById = new Map<string, import("@/lib/ai-sidecar").AIToolCall>()

        const pushSidecar = () => {
          const snapshot: import("@/lib/ai-sidecar").AISidecar = {
            reasoning: sidecar.reasoning,
            tool_calls: toolById.size > 0 ? Array.from(toolById.values()) : sidecar.tool_calls,
            usage: sidecar.usage,
          }
          setMessages(prev =>
            prev.map(m => m.id === aiMsgId ? { ...m, sidecar: snapshot } : m),
          )
        }

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
            // Use the unified parser that understands BOTH the new normative
            // `{ kind, message_id, payload }` envelope (Gemini v0.6.51+) and
            // the legacy `event: chunk|reasoning|error` + `{ text }` form
            // older backends still emit. Returns null for heartbeats.
            const parsedEvt = parseAIStreamEvent(evt, data)
            if (!parsedEvt) {
              // Heartbeat / unknown frame — if it's plain text-only with no
              // recognised wrapper, append to the answer so we don't drop
              // bytes from a misconfigured backend.
              if (evt === "chunk" || evt === "message") {
                try {
                  const j = JSON.parse(data)
                  if (typeof j.text === "string" && j.text) {
                    accText += j.text
                    setMessages(prev =>
                      prev.map(m => m.id === aiMsgId ? { ...m, text: accText } : m),
                    )
                  }
                } catch {
                  accText += data
                  setMessages(prev =>
                    prev.map(m => m.id === aiMsgId ? { ...m, text: accText } : m),
                  )
                }
              }
              continue
            }
            if (parsedEvt.kind === "answer" && parsedEvt.text) {
              accText += parsedEvt.text
              setMessages(prev =>
                prev.map(m => m.id === aiMsgId ? { ...m, text: accText } : m),
              )
            } else if (parsedEvt.kind === "reasoning" && parsedEvt.text) {
              accReasoning += parsedEvt.text
              // Maintain the legacy flat `reasoning` string alongside the
              // canonical structured form so anything still consuming the
              // pre-contract field keeps working.
              sidecar.reasoning = { summary: accReasoning, segments: [] }
              setMessages(prev =>
                prev.map(m => m.id === aiMsgId
                  ? { ...m, reasoning: accReasoning, sidecar: { ...sidecar, tool_calls: toolById.size ? Array.from(toolById.values()) : sidecar.tool_calls } }
                  : m),
              )
            } else if (parsedEvt.kind === "tool_call" && parsedEvt.tool) {
              // Merge by id so streaming `running → success` updates land
              // on the same row instead of duplicating.
              const tc = parsedEvt.tool
              const id = tc.id || `tc-${toolById.size}`
              const prev = toolById.get(id)
              const merged = {
                id,
                name: tc.name ?? prev?.name ?? "tool",
                status: tc.status ?? prev?.status ?? "running",
                input_summary: tc.input_summary ?? prev?.input_summary,
                output_summary: tc.output_summary ?? prev?.output_summary,
                duration_ms: tc.duration_ms ?? prev?.duration_ms,
              }
              toolById.set(id, merged)
              pushSidecar()
            } else if (parsedEvt.kind === "usage" && parsedEvt.usage) {
              sidecar.usage = parsedEvt.usage
              pushSidecar()
            } else if (parsedEvt.kind === "error") {
              throw new Error(parsedEvt.error || "AI streaming error")
            }
          }
        }
      } catch (err: unknown) {
        const isAbort = err instanceof DOMException && err.name === "AbortError"
        if (!isAbort) {
          console.error("Canvas AI dock failed:", err)
          const msg = err instanceof Error ? err.message : "AI request failed"
          // Longer-lived toast so users can actually read upstream errors
          // like "provider api key is empty" or "upstream returned 429".
          toast.error("AI request failed", {
            description: msg,
            duration: 12000,
          })
          setMessages(prev =>
            prev.map(m =>
              m.id === aiMsgId
                ? {
                    ...m,
                    // Keep any partial assistant text we received, but fall
                    // back to the diagnostic so the bubble isn't just "failed".
                    text: m.text || `⚠︎ ${msg}`,
                    isStreaming: false,
                    errored: true,
                  }
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

    // ── Send ─────────────────────────────────────────────────────────────────
    const handleSend = useCallback(() => {
      const trimmed = input.trim()
      if (!trimmed) return

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

    // ── Composer keydown ─────────────────────────────────────────────────────
    const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
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

      if (e.key === "Enter" && !e.shiftKey) {
        e.preventDefault()
        handleSend()
      }
    }

    // ── Post-apply feedback ──────────────────────────────────────────────────
    //
    // After Apply / Insert we:
    //  1. Refocus the editor (the button click stole focus).
    //  2. Select the newly-inserted range so the browser shows its highlight.
    //  3. Scroll that range into view.
    //  4. Fire a Sonner toast with an Undo action (calls the editor's own
    //     undo history so the transform is a single ⌘Z step).
    //
    // This is important because the user's original selection + cursor is
    // gone after the mutation, so without a visual confirmation they'd be
    // staring at a rewritten doc with no idea what changed where.
    const showAppliedToast = (scopeLabel: string) => {
      toast.success(`${scopeLabel} applied`, {
        description: "The highlighted range shows what changed. ⌘Z to undo.",
        duration: 6500,
        action: {
          label: "Undo",
          onClick: () => {
            // TipTap's undo history captures our insertContent txn as one step.
            // We dispatch via the DOM since we don't have a direct reference to
            // editor.commands here (would be cleaner as an additional handle
            // method, but the ⌘Z path is equivalent + already wired via the
            // toolbar's Undo button).
            const evt = new KeyboardEvent("keydown", {
              key: "z",
              code: "KeyZ",
              metaKey: true,
              bubbles: true,
              cancelable: true,
            })
            document.dispatchEvent(evt)
            // Best-effort: the editor has its own keydown handler, and the
            // toolbar's Undo button also wires up — either path undoes the
            // last transaction.
            editorRef.current?.focus()
          },
        },
      })
    }

    const handleApply = (msgId: string) => {
      const msg = messages.find(m => m.id === msgId)
      if (!msg?.text.trim()) return
      const h = editorRef.current
      if (!h) return
      const html = plainTextToHtml(msg.text)
      const range = msg.targetRange === "selection"
        ? h.applyHtmlToSelection(html)
        : h.applyHtmlToDoc(html)
      setMessages(prev => prev.map(m =>
        m.id === msgId ? { ...m, applied: true, appliedRange: range ?? undefined } : m,
      ))
      if (range) h.highlightRange(range)
      showAppliedToast(msg.targetRange === "selection" ? "Selection" : "Document")
    }

    const handleInsert = (msgId: string) => {
      const msg = messages.find(m => m.id === msgId)
      if (!msg?.text.trim()) return
      const h = editorRef.current
      if (!h) return
      const range = h.insertHtmlAtCursor(plainTextToHtml(msg.text))
      setMessages(prev => prev.map(m =>
        m.id === msgId ? { ...m, applied: true, appliedRange: range ?? undefined } : m,
      ))
      if (range) h.highlightRange(range)
      showAppliedToast("Insertion at cursor")
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
      const target = snapshotTarget()
      runAI(msg.instruction, userMsg?.text ?? "Retry", target)
    }

    const handleStop = () => abortCtrlRef.current?.abort()

    const handleClear = () => {
      setMessages([])
      if (layout === "bottom") setExpanded(false)
    }

    /** Collapse the current selection in the editor (the "clear quote" button
     *  on the selection-reference pill). */
    const handleClearSelection = () => {
      editorRef.current?.clearSelection()
      // Re-focus the composer so the user can keep typing.
      setTimeout(() => inputRef.current?.focus(), 30)
    }

    // ── Render ───────────────────────────────────────────────────────────────
    const isRail = layout === "rail"
    const showHistory = expanded && messages.length > 0

    return (
      <div className={cn(
        "flex flex-col bg-gradient-to-b from-white to-purple-50/30 dark:from-[#1a1d21] dark:to-purple-950/10 shrink-0",
        isRail
          ? "h-full w-full border-l"
          : "border-t",
      )}>
        {/* Rail header — only in rail mode. */}
        {isRail && (
          <div className="h-9 px-3 border-b flex items-center justify-between shrink-0 bg-muted/20">
            <div className="flex items-center gap-1.5 text-[10px] font-black uppercase tracking-widest text-purple-700 dark:text-purple-300">
              <Wand2 className="w-3 h-3" />
              AI Assistant
            </div>
            <div className="flex items-center gap-1">
              {messages.length > 0 && !streamingId && (
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-6 w-6 text-muted-foreground hover:text-red-600"
                  onClick={handleClear}
                  title="Clear chat history"
                >
                  <Trash2 className="w-3 h-3" />
                </Button>
              )}
              <Button
                variant="ghost"
                size="icon"
                className="h-6 w-6 text-muted-foreground"
                onClick={() => onLayoutChange?.("bottom")}
                title="Switch to bottom dock layout"
              >
                <PanelBottom className="w-3 h-3" />
              </Button>
            </div>
          </div>
        )}

        {/* Chat history */}
        {(showHistory || isRail) && (
          <div className={cn(
            "overflow-y-auto p-3 space-y-3",
            isRail
              ? "flex-1 min-h-0"
              : "max-h-[280px] border-b border-purple-500/10",
          )}>
            {messages.length === 0 ? (
              <div className="h-full flex flex-col items-center justify-center text-center gap-2 py-8 px-4">
                <Wand2 className="w-5 h-5 text-purple-400" />
                <p className="text-[11px] font-bold text-muted-foreground max-w-xs leading-relaxed">
                  Ask AI to rewrite, expand, or translate any part of your canvas.
                  Select text first to scope it, or leave blank for the whole doc.
                </p>
                <p className="text-[10px] text-muted-foreground/70">
                  Try <code className="px-1 py-0.5 bg-muted rounded font-mono">/expand</code>{" "}
                  or <code className="px-1 py-0.5 bg-muted rounded font-mono">/continue</code>
                </p>
              </div>
            ) : (
              messages.map(m => (
                <ChatBubble
                  key={m.id}
                  message={m}
                  onApply={() => handleApply(m.id)}
                  onInsert={() => handleInsert(m.id)}
                  onCopy={() => handleCopy(m.text)}
                  onRetry={() => handleRetry(m.id)}
                  onStop={handleStop}
                  onReSelect={() =>
                    m.appliedRange && editorRef.current?.highlightRange(m.appliedRange)
                  }
                  isStreaming={streamingId === m.id}
                />
              ))
            )}
            <div ref={chatEndRef} />
          </div>
        )}

        {/* Composer */}
        <div className="p-2 relative shrink-0">
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

          {/* Selection-reference pill — ChatGPT-style quote card above input. */}
          {targetInfo.hasSel && (
            <div className="mb-1.5 rounded-lg border-l-4 border-purple-500 bg-purple-500/10 dark:bg-purple-500/15 pl-2.5 pr-1.5 py-1.5 flex items-start gap-2 text-[11px]">
              <Quote className="w-3 h-3 text-purple-600 dark:text-purple-300 shrink-0 mt-0.5" />
              <div className="min-w-0 flex-1">
                <div className="flex items-center gap-1.5 mb-0.5">
                  <span className="text-[9px] font-black uppercase tracking-widest text-purple-700 dark:text-purple-300">
                    Referencing selection
                  </span>
                  <span className="text-[9px] text-muted-foreground">
                    · {targetInfo.selText.length.toLocaleString()} chars
                  </span>
                </div>
                <p className="italic leading-snug text-foreground/75 line-clamp-2 break-words">
                  “{targetInfo.selPreview}”
                </p>
              </div>
              <Button
                variant="ghost"
                size="icon"
                className="h-5 w-5 shrink-0 text-muted-foreground hover:text-foreground"
                onClick={handleClearSelection}
                title="Clear selection (rewrite whole document)"
              >
                <X className="w-3 h-3" />
              </Button>
            </div>
          )}

          <div className="flex items-end gap-2 bg-white dark:bg-[#19171d] border rounded-lg px-2.5 py-1.5 focus-within:border-purple-500/50 focus-within:ring-2 focus-within:ring-purple-500/15 transition-shadow">
            {/* Icon aligns to first text line, NOT the growing-textarea bottom. */}
            <Sparkles className="w-4 h-4 text-purple-500 shrink-0 self-start mt-[7px]" />
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
              {!isRail && messages.length > 0 && (
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
              {!isRail && onLayoutChange && (
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-7 w-7 text-muted-foreground"
                  onClick={() => onLayoutChange("rail")}
                  title="Open as side rail (full-height)"
                >
                  <PanelRight className="w-3.5 h-3.5" />
                </Button>
              )}
              {!isRail && messages.length > 0 && !streamingId && (
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

          {/* Target chip + keyboard hint row */}
          <div className="flex items-center justify-between mt-1 px-1 text-[10px] text-muted-foreground">
            <TargetChip targetInfo={targetInfo} />
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
  onReSelect: () => void
  isStreaming: boolean
}

function ChatBubble({
  message, onApply, onInsert, onCopy, onRetry, onStop, onReSelect, isStreaming,
}: ChatBubbleProps) {
  if (message.role === "user") {
    return (
      <div className="flex justify-end">
        <div className="max-w-[85%] bg-purple-600 text-white rounded-lg rounded-tr-sm px-3 py-2 text-xs leading-relaxed shadow-sm">
          <p className="whitespace-pre-wrap break-words">{message.text}</p>
          {message.targetPreview && (
            <div className="mt-1.5 pt-1.5 border-t border-white/20 text-[10px] opacity-80 italic line-clamp-2 break-words">
              → {message.targetRange === "selection" ? "Selection" : "Whole document"}
              : &ldquo;{message.targetPreview}&rdquo;
            </div>
          )}
        </div>
      </div>
    )
  }

  // Assistant
  return (
    <div className="flex justify-start">
      <div className="max-w-[90%] w-full">
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

        {/* Reasoning / Thinking section. Auto-expanded while streaming so the
            user can follow along, collapsed by default after the final answer
            lands (click to toggle). The custom amber-themed `ReasoningBlock`
            below is the canvas-specific look-and-feel. The shared
            `ToolTimeline` and `UsageChip` come straight from the unified
            contract. */}
        {message.reasoning && (
          <ReasoningBlock
            reasoning={message.reasoning}
            isStreaming={isStreaming}
          />
        )}

        {/* Tool timeline — shown the moment the first `tool_call` event lands. */}
        {message.sidecar?.tool_calls && message.sidecar.tool_calls.length > 0 && (
          <div className="mb-1.5">
            <ToolTimeline toolCalls={message.sidecar.tool_calls} />
          </div>
        )}

        {/* Final answer bubble */}
        <div className={cn(
          "rounded-lg rounded-tl-sm border px-3 py-2 text-xs leading-relaxed bg-white dark:bg-[#19171d] shadow-sm",
          message.applied && "border-emerald-500/30 bg-emerald-50/40 dark:bg-emerald-900/10",
          message.errored && "border-red-500/30 bg-red-50/40 dark:bg-red-900/10",
          !message.applied && !message.errored && "border-purple-500/20",
        )}>
          {message.text ? (
            <p className="whitespace-pre-wrap break-words">{message.text}</p>
          ) : message.reasoning ? (
            <p className="italic text-muted-foreground">Composing the answer…</p>
          ) : (
            <p className="italic text-muted-foreground">Thinking…</p>
          )}
        </div>

        {/* Actions */}
        {!isStreaming && message.text && !message.errored && (
          <div className="flex items-center flex-wrap gap-1 mt-1.5">
            {!message.applied ? (
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
            ) : (
              <Button
                size="sm"
                variant="outline"
                className="h-6 text-[10px] font-bold uppercase tracking-widest gap-1 border-purple-500/40 text-purple-700 dark:text-purple-300 hover:bg-purple-500/10"
                onClick={onReSelect}
                title="Re-highlight the range in the editor where this was applied"
              >
                Show in canvas
              </Button>
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
        {/* Usage chip from the unified contract — only renders when the
            backend supplies a real `usage` payload. The 4-char heuristic is
            deliberately not used here because the canvas dock has no need
            to display an estimated count when the model didn't cost anything. */}
        {message.sidecar?.usage && !isStreaming && (
          <div className="flex items-center gap-1 mt-1 px-1">
            <UsageChip usage={message.sidecar.usage} />
          </div>
        )}
      </div>
    </div>
  )
}

// ── Reasoning / Thinking block ───────────────────────────────────────────────
//
// Collapsible panel that shows the model's chain-of-thought tokens captured
// from `event: reasoning`. Auto-expands while the assistant is still
// streaming so users can follow along, then collapses to just the header
// once the final answer lands.

function ReasoningBlock({ reasoning, isStreaming }: { reasoning: string; isStreaming: boolean }) {
  const [open, setOpen] = useState(isStreaming)
  const scrollerRef = useRef<HTMLDivElement>(null)

  // When streaming completes, default to collapsed (but respect a user's
  // explicit toggle made during streaming).
  const prevStreaming = useRef(isStreaming)
  useEffect(() => {
    if (prevStreaming.current && !isStreaming) {
      setOpen(false)
    }
    prevStreaming.current = isStreaming
  }, [isStreaming])

  // Auto-stick-to-bottom while streaming so users see newest tokens.
  useEffect(() => {
    if (!open) return
    const el = scrollerRef.current
    if (!el) return
    el.scrollTop = el.scrollHeight
  }, [reasoning, open])

  const wordCount = reasoning.trim().split(/\s+/).filter(Boolean).length

  return (
    <div className="mb-1.5 rounded-lg border border-amber-400/30 bg-amber-50/40 dark:bg-amber-950/10">
      <button
        type="button"
        onClick={() => setOpen(v => !v)}
        className="w-full flex items-center gap-1.5 px-2.5 py-1.5 text-[10px] font-black uppercase tracking-widest text-amber-700 dark:text-amber-300 hover:bg-amber-500/5 rounded-lg"
      >
        <Brain className="w-3 h-3" />
        <span>Thinking</span>
        <span className="text-[9px] font-bold text-amber-600/70 dark:text-amber-300/70">
          · {wordCount.toLocaleString()} words
        </span>
        {isStreaming && (
          <Loader2 className="w-2.5 h-2.5 animate-spin text-amber-500" />
        )}
        <span className="ml-auto">
          {open
            ? <ChevronUp className="w-3 h-3" />
            : <ChevronDown className="w-3 h-3" />}
        </span>
      </button>
      {open && (
        <div
          ref={scrollerRef}
          className="max-h-48 overflow-y-auto px-3 pb-2 pt-0.5 text-[11px] text-amber-900/80 dark:text-amber-100/70 whitespace-pre-wrap leading-relaxed font-mono"
        >
          {reasoning.trim() || (
            <span className="italic text-muted-foreground">Reasoning tokens…</span>
          )}
        </div>
      )}
    </div>
  )
}

// ── Target chip ──────────────────────────────────────────────────────────────

function TargetChip({
  targetInfo,
}: {
  targetInfo: { hasSel: boolean; selText: string; docLen: number }
}) {
  const len = targetInfo.hasSel ? targetInfo.selText.length : targetInfo.docLen
  return (
    <span className="flex items-center gap-1.5">
      <span className={cn(
        "inline-block w-1.5 h-1.5 rounded-full",
        targetInfo.hasSel ? "bg-purple-500" : "bg-blue-400",
      )} />
      <span>
        Applies to{" "}
        <strong className="text-foreground/70">
          {targetInfo.hasSel
            ? `selection (${len.toLocaleString()} chars)`
            : `whole document (${len.toLocaleString()} chars)`}
        </strong>
      </span>
    </span>
  )
}

// ── Helpers ──────────────────────────────────────────────────────────────────

/** Read an error response body, preferring the `{error: "..."}` JSON shape
 *  the backend uses, and falling back to raw text. Never throws. */
async function readErrorBody(res: Response): Promise<string> {
  try {
    const text = await res.text()
    if (!text) return ""
    try {
      const parsed = JSON.parse(text)
      if (typeof parsed?.error === "string") return parsed.error
      if (typeof parsed?.message === "string") return parsed.message
      return text
    } catch {
      return text
    }
  } catch {
    return ""
  }
}

/** Build a user-facing error string from status + backend detail, with
 *  heuristic hints for the most common configuration / quota failures so the
 *  next step is obvious. */
function formatBackendError(status: number, detail: string): string {
  const d = detail.trim()
  const lower = d.toLowerCase()

  // Service-level: gateway not configured at all.
  if (status === 503 || lower.includes("ai gateway is not configured")) {
    return "AI gateway is not configured on the server. Set an LLM provider (OPENAI_API_KEY / GEMINI_API_KEY / …) and restart the API."
  }

  // Common upstream diagnostics we can act on.
  if (lower.includes("api key is empty") || lower.includes("api key")) {
    return `Provider API key is missing or invalid. Check the server's LLM provider config. (${d || status})`
  }
  if (lower.includes("provider not configured") || lower.includes("provider disabled")) {
    return `Requested LLM provider is not available. (${d || status})`
  }
  if (lower.includes("429") || lower.includes("rate limit") || lower.includes("quota")) {
    return `LLM provider rate-limited or over quota. Wait and retry. (${d || status})`
  }
  if (lower.includes("401") || lower.includes("403") || lower.includes("unauthorized")) {
    return `LLM provider rejected the API key (${status}). ${d}`
  }
  if (lower.includes("timeout") || lower.includes("deadline")) {
    return `LLM provider timed out. Try again or shorten the selection. (${d || status})`
  }

  // Fall back to the raw server error.
  if (d) return `${d} (HTTP ${status})`
  return `AI request failed (HTTP ${status})`
}

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
