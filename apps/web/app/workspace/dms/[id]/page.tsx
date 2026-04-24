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

import { useEffect, useMemo, useRef, useState } from "react"
import { useParams, useRouter } from "next/navigation"
import { useMessageStore } from "@/stores/message-store"
import { useDMStore } from "@/stores/dm-store"
import { useUserStore } from "@/stores/user-store"
import { MessageComposer } from "@/components/message/message-composer"
import { TypingIndicator } from "@/components/message/typing-indicator"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import {
  ArrowLeft, Sparkles, Brain, ChevronDown, ChevronRight, Wrench, Hash,
  Check, RotateCcw, Copy as CopyIcon,
} from "lucide-react"
import { formatDistanceToNow } from "date-fns"
import { cn } from "@/lib/utils"
import { UserAvatar } from "@/components/common/user-avatar"
import { toast } from "sonner"

function isAIUserCheck(user: { id: string; name: string; email?: string }) {
  const name = user.name.toLowerCase()
  const email = (user.email ?? "").toLowerCase()
  return name.includes("assistant") || name.includes("ai") || email.startsWith("ai@") || user.id === "user-2"
}

// Quick-and-dirty token estimate. The OpenAI rule of thumb is ~4 chars
// per token for English; close enough for a footer chip until the
// backend exposes real `usage.total_tokens` on DM messages.
function estimateTokens(s: string): number {
  if (!s) return 0
  return Math.max(1, Math.round(s.length / 4))
}

