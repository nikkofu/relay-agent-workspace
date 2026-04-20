"use client"

import { useState, useEffect } from "react"
import { useUIStore } from "@/stores/ui-store"
import { useArtifactStore, ArtifactVersion } from "@/stores/artifact-store"
import { X, Maximize2, RotateCcw, Share2, Save, Wand2, History, MessageSquare, Copy, Code, Type, ExternalLink, MoreVertical, ChevronLeft, Clock, Loader2 } from "lucide-react"
import { Button } from "@/components/ui/button"
import { ScrollArea } from "@/components/ui/scroll-area"
import { cn } from "@/lib/utils"
import { UserAvatar } from "@/components/common/user-avatar"
import { formatDistanceToNow } from "date-fns"

export function CanvasPanel() {
  const { isCanvasOpen, closeCanvas, activeCanvasId } = useUIStore()
  const { 
    activeArtifact, 
    fetchArtifactDetail, 
    updateArtifact, 
    isLoading,
    versions,
    fetchVersions,
    isHistoryLoading,
    fetchVersionDetail
  } = useArtifactStore()
  
  const [content, setContent] = useState("")
  const [isEditing, setIsEditing] = useState(false)
  const [showHistory, setShowHistory] = useState(false)
  const [selectedVersion, setSelectedVersion] = useState<ArtifactVersion | null>(null)

  useEffect(() => {
    if (activeCanvasId) {
      fetchArtifactDetail(activeCanvasId)
      setSelectedVersion(null)
      setShowHistory(false)
    }
  }, [activeCanvasId, fetchArtifactDetail])

  useEffect(() => {
    if (activeArtifact && !selectedVersion) {
      setContent(activeArtifact.content)
    }
  }, [activeArtifact, selectedVersion])

  const handleSave = async () => {
    if (activeArtifact) {
      await updateArtifact(activeArtifact.id, { content })
      setIsEditing(false)
    }
  }

  const handleToggleHistory = () => {
    const nextShow = !showHistory
    setShowHistory(nextShow)
    if (nextShow && activeArtifact) {
      fetchVersions(activeArtifact.id)
    }
    if (!nextShow) {
      setSelectedVersion(null)
    }
  }

  const handleVersionClick = async (versionNum: number) => {
    if (activeArtifact) {
      const detail = await fetchVersionDetail(activeArtifact.id, versionNum)
      if (detail) {
        setSelectedVersion(detail)
        setContent(detail.content)
      }
    }
  }

  const handleRestore = async () => {
    if (selectedVersion && activeArtifact) {
      await updateArtifact(activeArtifact.id, { 
        content: selectedVersion.content,
        title: selectedVersion.title 
      })
      setSelectedVersion(null)
      setShowHistory(false)
    }
  }

  if (!isCanvasOpen || !activeArtifact) return null

  const displayArtifact = selectedVersion || activeArtifact

  return (
    <div className="w-full h-full flex flex-col bg-white dark:bg-[#1a1d21] border-l shadow-2xl relative overflow-hidden animate-in slide-in-from-right duration-300">
      {/* Header */}
      <header className="h-14 px-4 flex items-center justify-between border-b shrink-0 bg-muted/30">
        <div className="flex items-center gap-3 min-w-0">
          <div className="w-8 h-8 rounded-lg bg-blue-500/10 flex items-center justify-center shrink-0">
            {activeArtifact.type === 'code' ? <Code className="w-4 h-4 text-blue-600" /> : <Type className="w-4 h-4 text-blue-600" />}
          </div>
          <div className="flex flex-col min-w-0">
            <h3 className="font-bold text-sm truncate">{displayArtifact.title}</h3>
            <div className="flex items-center gap-1.5">
              <span className="text-[10px] text-muted-foreground uppercase font-bold tracking-wider whitespace-nowrap">
                v{displayArtifact.version} • {activeArtifact.type}
              </span>
              {displayArtifact.updatedByUser && (
                <>
                  <span className="text-[10px] text-muted-foreground">•</span>
                  <div className="flex items-center gap-1 min-w-0">
                    <UserAvatar src={displayArtifact.updatedByUser.avatar} name={displayArtifact.updatedByUser.name} className="h-3 w-3" />
                    <span className="text-[10px] text-muted-foreground truncate italic opacity-80">
                      {selectedVersion ? 'Version by' : 'Edited by'} {displayArtifact.updatedByUser.name}
                    </span>
                  </div>
                </>
              )}
            </div>
          </div>
        </div>
        <div className="flex items-center gap-1">
          {selectedVersion ? (
            <Button size="sm" variant="default" className="h-8 bg-purple-600 hover:bg-purple-700 text-white font-bold" onClick={handleRestore}>
              <RotateCcw className="w-3.5 h-3.5 mr-2" />
              Restore this version
            </Button>
          ) : isEditing ? (
            <Button size="sm" variant="default" className="h-8 bg-blue-600 hover:bg-blue-700 text-white" onClick={handleSave} disabled={isLoading}>
              <Save className="w-3.5 h-3.5 mr-2" />
              Save
            </Button>
          ) : (
            <Button size="sm" variant="ghost" className="h-8" onClick={() => setIsEditing(true)}>
              Edit
            </Button>
          )}
          <Button variant="ghost" size="icon" className="h-8 w-8 rounded-full">
            <Share2 className="w-4 h-4" />
          </Button>
          <Button variant="ghost" size="icon" className="h-8 w-8 rounded-full" onClick={closeCanvas}>
            <X className="w-4 h-4" />
          </Button>
        </div>
      </header>

      {/* Toolbar */}
      <div className="h-10 px-4 flex items-center gap-2 border-b bg-muted/10">
        <div className="flex items-center gap-1 pr-2 border-r">
          <Button variant="ghost" size="icon" className={cn("h-7 w-7", showHistory && "text-purple-600 bg-purple-50 dark:bg-purple-900/20")} onClick={handleToggleHistory}>
            <History className="w-3.5 h-3.5" />
          </Button>
        </div>
        {!showHistory && (
          <div className="flex items-center gap-1 pl-1">
            <Button variant="ghost" size="sm" className="h-7 px-2 text-xs font-bold text-purple-600 bg-purple-50 dark:bg-purple-900/20">
              <Wand2 className="w-3 h-3 mr-1.5" />
              AI Fix
            </Button>
            <Button variant="ghost" size="sm" className="h-7 px-2 text-xs font-medium">
              <Copy className="w-3 h-3 mr-1.5" />
              Copy
            </Button>
          </div>
        )}
        {showHistory && selectedVersion && (
          <Button variant="ghost" size="sm" className="h-7 px-2 text-xs text-muted-foreground" onClick={() => setSelectedVersion(null)}>
            <ChevronLeft className="w-3 h-3 mr-1" />
            Back to current
          </Button>
        )}
        <div className="ml-auto flex items-center gap-1">
          <Button variant="ghost" size="icon" className="h-7 w-7"><Maximize2 className="w-3.5 h-3.5" /></Button>
          <Button variant="ghost" size="icon" className="h-7 w-7"><MoreVertical className="w-3.5 h-3.5" /></Button>
        </div>
      </div>

      {/* Main Content Area */}
      <div className="flex-1 min-h-0 relative flex">
        {/* Version History Sidebar */}
        {showHistory && (
          <div className="w-64 border-r bg-muted/5 flex flex-col shrink-0">
            <header className="px-4 h-10 flex items-center border-b shrink-0">
              <span className="text-[10px] font-bold uppercase tracking-widest text-muted-foreground">Version History</span>
            </header>
            <ScrollArea className="flex-1">
              <div className="p-2 space-y-1">
                {isHistoryLoading ? (
                  <div className="flex items-center justify-center py-8">
                    <Loader2 className="w-4 h-4 animate-spin text-muted-foreground" />
                  </div>
                ) : (
                  versions.map((v) => (
                    <button
                      key={v.version}
                      onClick={() => handleVersionClick(v.version)}
                      className={cn(
                        "w-full text-left p-3 rounded-lg transition-all group flex flex-col gap-1",
                        (selectedVersion?.version === v.version || (!selectedVersion && v.version === activeArtifact.version))
                          ? "bg-purple-50 dark:bg-purple-900/20 border-purple-100 dark:border-purple-800"
                          : "hover:bg-muted/50 border-transparent"
                      )}
                    >
                      <div className="flex items-center justify-between">
                        <span className="text-xs font-bold">Version {v.version}</span>
                        <span className="text-[9px] text-muted-foreground">
                          {formatDistanceToNow(new Date(v.updatedAt), { addSuffix: true })}
                        </span>
                      </div>
                      {v.updatedByUser && (
                        <div className="flex items-center gap-1.5 mt-0.5">
                          <UserAvatar src={v.updatedByUser.avatar} name={v.updatedByUser.name} className="h-3 w-3" />
                          <span className="text-[9px] text-muted-foreground truncate">{v.updatedByUser.name}</span>
                        </div>
                      )}
                    </button>
                  ))
                )}
              </div>
            </ScrollArea>
          </div>
        )}

        {/* Editor/Content Area */}
        <div className="flex-1 min-h-0 relative flex flex-col">
          {isLoading ? (
            <div className="absolute inset-0 flex items-center justify-center bg-white/50 dark:bg-black/20 z-10">
              <div className="flex flex-col items-center gap-3">
                <Wand2 className="w-8 h-8 text-purple-500 animate-bounce" />
                <span className="text-xs font-bold text-purple-600 animate-pulse uppercase tracking-widest">Processing...</span>
              </div>
            </div>
          ) : null}

          <ScrollArea className="flex-1">
            <div className="p-8 max-w-3xl mx-auto">
              {isEditing && !selectedVersion ? (
                <textarea
                  className="w-full min-h-[500px] bg-transparent border-none outline-none resize-none font-mono text-sm leading-relaxed"
                  value={content}
                  onChange={(e) => setContent(e.target.value)}
                  spellCheck={false}
                />
              ) : (
                <div 
                  className={cn(
                    "prose dark:prose-invert max-w-none",
                    activeArtifact.type === 'code' && "font-mono text-sm bg-slate-50 dark:bg-slate-900/50 p-6 rounded-xl border border-slate-100 dark:border-slate-800 whitespace-pre"
                  )}
                  dangerouslySetInnerHTML={{ __html: content }}
                />
              )}
            </div>
          </ScrollArea>
        </div>
      </div>

      {/* Footer / Context */}
      <footer className="h-12 px-4 border-t flex items-center justify-between shrink-0 bg-muted/5">
        <div className="flex items-center gap-2 text-[10px] font-bold text-muted-foreground uppercase tracking-widest">
          <MessageSquare className="w-3 h-3" />
          Refers to #general discussion
        </div>
        <Button variant="ghost" size="sm" className="h-8 text-[10px] font-bold uppercase tracking-widest">
          Open in Tab
          <ExternalLink className="w-3 h-3 ml-1.5" />
        </Button>
      </footer>
    </div>
  )
}
