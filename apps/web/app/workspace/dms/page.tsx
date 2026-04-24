"use client"

// ── /workspace/dms (empty-state) ─────────────────────────────────────────────
//
// Right-pane empty state for the new two-pane DMs surface (request #1).
// The conversation list is rendered by the shared `dms/layout.tsx`; this
// page only renders when no `[id]` route segment is active. Picking a
// conversation from the left rail navigates to `/workspace/dms/:dmId`
// which Next.js mounts inside the same layout.

import { MessageSquare, Sparkles } from "lucide-react"

export default function DMsEmptyStatePage() {
  return (
    <div className="flex-1 flex items-center justify-center p-12 bg-gradient-to-br from-violet-50/40 via-white to-white dark:from-violet-950/10 dark:via-[#1a1d21] dark:to-[#1a1d21]">
      <div className="max-w-md text-center space-y-5">
        <div className="mx-auto w-16 h-16 rounded-2xl bg-violet-100 dark:bg-violet-900/20 flex items-center justify-center">
          <MessageSquare className="w-8 h-8 text-violet-600" />
        </div>
        <div className="space-y-1.5">
          <h2 className="text-xl font-black tracking-tight">Pick a conversation</h2>
          <p className="text-sm text-muted-foreground leading-relaxed">
            Choose someone from the left to start chatting, or pick the
            <span className="inline-flex items-center gap-1 mx-1 px-1.5 py-0.5 rounded bg-violet-500/10 text-violet-600 text-[10px] font-black uppercase tracking-wider">
              <Sparkles className="w-3 h-3" /> AI Assistant
            </span>
            to think through a problem.
          </p>
        </div>
      </div>
    </div>
  )
}