// Extract any HTML markup so the token estimate doesn't include `<p>`/`<br>`.
function stripHtml(html: string): string {
  if (typeof document === "undefined") return html.replace(/<[^>]+>/g, "")
  const tmp = document.createElement("div")
  tmp.innerHTML = html
  return tmp.textContent || ""
}

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

  // Track which AI bubbles have their reasoning expanded. Defaults to the
  // currently-streaming bubble being open so the user sees the live
  // thinking unfold (matches ChatGPT's behaviour).
  const [expandedReasoning, setExpandedReasoning] = useState<Record<string, boolean>>({})

  const conversation = conversations.find(c => c.id === dmId)
  const otherUserId = conversation?.userIds?.find((id: string) => id !== currentUser?.id)
  const otherUser = otherUserId ? users.find(u => u.id === otherUserId) : null
  const isAI = otherUser ? isAIUserCheck(otherUser) : false

  const dmMessages = messages.filter(m => m.dmId === dmId)
  const streamEntries = Object.entries(streamingDMMessages).filter(([, v]) => v.dmId === dmId)
  const isStreaming = streamEntries.length > 0

  // Total tokens consumed in this conversation — a rough running cost
  // estimate so the user gets the same "300 tokens / $0.0006" footer feel
  // as the ChatGPT web UI.
  const sessionTokens = useMemo(() => {
    let input = 0, output = 0
    for (const m of dmMessages) {
      const text = stripHtml(m.content || "")
      if (m.senderId === currentUser?.id) input += estimateTokens(text)
      else output += estimateTokens(text)
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

  const toggleReasoning = (id: string) => {
    setExpandedReasoning(prev => ({ ...prev, [id]: !prev[id] }))
  }

  const copyMessage = async (id: string, html: string) => {
    try {
      await navigator.clipboard.writeText(stripHtml(html))
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
          const senderIsAI = sender ? isAIUserCheck(sender) : false
          const isOwn = msg.senderId === currentUser?.id
          const text = stripHtml(msg.content || "")
          const tokens = estimateTokens(text)

          // Pull AI side-channel data out of the message metadata when the
          // backend provides it. Falls back to undefined on legacy rows so
          // we just render the bubble like before.
          const meta = (msg as any).metadata || {}
          const reasoning: string | undefined = meta.reasoning || meta.thinking
          const toolCalls: ToolCallStep[] | undefined = meta.tool_calls || meta.tools
          const usage = meta.usage as { input_tokens?: number, output_tokens?: number, total_tokens?: number } | undefined

          const reasoningOpen = expandedReasoning[msg.id] ?? false
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

                {/* AI side-channel: reasoning + tool steps render ABOVE the answer
                    bubble so the user sees the path the AI took. */}
                {senderIsAI && reasoning && (
                  <ReasoningBlock
                    text={reasoning}
                    expanded={reasoningOpen}
                    onToggle={() => toggleReasoning(msg.id)}
                    durationMs={meta.thinking_ms}
                  />
                )}
                {senderIsAI && toolCalls && toolCalls.length > 0 && (
                  <ToolTimeline calls={toolCalls} />
                )}

                <div className={cn(
                  "px-3.5 py-2 rounded-2xl text-sm leading-relaxed",
                  isOwn
                    ? "bg-violet-600 text-white rounded-br-sm"
                    : "bg-muted dark:bg-muted/60 text-foreground rounded-bl-sm",
                )}>
                  <div dangerouslySetInnerHTML={{ __html: msg.content }} />
                </div>

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
                      <span className="inline-flex items-center gap-0.5" title={usage ? "Reported by backend" : "Client-side estimate"}>
                        <Hash className="w-2.5 h-2.5" />
                        {(usage?.total_tokens ?? usage?.output_tokens ?? tokens).toLocaleString()} tok
                      </span>
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
              <ReasoningBlock
                text={s.text || "Working through your request…"}
                expanded
                onToggle={() => { /* always-on while streaming */ }}
                live
              />
              <div className="bg-muted dark:bg-muted/60 px-3.5 py-2 rounded-2xl rounded-bl-sm text-sm leading-relaxed text-foreground">
                <span>{s.text}</span>
                <span className="inline-block w-0.5 h-4 bg-violet-500 ml-0.5 animate-[blink_0.9s_step-end_infinite] align-middle" />
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

// ── AI side-channel building blocks ─────────────────────────────────────────

interface ToolCallStep {
  name?: string
  label?: string
  status?: "pending" | "running" | "success" | "failed"
  detail?: string
  duration_ms?: number
}

function ReasoningBlock({
  text, expanded, onToggle, live, durationMs,
}: {
  text: string
  expanded: boolean
  onToggle: () => void
  live?: boolean
  durationMs?: number
}) {
  return (
    <div className={cn(
      "rounded-xl border text-[11px] leading-relaxed",
      live
        ? "border-violet-500/30 bg-violet-50/50 dark:bg-violet-950/20"
        : "border-border bg-muted/40",
    )}>
      <button
        type="button"
        onClick={onToggle}
        className="w-full flex items-center gap-1.5 px-3 py-1.5 hover:bg-muted/40 transition-colors rounded-t-xl"
      >
        {expanded ? <ChevronDown className="w-3 h-3" /> : <ChevronRight className="w-3 h-3" />}
        <Brain className={cn(
          "w-3 h-3",
          live ? "text-violet-600 animate-pulse" : "text-muted-foreground",
        )} />
        <span className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">
          {live ? "Thinking…" : "Thoughts"}
        </span>
        {durationMs !== undefined && !live && (
          <span className="text-[9px] text-muted-foreground/70 ml-auto">
            {(durationMs / 1000).toFixed(1)}s
          </span>
        )}
      </button>
      {expanded && (
        <div className="px-3 pb-2 pt-1 max-h-48 overflow-y-auto">
          <p className="text-muted-foreground whitespace-pre-wrap">{text}</p>
        </div>
      )}
    </div>
  )
}

function ToolTimeline({ calls }: { calls: ToolCallStep[] }) {
  return (
    <div className="rounded-xl border bg-muted/40 px-3 py-2 space-y-1">
      <div className="flex items-center gap-1.5">
        <Wrench className="w-3 h-3 text-muted-foreground" />
        <span className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">
          Tools used · {calls.length}
        </span>
      </div>
      <ol className="space-y-0.5">
        {calls.map((c, i) => {
          const status = c.status ?? "success"
          return (
            <li key={i} className="flex items-center gap-1.5 text-[11px]">
              <span className={cn(
                "inline-flex w-3.5 h-3.5 rounded-full items-center justify-center shrink-0",
                status === "running" && "bg-blue-500/15 text-blue-600",
                status === "success" && "bg-emerald-500/15 text-emerald-600",
                status === "failed" && "bg-rose-500/15 text-rose-600",
                status === "pending" && "bg-muted text-muted-foreground",
              )}>
                {status === "success" ? <Check className="w-2 h-2" /> : <span className="text-[9px] font-black">{i + 1}</span>}
              </span>
              <span className="font-semibold truncate">{c.label ?? c.name ?? "Tool"}</span>
              {c.detail && <span className="text-muted-foreground truncate">— {c.detail}</span>}
              {c.duration_ms !== undefined && (
                <span className="text-muted-foreground/70 ml-auto text-[9px]">
                  {c.duration_ms < 1000 ? `${c.duration_ms}ms` : `${(c.duration_ms / 1000).toFixed(1)}s`}
                </span>
              )}
            </li>
          )
        })}
      </ol>
    </div>
  )
}

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
