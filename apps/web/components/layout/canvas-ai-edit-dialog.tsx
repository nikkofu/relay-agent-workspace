"use client"

// ── Canvas AI edit dialog ────────────────────────────────────────────────────
//
// Invoked from the TipTap toolbar's "AI Edit" button. Offers preset intents
// (Expand / Shorten / Rephrase / Fix grammar / Tone …) plus a free-form
// instruction field. Streams the result from POST /api/v1/ai/execute using
// the artifact's channel scope and shows a live preview. The user can accept
// (replaces the target range via onApply) or discard.
//
// Output is plain text from the LLM; we convert newline-separated blocks into
// <p> tags so TipTap renders paragraph breaks correctly when inserted.

import { useState, useMemo } from "react"
import { Wand2, Loader2, Sparkles, Check, X, RefreshCw } from "lucide-react"
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Label } from "@/components/ui/label"
import { API_BASE_URL } from "@/lib/constants"
import { cn } from "@/lib/utils"
import { toast } from "sonner"

interface CanvasAIEditDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  channelId: string
  targetText: string
  onApply: (html: string) => void
}

type PresetKey =
  | "expand"
  | "shorten"
  | "rephrase"
  | "fix_grammar"
  | "tone_formal"
  | "tone_casual"
  | "translate_en"
  | "translate_zh"
  | "custom"

interface Preset {
  key: PresetKey
  label: string
  instruction: string
}

const PRESETS: Preset[] = [
  { key: "expand",       label: "Expand",          instruction: "Expand the text below with additional detail, examples, and context, while preserving the original meaning." },
  { key: "shorten",      label: "Shorten",         instruction: "Shorten the text below to its essential points while preserving meaning and tone." },
  { key: "rephrase",     label: "Rephrase",        instruction: "Rephrase the text below with clearer wording while preserving meaning." },
  { key: "fix_grammar",  label: "Fix grammar",     instruction: "Correct spelling, grammar, and punctuation in the text below without changing its meaning." },
  { key: "tone_formal",  label: "Formal tone",     instruction: "Rewrite the text below in a more professional, formal tone." },
  { key: "tone_casual",  label: "Casual tone",     instruction: "Rewrite the text below in a more relaxed, conversational tone." },
  { key: "translate_en", label: "Translate → EN",  instruction: "Translate the text below into natural, fluent English." },
  { key: "translate_zh", label: "翻译 → 中文",      instruction: "将下面的文字翻译成自然流畅的中文。" },
]

