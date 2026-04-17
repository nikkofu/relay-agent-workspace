"use client"

import { Search } from "lucide-react"
import { Input } from "@/components/ui/input"
import { ScrollArea } from "@/components/ui/scroll-area"
import { UserAvatar } from "@/components/common/user-avatar"
import { useUserStore } from "@/stores/user-store"

export default function DMsPage() {
  const { users } = useUserStore()
  
  return (
    <div className="flex-1 flex flex-col h-full bg-white dark:bg-[#1a1d21]">
      <header className="h-14 px-4 flex items-center justify-between border-b shrink-0 bg-white dark:bg-[#1a1d21] z-10">
        <h2 className="font-bold text-lg text-foreground">Direct Messages</h2>
      </header>
      
      <div className="p-4 border-b">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <Input placeholder="Search for people" className="pl-9 bg-muted/50 border-none focus-visible:ring-1 text-foreground" />
        </div>
      </div>

      <ScrollArea className="flex-1">
        <div className="p-4 flex flex-col gap-1">
          <h3 className="text-xs font-bold text-muted-foreground uppercase px-2 mb-2">Recent conversations</h3>
          {users.map(user => (
            <div key={user.id} className="flex items-center gap-3 p-2 hover:bg-muted/50 rounded-lg cursor-pointer group transition-colors">
              <UserAvatar src={user.avatar} name={user.name} status={user.status} className="h-10 w-10" />
              <div className="flex-1 min-w-0">
                <div className="flex items-center justify-between">
                  <span className="text-sm font-bold truncate text-foreground">{user.name}</span>
                  <span className="text-[10px] text-muted-foreground">Just now</span>
                </div>
                <p className="text-xs text-muted-foreground truncate leading-tight">
                  {user.statusText || "Online"}
                </p>
              </div>
            </div>
          ))}
        </div>
      </ScrollArea>
    </div>
  )
}
