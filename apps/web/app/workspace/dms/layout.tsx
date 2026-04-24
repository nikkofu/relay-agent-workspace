"use client"

// ── DMs shared layout (WhatsApp / WeChat-style two-pane) ────────────────────
//
// Renders the conversation list on the left and the active conversation on
// the right (provided as `children` by Next.js). The left rail uses
// `useParams()` to highlight the active row, so `Suspense` is required at
// this boundary by the Next 16 cacheComponents model — that suspense
// wrapper also fixes the "Data that blocks navigation was accessed
// outside of <Suspense>" error the user hit when opening Canvas from the
// DM editor (request #3): the canvas trigger forces a layout re-render
// that re-evaluates `useParams()` from the DM tree.

import { Suspense } from "react"
import { DMConversationList } from "@/components/dm/dm-conversation-list"

export default function DMsLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex h-full w-full overflow-hidden bg-white dark:bg-[#1a1d21]">
      <Suspense
        fallback={
          <aside className="w-[320px] shrink-0 border-r bg-muted/5 animate-pulse" />
        }
      >
        <DMConversationList />
      </Suspense>
      <main className="flex-1 min-w-0 flex flex-col">
        <Suspense
          fallback={
            <div className="flex-1 flex items-center justify-center text-muted-foreground text-sm">
              Loading conversation…
            </div>
          }
        >
          {children}
        </Suspense>
      </main>
    </div>
  )
}
