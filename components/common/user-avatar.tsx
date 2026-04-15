"use client"

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { cn } from "@/lib/utils"
import { UserStatus } from "@/types"

interface UserAvatarProps {
  src?: string
  name: string
  status?: UserStatus
  className?: string
  fallbackClassName?: string
}

export function UserAvatar({ src, name, status, className, fallbackClassName }: UserAvatarProps) {
  return (
    <div className="relative inline-block">
      <Avatar className={cn("h-9 w-9 rounded-md", className)}>
        <AvatarImage src={src} alt={name} />
        <AvatarFallback className={cn("rounded-md text-xs font-bold", fallbackClassName)}>
          {name.substring(0, 2).toUpperCase()}
        </AvatarFallback>
      </Avatar>
      {status && (
        <div 
          className={cn(
            "absolute -bottom-0.5 -right-0.5 h-3 w-3 rounded-full border-2 border-background",
            status === "online" && "bg-green-500",
            status === "away" && "bg-amber-500",
            status === "busy" && "bg-red-500",
            status === "offline" && "bg-slate-500"
          )}
        />
      )}
    </div>
  )
}
