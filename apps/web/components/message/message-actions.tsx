"use client"

import { Smile, MessageSquare, Share2, Bookmark, MoreHorizontal } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"

interface MessageActionsProps {
  onReply?: () => void
  onReact?: (emoji: string) => void
  onForward?: () => void
  onSave?: () => void
  onCopyLink?: () => void
  onMarkUnread?: () => void
  onPin?: () => void
  onDelete?: () => void
}

export function MessageActions({ 
  onReply, 
  onReact, 
  onForward, 
  onSave, 
  onCopyLink, 
  onMarkUnread, 
  onPin, 
  onDelete 
}: MessageActionsProps) {
  return (
    <div className="flex items-center bg-white dark:bg-[#1a1d21] border rounded-lg shadow-sm p-0.5">
      <Tooltip>
        <TooltipTrigger asChild>
          <Button variant="ghost" size="icon" className="h-8 w-8 hover:bg-muted" onClick={() => onReact?.("👍")}>
            <Smile className="h-4 w-4" />
          </Button>
        </TooltipTrigger>
        <TooltipContent side="top">Add reaction</TooltipContent>
      </Tooltip>

      <Tooltip>
        <TooltipTrigger asChild>
          <Button variant="ghost" size="icon" className="h-8 w-8 hover:bg-muted" onClick={onReply}>
            <MessageSquare className="h-4 w-4" />
          </Button>
        </TooltipTrigger>
        <TooltipContent side="top">Reply to thread</TooltipContent>
      </Tooltip>

      <Tooltip>
        <TooltipTrigger asChild>
          <Button variant="ghost" size="icon" className="h-8 w-8 hover:bg-muted" onClick={onForward}>
            <Share2 className="h-4 w-4" />
          </Button>
        </TooltipTrigger>
        <TooltipContent side="top">Forward message</TooltipContent>
      </Tooltip>

      <Tooltip>
        <TooltipTrigger asChild>
          <Button variant="ghost" size="icon" className="h-8 w-8 hover:bg-muted" onClick={onSave}>
            <Bookmark className="h-4 w-4" />
          </Button>
        </TooltipTrigger>
        <TooltipContent side="top">Save for later</TooltipContent>
      </Tooltip>

      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" size="icon" className="h-8 w-8 hover:bg-muted">
            <MoreHorizontal className="h-4 w-4" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" className="w-48 text-foreground">
          <DropdownMenuItem onClick={onCopyLink}>Copy link to message</DropdownMenuItem>
          <DropdownMenuItem onClick={onMarkUnread}>Mark as unread</DropdownMenuItem>
          <DropdownMenuSeparator />
          <DropdownMenuItem onClick={onPin}>Pin to channel</DropdownMenuItem>
          <DropdownMenuSeparator />
          <DropdownMenuItem className="text-destructive focus:bg-destructive/10 focus:text-destructive" onClick={onDelete}>Delete message</DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  )
}
