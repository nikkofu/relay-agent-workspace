"use client"

import { useEffect, useState } from "react"
import { useFileStore } from "@/stores/file-store"
import { Search, Folder, FileIcon, Archive, Trash2, Download, Filter, MoreVertical, RefreshCcw } from "lucide-react"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { ScrollArea } from "@/components/ui/scroll-area"
import { 
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger 
} from "@/components/ui/dropdown-menu"
import { format } from "date-fns"
import { useChannelStore } from "@/stores/channel-store"

export default function FilesPage() {
  const { 
    files, archivedFiles, 
    fetchFiles, fetchArchivedFiles, archiveFile 
  } = useFileStore()
  const { currentChannel } = useChannelStore()
  
  const [q, setQ] = useState("")
  const [showArchived, setShowArchived] = useState(false)

  useEffect(() => {
    if (showArchived) {
      fetchArchivedFiles({ q })
    } else if (currentChannel) {
      fetchFiles(currentChannel.id)
    }
  }, [showArchived, q, currentChannel, fetchFiles, fetchArchivedFiles])

  const activeFiles = showArchived ? archivedFiles : files

  return (
    <div className="flex-1 flex flex-col bg-white dark:bg-[#1a1d21] h-full overflow-hidden">
      <header className="h-14 px-6 flex items-center border-b shrink-0 justify-between">
        <div className="flex items-center gap-2">
          <Folder className="w-5 h-5 text-blue-600" />
          <h1 className="text-lg font-black tracking-tight uppercase">Files & Assets</h1>
        </div>
        <div className="flex items-center gap-2">
          <Button 
            variant={showArchived ? "secondary" : "ghost"} 
            size="sm" 
            className="text-xs font-bold gap-2"
            onClick={() => setShowArchived(!showArchived)}
          >
            <Archive className="w-3.5 h-3.5" />
            {showArchived ? "Back to Active" : "View Archive"}
          </Button>
        </div>
      </header>

      {/* Filter Bar */}
      <div className="p-4 border-b bg-muted/10 flex gap-3">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <Input 
            placeholder={showArchived ? "Search archive..." : "Search active files..."} 
            className="pl-9 bg-white dark:bg-[#1a1d21]"
            value={q}
            onChange={(e) => setQ(e.target.value)}
          />
        </div>
        <Button variant="ghost" size="icon" onClick={() => setQ("")}>
          <Filter className="w-4 h-4" />
        </Button>
      </div>

      <ScrollArea className="flex-1">
        <div className="p-0">
          <div className="min-w-full divide-y divide-border">
            {activeFiles.map(file => (
              <div 
                key={file.id}
                className="group flex items-center justify-between p-4 hover:bg-muted/30 transition-colors"
              >
                <div className="flex items-center gap-4 min-w-0">
                  <div className="w-10 h-10 rounded-lg bg-blue-500/10 flex items-center justify-center text-blue-600 shrink-0">
                    <FileIcon className="w-5 h-5" />
                  </div>
                  <div className="min-w-0">
                    <h3 className="font-bold text-sm truncate">{file.name}</h3>
                    <p className="text-[10px] text-muted-foreground uppercase font-black tracking-tighter">
                      {file.type} • {(file.size / 1024).toFixed(1)} KB • {format(new Date(file.createdAt), 'MMM d, yyyy')}
                    </p>
                  </div>
                </div>

                <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                  <Button variant="ghost" size="icon" className="h-8 w-8" asChild>
                    <a href={file.url} target="_blank" rel="noopener noreferrer">
                      <Download className="w-4 h-4" />
                    </a>
                  </Button>
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button variant="ghost" size="icon" className="h-8 w-8">
                        <MoreVertical className="w-4 h-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      <DropdownMenuItem onClick={() => archiveFile(file.id, !showArchived)}>
                        {showArchived ? (
                          <div className="flex items-center gap-2">
                            <RefreshCcw className="w-3.5 h-3.5" />
                            <span>Restore File</span>
                          </div>
                        ) : (
                          <div className="flex items-center gap-2 text-amber-600">
                            <Archive className="w-3.5 h-3.5" />
                            <span>Archive File</span>
                          </div>
                        )}
                      </DropdownMenuItem>
                      <DropdownMenuItem className="text-red-600">
                        <Trash2 className="w-3.5 h-3.5 mr-2" />
                        <span>Delete Permanently</span>
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </div>
              </div>
            ))}
          </div>

          {activeFiles.length === 0 && (
            <div className="py-32 text-center flex flex-col items-center gap-4">
              <Folder className="w-16 h-16 text-muted-foreground/10" />
              <div className="space-y-1">
                <p className="text-muted-foreground font-black uppercase text-xs tracking-widest">No Files Found</p>
                <p className="text-[11px] text-muted-foreground italic">
                  {showArchived ? "Your archive is empty." : "Start by uploading files in a channel."}
                </p>
              </div>
            </div>
          )}
        </div>
      </ScrollArea>
    </div>
  )
}
