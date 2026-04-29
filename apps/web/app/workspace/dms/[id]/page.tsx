"use client"

// ── /workspace/dms/[id] (single conversation) ───────────────────────────────
//
// Right-pane content of the WhatsApp-style two-pane layout. The DM list
// lives in `dms/layout.tsx`; this page only renders the active
// conversation. `useParams()` is called inside `<DMConversationContent />`
// which the layout already wraps with Suspense (Next 16 cacheComponents
// requirement that fixes request #3 — opening the Canvas from the DM
// editor used to crash with "Data that blocks navigation was accessed
// outside of <Suspense>").
//
// AI Assistant DMs (e.g. `dm-1`) get the ChatGPT-style affordances
// requested in #2:
//   • An always-visible "Thinking…" reasoning panel that streams the live
//     stream chunk into a collapsible block before the final answer
//     materialises.
//   • A tool-call timeline (Searched the web, Read file …) that surfaces
//     when the backend emits structured tool steps.
//   • A token-count footer (input / output / total) on every AI bubble,
//     estimated client-side until the backend exposes real telemetry.

import { useEffect, useMemo, useRef } from "react"
import { useParams, useRouter } from "next/navigation"
import { useMessageStore } from "@/stores/message-store"
import { useDMStore } from "@/stores/dm-store"
import { isAIUserLike, useUserStore } from "@/stores/user-store"
import { MessageComposer } from "@/components/message/message-composer"
import { TypingIndicator } from "@/components/message/typing-indicator"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import {
  ArrowLeft, Sparkles, Hash, RotateCcw, Copy as CopyIcon, Zap,
} from "lucide-react"
import { formatDistanceToNow } from "date-fns"
import { cn } from "@/lib/utils"
import { UserAvatar } from "@/components/common/user-avatar"
import { toast } from "sonner"
import { normalizeAISidecar, estimateTokens, plainTextOf } from "@/lib/ai-sidecar"
import { ReasoningPanel, ToolTimeline, UsageChip } from "@/components/ai/ai-sidecar-blocks"
import { normalizeExecutionTarget, TARGET_LABELS, TARGET_STYLES } from "@/lib/execution-target"

// Token-estimate + HTML-stripping helpers live in `@/lib/ai-sidecar` so the
// fallback heuristic (~4 chars / token) and the HTML strip behave identically
// on every AI surface. The contract (Frontend Consumption Rule §2) requires
// the heuristic to be used ONLY when `metadata.ai_sidecar.usage` is absent.

export default function DMConversationPage() {
  return <DMConversationContent />
}

