"use client"

import { Command, CommandGroup, CommandItem, CommandList } from "@/components/ui/command"
import { useUserStore } from "@/stores/user-store"
import { UserAvatar } from "@/components/common/user-avatar"
import { Sparkles } from "lucide-react"

export function MentionPopover({ onSelect }: { onSelect: (name: string) => void }) {
  const { users } = useUserStore()
  return (
    <div className="absolute bottom-full left-0 mb-2 w-[240px] z-50 animate-in slide-in-from-bottom-2 duration-200">
      <Command className="border shadow-xl rounded-xl overflow-hidden bg-white dark:bg-[#1a1d21]">
        <CommandList className="max-h-[300px]">
          <CommandGroup heading="People">
            {users.map(user => (
              <CommandItem
                key={user.id}
                onSelect={() => onSelect(user.name)}
                className="flex items-center gap-2 p-2 cursor-pointer hover:bg-muted text-foreground"
              >                <UserAvatar src={user.avatar} name={user.name} className="h-6 w-6" />
                <span className="text-sm font-medium">{user.name}</span>
                {user.id === "user-2" && <Sparkles className="w-3 h-3 text-purple-500 ml-auto" />}
              </CommandItem>
            ))}
          </CommandGroup>
        </CommandList>
      </Command>
    </div>
  )
}
