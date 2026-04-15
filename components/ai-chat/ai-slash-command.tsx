"use client"

import { Command, CommandGroup, CommandItem, CommandList } from "@/components/ui/command"
import { Sparkles, FileText, Languages, Code, MessageSquarePlus, Zap } from "lucide-react"

export function AISlashCommand({ onSelect }: { onSelect: (cmd: string) => void }) {
  const commands = [
    { icon: Sparkles, label: "Ask AI", value: "/ai", desc: "Start a conversation" },
    { icon: FileText, label: "Summarize", value: "/summarize", desc: "Summarize current channel" },
    { icon: Languages, label: "Translate", value: "/translate", desc: "Translate last message" },
    { icon: Code, label: "Explain Code", value: "/explain", desc: "Explain code snippets" },
    { icon: MessageSquarePlus, label: "Draft Reply", value: "/draft", desc: "Help me draft a reply" },
    { icon: Zap, label: "Optimize", value: "/optimize", desc: "Rewrite my message" },
  ]

  return (
    <div className="absolute bottom-full left-0 mb-2 w-[280px] z-50 animate-in slide-in-from-bottom-2 duration-200">
      <Command className="border shadow-xl rounded-xl overflow-hidden bg-white dark:bg-[#1a1d21]">
        <CommandList className="max-h-[300px]">
          <CommandGroup heading="AI Superpowers">
            {commands.map(cmd => (
              <CommandItem 
                key={cmd.value} 
                onSelect={() => onSelect(cmd.value)} 
                className="flex items-center gap-3 p-2.5 cursor-pointer hover:bg-muted transition-colors"
              >
                <div className="w-8 h-8 rounded-lg bg-purple-100 dark:bg-purple-900/30 flex items-center justify-center shrink-0">
                  <cmd.icon className="w-4 h-4 text-purple-600 dark:text-purple-400" />
                </div>
                <div className="flex flex-col min-w-0">
                  <span className="text-sm font-bold">{cmd.label}</span>
                  <span className="text-[10px] text-muted-foreground truncate leading-tight">{cmd.desc}</span>
                </div>
                <span className="ml-auto text-[10px] font-mono opacity-40 bg-muted px-1 rounded uppercase tracking-tighter shrink-0 font-bold">
                  {cmd.value}
                </span>
              </CommandItem>
            ))}
          </CommandGroup>
        </CommandList>
      </Command>
    </div>
  )
}