function DMConversationContent() {
  const params = useParams()
  const router = useRouter()
  const dmId = params.id as string

  const { messages, fetchDMMessages, sendDMMessage, streamingDMMessages } = useMessageStore()
  const { conversations, fetchConversations } = useDMStore()
  const { users, currentUser } = useUserStore()
  const scrollRef = useRef<HTMLDivElement>(null)

  const conversation = conversations.find(c => c.id === dmId)
  const otherUserId = conversation?.userIds?.find((id: string) => id !== currentUser?.id)
  const otherUser = otherUserId ? users.find(u => u.id === otherUserId) : null
  const isAI = otherUser ? isAIUserLike(otherUser) : false

  const dmMessages = messages.filter(m => m.dmId === dmId)
  const streamEntries = Object.entries(streamingDMMessages).filter(([, v]) => v.dmId === dmId)
  const isStreaming = streamEntries.length > 0

  // Session-level token meter. The Unified AI Side-Channel Contract
  // (Frontend Consumption Rule §2) requires us to PREFER
  // `metadata.ai_sidecar.usage` over heuristics when the backend has
  // supplied real usage telemetry. We sum the authoritative payload for
  // every AI bubble that has one, and fall back to the ~4 chars/token
  // estimate for bubbles that don't — plus stream chunks (always estimated
  // because `usage` arrives only at completion).
  const sessionTokens = useMemo(() => {
    let input = 0, output = 0
    for (const m of dmMessages) {
      const text = plainTextOf(m.content || "")
      const sidecar = normalizeAISidecar((m as any).metadata)
      if (m.senderId === currentUser?.id) {
        // User-authored — always an estimate (we don't track upstream
        // request usage on user bubbles).
        input += estimateTokens(text)
      } else if (sidecar?.usage?.total_tokens || sidecar?.usage?.output_tokens) {
        // Authoritative path: prefer real LLM telemetry.
        output += sidecar.usage.output_tokens ?? sidecar.usage.total_tokens ?? 0
        input += sidecar.usage.input_tokens ?? 0
      } else {
        output += estimateTokens(text)
      }
    }
    for (const [, s] of streamEntries) output += estimateTokens(s.text)
    return { input, output, total: input + output }
  }, [dmMessages, streamEntries, currentUser])

  useEffect(() => {
    if (!conversations.length) fetchConversations()
  }, [conversations.length, fetchConversations])

  useEffect(() => {
    if (dmId) fetchDMMessages(dmId)
  }, [dmId, fetchDMMessages])

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight
    }
  }, [dmMessages.length, streamEntries.length])

  const copyMessage = async (id: string, html: string) => {
    try {
      await navigator.clipboard.writeText(plainTextOf(html))
      toast.success("Copied")
    } catch {
      toast.error("Copy failed")
    }
  }

  return (
    <div className="flex flex-col h-full bg-white dark:bg-[#1a1d21]">
      {/* Header */}
      <header className="h-14 px-4 flex items-center gap-3 border-b shrink-0 bg-white dark:bg-[#1a1d21]">
        <Button
          variant="ghost"
          size="icon"
          className="h-8 w-8 shrink-0 md:hidden"
          onClick={() => router.push('/workspace/dms')}
          title="Back"
        >
          <ArrowLeft className="w-4 h-4" />
        </Button>
        {otherUser ? (
          <>
            <div className={cn(
              "w-9 h-9 rounded-full flex items-center justify-center shrink-0",
              isAI ? "bg-violet-100 dark:bg-violet-900/20" : "bg-muted"
            )}>
              {isAI ? (
                <Sparkles className="w-4 h-4 text-violet-600" />
              ) : (
                <Avatar className="w-9 h-9">
                  <AvatarImage src={(otherUser as any).avatar} />
                  <AvatarFallback className="text-xs font-bold">{otherUser.name.substring(0, 2).toUpperCase()}</AvatarFallback>
                </Avatar>
              )}
            </div>
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-1.5">
                <span className="font-bold text-sm truncate">{otherUser.name}</span>
                {isAI && (
                  <Badge className="text-[8px] font-black bg-violet-500/10 text-violet-600 border-none h-4 px-1">AI</Badge>
                )}
              </div>
              <p className="text-[10px] text-muted-foreground leading-none mt-0.5">
                {isAI
                  ? (isStreaming ? "Thinking…" : "AI Assistant · Always available")
                  : (otherUser as any).title || "Direct message"}
              </p>
            </div>
            {/* Session-level token meter for AI conversations (request #2). */}
            {isAI && sessionTokens.total > 0 && (
              <div className="hidden sm:flex items-center gap-1.5 text-[10px] font-bold uppercase tracking-widest text-muted-foreground shrink-0">
                <Hash className="w-3 h-3" />
                <span title="Estimated tokens this session (in/out)">
                  {sessionTokens.input.toLocaleString()} in · {sessionTokens.output.toLocaleString()} out
                </span>
              </div>
            )}
          </>
        ) : (
          <span className="font-bold text-sm text-muted-foreground">Loading…</span>
        )}
      </header>

      {/* Messages */}
      <div ref={scrollRef} className="flex-1 overflow-y-auto px-4 py-4 space-y-3">
        {isAI && dmMessages.length === 0 && !isStreaming && (
          <AIWelcome name={otherUser?.name ?? "AI Assistant"} />
        )}

        {dmMessages.map((msg, idx) => {
          const sender = users.find(u => u.id === msg.senderId)
          const senderIsAI = sender ? isAIUserLike(sender) : false
          const isOwn = msg.senderId === currentUser?.id
          const text = plainTextOf(msg.content || "")
          const renderedContent = senderIsAI ? renderAssistantMessageHtml(msg.content || "") : (msg.content || "")

          // Unified AI Side-Channel Contract: every AI surface funnels its
          // metadata through `normalizeAISidecar`, which accepts the
          // canonical `metadata.ai_sidecar` shape AND the legacy flat
          // fields (`metadata.reasoning`/`tool_calls`/`usage`) so mixed-
          // version backends keep rendering correctly during rollout.
          const sidecar = senderIsAI ? normalizeAISidecar((msg as any).metadata) : null
          const isLastAI = senderIsAI && idx === dmMessages.length - 1

          return (
            <div key={msg.id} className={cn("flex gap-2.5 items-end group", isOwn && "flex-row-reverse")}>
              {senderIsAI ? (
                <div className="w-7 h-7 rounded-full bg-violet-100 dark:bg-violet-900/20 flex items-center justify-center shrink-0">
                  <Sparkles className="w-3.5 h-3.5 text-violet-600" />
                </div>
              ) : (
                <UserAvatar
                  name={sender?.name ?? "?"}
                  src={(sender as any)?.avatar}
                  className="w-7 h-7 shrink-0"
                />
              )}
              <div className={cn("max-w-[78%] space-y-1", isOwn && "items-end flex flex-col")}>
                {!isOwn && (
                  <span className="text-[10px] font-semibold text-muted-foreground px-1">
                    {sender?.name ?? "Unknown"}
                  </span>
                )}

                {/* AI side-channel: reasoning + tool steps render ABOVE the
                    answer bubble so the user sees the path the AI took.
                    Both blocks read directly from `normalizeAISidecar`'s
                    canonical output — no per-surface special cases. */}
                {senderIsAI && sidecar?.reasoning && (
                  <ReasoningPanel reasoning={sidecar.reasoning} />
                )}
                {senderIsAI && sidecar?.tool_calls && sidecar.tool_calls.length > 0 && (
                  <ToolTimeline toolCalls={sidecar.tool_calls} />
                )}

                <div className={cn(
                  "px-3.5 py-2 rounded-2xl text-sm leading-relaxed",
                  isOwn
                    ? "bg-violet-600 text-white rounded-br-sm"
                    : "bg-muted dark:bg-muted/60 text-foreground rounded-bl-sm",
                )}>
                  <div
                    className="break-words [&_a]:underline [&_blockquote]:border-l-2 [&_blockquote]:border-current/20 [&_blockquote]:my-2 [&_blockquote]:pl-3 [&_blockquote]:italic [&_code]:rounded [&_code]:bg-black/10 [&_code]:px-1 [&_code]:py-0.5 [&_h1]:mb-2 [&_h1]:text-base [&_h1]:font-black [&_h2]:mb-2 [&_h2]:text-sm [&_h2]:font-black [&_h3]:mb-2 [&_h3]:text-sm [&_h3]:font-bold [&_li]:my-1 [&_ol]:my-2 [&_ol]:list-decimal [&_ol]:pl-5 [&_p]:my-2 [&_p:first-child]:mt-0 [&_p:last-child]:mb-0 [&_pre]:my-2 [&_pre]:overflow-x-auto [&_pre]:rounded-lg [&_pre]:bg-black/10 [&_pre]:px-3 [&_pre]:py-2 [&_pre]:font-mono [&_pre]:text-xs [&_strong]:font-black [&_ul]:my-2 [&_ul]:list-disc [&_ul]:pl-5"
                    dangerouslySetInnerHTML={{ __html: renderedContent }}
                  />
                </div>

                {/* Phase 70B: light execution-target chip — shown only when
                    the AI message carries a normalized execution target via
                    `ai_sidecar.analysis.default_execution_target`. No action
                    is triggered from DM surface per Codex spec. */}
                {(() => {
                  if (!senderIsAI || !sidecar?.analysis) return null
                  const et = normalizeExecutionTarget(sidecar.analysis?.default_execution_target)
                  if (!et) return null
                  const style = TARGET_STYLES[et.type]
                  return (
                    <div className="px-1">
                      <span className={cn(
                        "inline-flex items-center gap-1 text-[9px] font-black uppercase tracking-widest px-1.5 py-0.5 rounded ring-1",
                        style.bg, style.text, style.ring,
                      )}>
                        <Zap className="w-2.5 h-2.5" />
                        {TARGET_LABELS[et.type]}
                      </span>
                    </div>
                  )
                })()}

                {/* Token + action footer (request #2). Hidden on user bubbles
                    in the user's preferred dense layout — only AI bubbles get
                    the ChatGPT-style chips so the cost feedback is helpful
                    without being noisy on either side. */}
                <div className={cn(
                  "flex items-center gap-2 text-[9px] text-muted-foreground/80 px-1",
                  isOwn ? "justify-end" : "justify-start",
                )}>
                  <span>
                    {msg.createdAt ? formatDistanceToNow(new Date(msg.createdAt), { addSuffix: true }) : ""}
                  </span>
                  {senderIsAI && (
                    <>
                      <span aria-hidden>·</span>
                      {/* Authoritative usage chip when the backend supplied
                          `metadata.ai_sidecar.usage`; otherwise fall back to
                          the 4-char heuristic per Frontend Consumption Rule §2. */}
                      {sidecar?.usage
                        ? <UsageChip usage={sidecar.usage} />
                        : (
                          <span
                            className="inline-flex items-center gap-0.5"
                            title="Client-side estimate (no usage payload yet)"
                          >
                            <Hash className="w-2.5 h-2.5" />
                            {estimateTokens(text).toLocaleString()} tok
                          </span>
                        )}
                    </>
                  )}
                  <button
                    type="button"
                    onClick={() => copyMessage(msg.id, msg.content || "")}
                    className="opacity-0 group-hover:opacity-100 transition-opacity inline-flex items-center gap-0.5 hover:text-foreground"
                    title="Copy"
                  >
                    <CopyIcon className="w-2.5 h-2.5" />
                  </button>
                  {isLastAI && (
                    <button
                      type="button"
                      onClick={() => {
                        // Find the most recent user message and re-send it.
                        const lastUser = [...dmMessages].reverse().find(m => m.senderId === currentUser?.id)
                        if (currentUser && dmId && lastUser) {
                          sendDMMessage(dmId, lastUser.content || "", currentUser.id)
                        }
                      }}
                      className="opacity-0 group-hover:opacity-100 transition-opacity inline-flex items-center gap-0.5 hover:text-foreground"
                      title="Regenerate"
                    >
                      <RotateCcw className="w-2.5 h-2.5" />
                    </button>
                  )}
                </div>
              </div>
            </div>
          )
        })}

        {/* Streaming AI response — live chunks before the final message arrives.
            We render the partial text inside a "Thinking…" reasoning block
            (collapsible) PLUS the live cursor in the regular bubble, so the
            user sees both the chain-of-thought feel and the eventual answer. */}
        {streamEntries.map(([tempId, s]) => (
          <div key={tempId} className="flex gap-2.5 items-end">
            <div className="w-7 h-7 rounded-full bg-violet-100 dark:bg-violet-900/20 flex items-center justify-center shrink-0 animate-pulse">
              <Sparkles className="w-3.5 h-3.5 text-violet-600" />
            </div>
            <div className="max-w-[78%] space-y-1">
              <span className="text-[10px] font-semibold text-muted-foreground px-1">
                {otherUser?.name ?? "AI Assistant"}
              </span>
              <ReasoningPanel
                reasoning={{ summary: s.text || "Working through your request…", segments: [] }}
                live
              />
              <div className="bg-muted dark:bg-muted/60 px-3.5 py-2 rounded-2xl rounded-bl-sm text-sm leading-relaxed text-foreground">
                <div
                  className="break-words [&_a]:underline [&_blockquote]:border-l-2 [&_blockquote]:border-current/20 [&_blockquote]:my-2 [&_blockquote]:pl-3 [&_blockquote]:italic [&_code]:rounded [&_code]:bg-black/10 [&_code]:px-1 [&_code]:py-0.5 [&_h1]:mb-2 [&_h1]:text-base [&_h1]:font-black [&_h2]:mb-2 [&_h2]:text-sm [&_h2]:font-black [&_h3]:mb-2 [&_h3]:text-sm [&_h3]:font-bold [&_li]:my-1 [&_ol]:my-2 [&_ol]:list-decimal [&_ol]:pl-5 [&_p]:my-2 [&_p:first-child]:mt-0 [&_p:last-child]:mb-0 [&_pre]:my-2 [&_pre]:overflow-x-auto [&_pre]:rounded-lg [&_pre]:bg-black/10 [&_pre]:px-3 [&_pre]:py-2 [&_pre]:font-mono [&_pre]:text-xs [&_strong]:font-black [&_ul]:my-2 [&_ul]:list-disc [&_ul]:pl-5"
                  dangerouslySetInnerHTML={{ __html: `${renderAssistantMessageHtml(s.text || "")}<span class="inline-block w-0.5 h-4 bg-violet-500 ml-0.5 animate-[blink_0.9s_step-end_infinite] align-middle"></span>` }}
                />
              </div>
              <div className="flex items-center gap-2 text-[9px] text-muted-foreground/80 px-1">
                <span className="inline-flex items-center gap-0.5">
                  <Hash className="w-2.5 h-2.5" />
                  {estimateTokens(s.text).toLocaleString()} tok · streaming
                </span>
              </div>
            </div>
          </div>
        ))}

        {/* Typing indicator (shows "AI Assistant is typing..." from backend typing.updated events) */}
        <TypingIndicator scope={`dm:${dmId}`} />
      </div>

      {/* Composer */}
      <div className="border-t bg-white dark:bg-[#1a1d21] shrink-0">
        <MessageComposer
          placeholder={isAI ? `Message ${otherUser?.name ?? "AI Assistant"}…` : "Message…"}
          scope={`dm:${dmId}`}
          onSend={(content) => {
            if (currentUser && dmId) {
              sendDMMessage(dmId, content, currentUser.id)
            }
          }}
        />
      </div>
    </div>
  )
}

