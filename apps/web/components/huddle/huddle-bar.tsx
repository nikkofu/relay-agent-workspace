"use client"

import { Headphones, MicOff, MonitorUp, Sparkles, X, Mic } from "lucide-react"
import { Button } from "@/components/ui/button"
import { useState } from "react"
import { cn } from "@/lib/utils"

export function HuddleBar() {
  const [isActive, setIsActive] = useState(false)
  const [isMuted, setIsMuted] = useState(true)

  if (!isActive) {
    return (
      <div className="h-14 border-t border-white/5 px-4 flex items-center justify-between shrink-0 bg-[#350d36] dark:bg-[#1a1d21]/50 group transition-colors">
        <div 
          className="flex items-center gap-2 cursor-pointer group/item"
          onClick={() => setIsActive(true)}
        >
          <div className="w-8 h-8 rounded-full bg-white/10 flex items-center justify-center group-hover/item:bg-white/20 transition-colors shadow-inner">
            <Headphones className="w-4 h-4 text-white" />
          </div>
          <div className="flex flex-col">
            <span className="text-[13px] font-bold text-white leading-tight">Huddle</span>
            <span className="text-[10px] text-white/50 flex items-center gap-1">
              Start a huddle
            </span>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="h-auto border-t border-white/5 px-3 py-3 flex flex-col gap-3 shrink-0 bg-[#007a5a] text-white animate-in slide-in-from-bottom-2 duration-300">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <div className="relative">
            <div className="w-8 h-8 rounded-full bg-white/20 flex items-center justify-center">
              <Headphones className="w-4 h-4 text-white" />
            </div>
            <div className="absolute -bottom-0.5 -right-0.5 h-2.5 w-2.5 rounded-full bg-green-400 border-2 border-[#007a5a] animate-pulse" />
          </div>
          <div className="flex flex-col">
            <span className="text-[12px] font-bold leading-tight">General Huddle</span>
            <span className="text-[10px] text-white/70 flex items-center gap-1">
              <Sparkles className="w-2.5 h-2.5 text-yellow-300" /> AI notes active
            </span>
          </div>
        </div>
        <Button 
          variant="ghost" 
          size="icon" 
          className="h-6 w-6 text-white hover:bg-white/10"
          onClick={() => setIsActive(false)}
        >
          <X className="w-3.5 h-3.5" />
        </Button>
      </div>

      <div className="flex items-center justify-center gap-2 bg-black/10 rounded-lg p-1">
        <Button 
          variant="ghost" 
          size="icon" 
          className={cn(
            "h-8 w-8 rounded-md transition-colors",
            isMuted ? "text-white/50 hover:bg-white/10" : "bg-white text-[#007a5a] hover:bg-white/90"
          )}
          onClick={() => setIsMuted(!isMuted)}
        >
          {isMuted ? <MicOff className="w-4 h-4" /> : <Mic className="w-4 h-4" />}
        </Button>
        <Button variant="ghost" size="icon" className="h-8 w-8 text-white hover:bg-white/10 rounded-md">
          <MonitorUp className="w-4 h-4" />
        </Button>
        <div className="w-[1px] h-4 bg-white/10 mx-0.5" />
        <Button variant="ghost" size="icon" className="h-8 w-8 text-white hover:bg-white/10 rounded-md">
          <Sparkles className="w-4 h-4" />
        </Button>
      </div>
    </div>
  )
}
