"use client"

import { useUIStore } from "@/stores/ui-store"
import { DockedChatWindow } from "./docked-chat-window"

export function DockedChatContainer() {
  const { dockedChats } = useUIStore()

  if (dockedChats.length === 0) return null

  return (
    <div className="fixed bottom-0 right-0 z-[100] flex flex-row-reverse gap-4 px-4 pointer-events-none">
      {dockedChats.map((userId, index) => (
        <div key={userId} className="pointer-events-auto">
          <DockedChatWindow userId={userId} index={index} />
        </div>
      ))}
    </div>
  )
}