export function CanvasAIEditDialog({
  open, onOpenChange, channelId, targetText, onApply,
}: CanvasAIEditDialogProps) {
  const [selectedPreset, setSelectedPreset] = useState<PresetKey>("expand")
  const [customInstruction, setCustomInstruction] = useState("")
  const [result, setResult] = useState("")
  const [isStreaming, setIsStreaming] = useState(false)
  const [abortCtrl, setAbortCtrl] = useState<AbortController | null>(null)

  const effectiveInstruction = useMemo(() => {
    if (selectedPreset === "custom") return customInstruction.trim()
    const preset = PRESETS.find(p => p.key === selectedPreset)
    return preset?.instruction ?? ""
  }, [selectedPreset, customInstruction])

  const targetPreview = targetText.length > 280 ? targetText.slice(0, 280) + "…" : targetText

  const runAI = async () => {
    if (!effectiveInstruction) {
      toast.error("Pick a preset or enter an instruction")
      return
    }
    if (!channelId) {
      toast.error("Missing channel context for AI request")
      return
    }

    // Cancel any in-flight request first.
    abortCtrl?.abort()
    const ctrl = new AbortController()
    setAbortCtrl(ctrl)
    setResult("")
    setIsStreaming(true)

    const prompt =
      `${effectiveInstruction}\n\n---\n\n${targetText}\n\n---\n\n` +
      `Return ONLY the transformed text. Do not include explanations, preambles, or markdown fencing.`

    try {
      const res = await fetch(`${API_BASE_URL}/ai/execute`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ prompt, channel_id: channelId }),
        signal: ctrl.signal,
      })
      if (!res.ok || !res.body) {
        throw new Error(`AI request failed (${res.status})`)
      }
      const reader = res.body.getReader()
      const decoder = new TextDecoder()
      let buffer = ""
      let acc = ""

      while (true) {
        const { value, done } = await reader.read()
        if (done) break
        buffer += decoder.decode(value, { stream: true })
        // Server-Sent Events frames are separated by a blank line.
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
            // /ai/execute emits event:"chunk" data:{text:"..."} for content tokens
            // and event:"reasoning" for reasoning — we skip those. Also tolerate
            // alternate field names used by other endpoints.
            if (evt === "chunk" || evt === "message" || evt === "token") {
              const chunk =
                (typeof parsed.text === "string" && parsed.text) ||
                (typeof parsed.content === "string" && parsed.content) ||
                (typeof parsed.delta === "string" && parsed.delta) ||
                ""
              if (chunk) {
                acc += chunk
                setResult(acc)
              }
            } else if (evt === "error") {
              throw new Error(parsed.message || "AI streaming error")
            }
            // "start", "reasoning", "done", "conversation" are intentionally ignored.
          } catch {
            // Non-JSON data — treat as raw chunk.
            acc += data
            setResult(acc)
          }
        }
      }
    } catch (err: unknown) {
      const isAbort =
        err instanceof DOMException && err.name === "AbortError"
      if (!isAbort) {
        console.error("Canvas AI edit failed:", err)
        toast.error(err instanceof Error ? err.message : "AI request failed")
      }
    } finally {
      setIsStreaming(false)
      setAbortCtrl(null)
    }
  }

  const stop = () => {
    abortCtrl?.abort()
  }

  const accept = () => {
    if (!result.trim()) return
    // Normalize into paragraph-wrapped HTML so TipTap renders line breaks.
    const html = result
      .split(/\n{2,}/)
      .map(block => `<p>${escapeHtml(block).replace(/\n/g, "<br/>")}</p>`)
      .join("")
    onApply(html)
    reset()
    onOpenChange(false)
  }

  const reset = () => {
    setResult("")
    setSelectedPreset("expand")
    setCustomInstruction("")
    abortCtrl?.abort()
    setAbortCtrl(null)
    setIsStreaming(false)
  }

  return (
    <Dialog
      open={open}
      onOpenChange={next => {
        if (!next) reset()
        onOpenChange(next)
      }}
    >
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2 text-sm font-black uppercase tracking-tight">
            <Wand2 className="w-4 h-4 text-purple-600" />
            AI Edit
          </DialogTitle>
          <DialogDescription className="text-xs">
            Pick a preset or type a custom instruction. The result previews below — accept to replace the selected range in the canvas.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-3 py-1">
          {/* Target preview */}
          <div className="rounded-md border border-purple-400/30 bg-purple-500/5 px-3 py-2">
            <p className="text-[9px] font-black uppercase tracking-widest text-purple-700 dark:text-purple-300 mb-1">
              Target text ({targetText.length.toLocaleString()} chars)
            </p>
            <p className="text-[11px] text-foreground/80 italic leading-snug whitespace-pre-wrap line-clamp-4">
              {targetPreview || "(empty — AI will draft something from scratch)"}
            </p>
          </div>

          {/* Preset chips */}
          <div>
            <Label className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">
              Action
            </Label>
            <div className="flex flex-wrap gap-1.5 mt-1.5">
              {PRESETS.map(p => (
                <button
                  key={p.key}
                  type="button"
                  onClick={() => setSelectedPreset(p.key)}
                  className={cn(
                    "text-[11px] font-bold px-2.5 py-1 rounded-full border transition-colors",
                    selectedPreset === p.key
                      ? "bg-purple-600 border-purple-600 text-white"
                      : "bg-transparent border-border text-foreground/70 hover:bg-muted",
                  )}
                >
                  {p.label}
                </button>
              ))}
              <button
                type="button"
                onClick={() => setSelectedPreset("custom")}
                className={cn(
                  "text-[11px] font-bold px-2.5 py-1 rounded-full border transition-colors",
                  selectedPreset === "custom"
                    ? "bg-purple-600 border-purple-600 text-white"
                    : "bg-transparent border-border text-foreground/70 hover:bg-muted",
                )}
              >
                Custom
              </button>
            </div>
          </div>

          {/* Custom instruction */}
          {selectedPreset === "custom" && (
            <div>
              <Label className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">
                Instruction
              </Label>
              <textarea
                value={customInstruction}
                onChange={e => setCustomInstruction(e.target.value)}
                rows={2}
                placeholder="e.g. Turn this into bullet points with a clear action per line."
                className="w-full mt-1.5 p-2 text-xs rounded border bg-white dark:bg-[#19171d] focus:outline-none focus:ring-1 focus:ring-purple-500"
              />
            </div>
          )}

          {/* Result preview */}
          <div>
            <div className="flex items-center justify-between">
              <Label className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">
                Preview
              </Label>
              {isStreaming && (
                <span className="text-[9px] font-black uppercase tracking-widest text-purple-600 flex items-center gap-1">
                  <Loader2 className="w-2.5 h-2.5 animate-spin" />
                  Streaming…
                </span>
              )}
            </div>
            <div className="mt-1.5 min-h-[120px] max-h-[260px] overflow-y-auto rounded border bg-muted/10 p-3">
              {result ? (
                <p className="text-[12px] text-foreground/90 whitespace-pre-wrap leading-relaxed">
                  {result}
                </p>
              ) : (
                <p className="text-[11px] text-muted-foreground italic">
                  Run AI to see the preview here.
                </p>
              )}
            </div>
          </div>
        </div>

        <DialogFooter className="gap-2">
          <Button variant="ghost" size="sm" onClick={() => { reset(); onOpenChange(false) }}>
            Cancel
          </Button>
          {isStreaming ? (
            <Button size="sm" variant="outline" onClick={stop} className="gap-1.5">
              <X className="w-3 h-3" />
              Stop
            </Button>
          ) : (
            <>
              {result && (
                <Button
                  size="sm"
                  variant="outline"
                  onClick={runAI}
                  className="gap-1.5"
                  disabled={!effectiveInstruction}
                  title="Regenerate"
                >
                  <RefreshCw className="w-3 h-3" />
                  Retry
                </Button>
              )}
              {!result ? (
                <Button
                  size="sm"
                  onClick={runAI}
                  disabled={!effectiveInstruction}
                  className="bg-purple-600 hover:bg-purple-700 text-white gap-1.5"
                >
                  <Sparkles className="w-3 h-3" />
                  Run AI
                </Button>
              ) : (
                <Button
                  size="sm"
                  onClick={accept}
                  className="bg-emerald-600 hover:bg-emerald-700 text-white gap-1.5"
                >
                  <Check className="w-3 h-3" />
                  Accept
                </Button>
              )}
            </>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

function escapeHtml(s: string): string {
  return s
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
}