function renderAssistantMessageHtml(content: string): string {
  const normalized = content.replace(/\r\n?/g, "\n").trim()
  if (!normalized) return ""

  let html = ""
  let paragraphBuffer: string[] = []
  let quoteBuffer: string[] = []
  let listType: "ul" | "ol" | null = null
  let listItems: string[] = []
  let inCodeBlock = false
  let codeLanguage = ""
  let codeLines: string[] = []

  const flushParagraph = () => {
    if (paragraphBuffer.length === 0) return
    html += `<p>${paragraphBuffer.map(formatInlineMarkdown).join("<br/>")}</p>`
    paragraphBuffer = []
  }

  const flushQuote = () => {
    if (quoteBuffer.length === 0) return
    html += `<blockquote><p>${quoteBuffer.map(formatInlineMarkdown).join("<br/>")}</p></blockquote>`
    quoteBuffer = []
  }

  const flushList = () => {
    if (!listType || listItems.length === 0) return
    html += `<${listType}>${listItems.map(item => `<li>${formatInlineMarkdown(item)}</li>`).join("")}</${listType}>`
    listType = null
    listItems = []
  }

  const flushCodeBlock = () => {
    html += `<pre><code${codeLanguage ? ` class="language-${escapeAttribute(codeLanguage)}"` : ""}>${escapeHtml(codeLines.join("\n"))}</code></pre>`
    codeLanguage = ""
    codeLines = []
  }

  for (const line of normalized.split("\n")) {
    const codeFence = line.match(/^```([\w-]+)?\s*$/)
    if (codeFence) {
      flushParagraph()
      flushQuote()
      flushList()
      if (inCodeBlock) {
        flushCodeBlock()
      } else {
        codeLanguage = codeFence[1] ?? ""
      }
      inCodeBlock = !inCodeBlock
      continue
    }

    if (inCodeBlock) {
      codeLines.push(line)
      continue
    }

    if (!line.trim()) {
      flushParagraph()
      flushQuote()
      flushList()
      continue
    }

    const heading = line.match(/^(#{1,6})\s+(.*)$/)
    if (heading) {
      flushParagraph()
      flushQuote()
      flushList()
      const level = Math.min(heading[1].length, 6)
      html += `<h${level}>${formatInlineMarkdown(heading[2])}</h${level}>`
      continue
    }

    if (/^(?:-{3,}|\*{3,}|_{3,})$/.test(line.trim())) {
      flushParagraph()
      flushQuote()
      flushList()
      html += "<hr/>"
      continue
    }

    const quote = line.match(/^>\s?(.*)$/)
    if (quote) {
      flushParagraph()
      flushList()
      quoteBuffer.push(quote[1])
      continue
    }

    flushQuote()

    const ordered = line.match(/^\d+\.\s+(.*)$/)
    if (ordered) {
      flushParagraph()
      if (listType !== "ol") {
        flushList()
        listType = "ol"
      }
      listItems.push(ordered[1])
      continue
    }

    const unordered = line.match(/^[-*+]\s+(.*)$/)
    if (unordered) {
      flushParagraph()
      if (listType !== "ul") {
        flushList()
        listType = "ul"
      }
      listItems.push(unordered[1])
      continue
    }

    flushList()
    paragraphBuffer.push(line)
  }

  if (inCodeBlock) {
    flushCodeBlock()
  }

  flushParagraph()
  flushQuote()
  flushList()
  return html
}

function formatInlineMarkdown(content: string): string {
  const tokens: string[] = []
  const stash = (value: string) => `@@MD_TOKEN_${tokens.push(value) - 1}@@`

  let html = escapeHtml(content)
  html = html.replace(/`([^`]+)`/g, (_, code: string) => stash(`<code>${code}</code>`))
  html = html.replace(/\[([^\]]+)\]\((https?:\/\/[^\s)]+)\)/g, (_, label: string, href: string) => (
    stash(`<a href="${escapeAttribute(href)}" target="_blank" rel="noreferrer noopener">${label}</a>`)
  ))
  html = html.replace(/\*\*([^*][\s\S]*?)\*\*/g, "<strong>$1</strong>")
  html = html.replace(/__([^_][\s\S]*?)__/g, "<strong>$1</strong>")
  html = html.replace(/~~([^~][\s\S]*?)~~/g, "<del>$1</del>")
  html = html.replace(/(^|[\s(])\*([^*\n]+)\*(?=$|[\s).,!?:;])/g, "$1<em>$2</em>")
  html = html.replace(/(^|[\s(])_([^_\n]+)_(?=$|[\s).,!?:;])/g, "$1<em>$2</em>")
  return html.replace(/@@MD_TOKEN_(\d+)@@/g, (_, index: string) => tokens[Number(index)] ?? "")
}

function escapeHtml(value: string): string {
  return value
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
}

function escapeAttribute(value: string): string {
  return value
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#39;")
}

// ── AI welcome state ────────────────────────────────────────────────────────
//
// Reasoning / tool / usage blocks now live in `@/components/ai/ai-sidecar-blocks`
// and are produced by `normalizeAISidecar()` from `@/lib/ai-sidecar`. See the
// Unified AI Side-Channel Contract spec for why this had to become cross-surface.

function AIWelcome({ name }: { name: string }) {
  return (
    <div className="flex flex-col items-center gap-4 py-16 text-center">
      <div className="w-16 h-16 rounded-2xl bg-violet-100 dark:bg-violet-900/20 flex items-center justify-center">
        <Sparkles className="w-8 h-8 text-violet-600" />
      </div>
      <div className="space-y-1">
        <h3 className="font-black text-base tracking-tight">Chat with {name}</h3>
        <p className="text-sm text-muted-foreground max-w-xs leading-relaxed">
          Ask anything — the AI thinks through your question and responds in real time, showing its reasoning, the tools it ran, and how many tokens it used.
        </p>
      </div>
    </div>
  )
}
