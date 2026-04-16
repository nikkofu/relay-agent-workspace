"use client"

import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { Button } from "@/components/ui/button"
import { Smile } from "lucide-react"

interface EmojiPickerProps {
  onSelect: (emoji: string) => void
  children?: React.ReactNode
}

const COMMON_EMOJIS = ["😀", "👍", "❤️", "🚀", "🎉", "🔥", "👀", "🙌", "💯", "✨"]

export function EmojiPicker({ onSelect, children }: EmojiPickerProps) {
  return (
    <Popover>
      <PopoverTrigger asChild>
        {children || (
          <Button variant="ghost" size="icon" className="h-7 w-7 text-muted-foreground hover:bg-muted">
            <Smile className="h-4 w-4" />
          </Button>
        )}
      </PopoverTrigger>
      <PopoverContent className="w-64 p-2" align="start">
        <div className="grid grid-cols-5 gap-2">
          {COMMON_EMOJIS.map(emoji => (
            <Button
              key={emoji}
              variant="ghost"
              className="h-10 w-10 text-xl"
              onClick={() => onSelect(emoji)}
            >
              {emoji}
            </Button>
          ))}
        </div>
      </PopoverContent>
    </Popover>
  )
}
