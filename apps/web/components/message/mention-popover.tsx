"use client"

import { Command, CommandGroup, CommandItem, CommandList } from "@/components/ui/command"
import { useUserStore } from "@/stores/user-store"
import { useDirectoryStore } from "@/stores/directory-store"
import { isAIUserLike } from "@/stores/user-store"
import { UserAvatar } from "@/components/common/user-avatar"
import { Sparkles, Shield } from "lucide-react"
import type { ComposerMentionTarget } from "@/types"

export function MentionPopover({ onSelect }: { onSelect: (target: ComposerMentionTarget) => void }) {
  const { users } = useUserStore()
  const { userGroups } = useDirectoryStore()
  
  return (
    <div className="absolute bottom-full left-0 mb-2 w-[240px] z-50 animate-in slide-in-from-bottom-2 duration-200">
      <Command className="border shadow-xl rounded-xl overflow-hidden bg-white dark:bg-[#1a1d21]">
        <CommandList className="max-h-[300px]">
          <CommandGroup heading="People">
            {users.map(user => (
              <CommandItem
                key={user.id}
                onSelect={() => onSelect({ kind: "user", user_id: user.id, name: user.name, user_type: user.userType, avatar: user.avatar })}
                className="flex items-center gap-2 p-2 cursor-pointer hover:bg-muted text-foreground"
              >
                <UserAvatar src={user.avatar} name={user.name} className="h-6 w-6" />
                <span className="text-sm font-medium">{user.name}</span>
                {isAIUserLike(user) && <Sparkles className="w-3 h-3 text-purple-500 ml-auto" />}
              </CommandItem>
            ))}
          </CommandGroup>

          {userGroups.length > 0 && (
            <CommandGroup heading="Groups">
              {userGroups.map(group => (
                <CommandItem
                  key={group.id}
                  onSelect={() => onSelect({ kind: "group", group_id: group.id, handle: group.handle, name: group.name })}
                  className="flex items-center gap-2 p-2 cursor-pointer hover:bg-muted text-foreground"
                >
                  <div className="w-6 h-6 rounded bg-purple-500/10 flex items-center justify-center text-purple-600">
                    <Shield className="w-3.5 h-3.5" />
                  </div>
                  <span className="text-sm font-bold">@{group.handle}</span>
                  <span className="text-[10px] text-muted-foreground ml-auto">{group.memberCount}</span>
                </CommandItem>
              ))}
            </CommandGroup>
          )}
        </CommandList>
      </Command>
    </div>
  )
}
