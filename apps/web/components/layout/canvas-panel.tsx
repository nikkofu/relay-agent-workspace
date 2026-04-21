"use client"

import { useState, useEffect } from "react"
import { useUIStore } from "@/stores/ui-store"
import { useArtifactStore, ArtifactVersion } from "@/stores/artifact-store"
import { useChannelStore } from "@/stores/channel-store"
import { useUserStore } from "@/stores/user-store"
import { X, Maximize2, RotateCcw, Share2, Save, Wand2, History, MessageSquare, Copy, Code, Type, ExternalLink, MoreVertical, ChevronLeft, Loader2, GitCompare, Sparkles } from "lucide-react"
import { Button } from "@/components/ui/button"
import { ScrollArea } from "@/components/ui/scroll-area"
import { cn } from "@/lib/utils"
import { UserAvatar } from "@/components/common/user-avatar"
import { formatDistanceToNow } from "date-fns"
import { ArtifactDiffView } from "./artifact-diff-view"

export function CanvasPanel() {
  const { isCanvasOpen, closeCanvas, activeCanvasId } = useUIStore()
  const { 
    activeArtifact, 
    fetchArtifactDetail, 
    createArtifact,
    createArtifactFromTemplate,
    updateArtifact, 
    restoreVersion,
    isLoading,
    versions,
    fetchVersions,
    isHistoryLoading,
    fetchVersionDetail,
    currentDiff,
    fetchDiff,
    isDiffLoading,
    clearDiff,
    references,
    fetchReferences,
    isReferencesLoading,
    templates,
    fetchTemplates,
    isTemplatesLoading
  } = useArtifactStore()
  const { currentChannel, setCurrentChannelById } = useChannelStore()
  const { currentUser } = useUserStore()
  
  const [content, setContent] = useState("")
  const [isEditing, setIsEditing] = useState(false)
  const [showHistory, setShowHistory] = useState(false)
  const [showReferences, setShowReferences] = useState(false)
  const [showTemplates, setShowTemplates] = useState(false)
  const [selectedVersion, setSelectedVersion] = useState<ArtifactVersion | null>(null)
  const [compareMode, setCompareMode] = useState(false)
  const [compareVersions, setCompareVersions] = useState<number[]>([])

  useEffect(() => {
    if (activeCanvasId) {
      if (activeCanvasId === "new-doc") {
        fetchTemplates()
        setShowTemplates(true)
      }
      fetchArtifactDetail(activeCanvasId)
      fetchReferences(activeCanvasId)
      setSelectedVersion(null)
      setShowHistory(false)
      setShowReferences(false)
      setCompareMode(false)
      setCompareVersions([])
      clearDiff()
    }
  }, [activeCanvasId, fetchArtifactDetail, fetchReferences, fetchTemplates, clearDiff])

  useEffect(() => {
    if (activeArtifact && !selectedVersion && !currentDiff) {
      setContent(activeArtifact.content)
    }
  }, [activeArtifact, selectedVersion, currentDiff])

  const handleSave = async () => {
    if (activeArtifact) {
      if (activeArtifact.id === "new-doc") {
        const created = await createArtifact({
          title: activeArtifact.title || "Untitled Canvas",
          content,
          type: activeArtifact.type || "document",
          channelId: currentChannel?.id || "ch-1",
        })
        if (created) {
          setIsEditing(false)
        }
        return
      }
      await updateArtifact(activeArtifact.id, { content })
      setIsEditing(false)
    }
  }

  const handleToggleHistory = () => {
    const nextShow = !showHistory
    setShowHistory(nextShow)
    if (nextShow) {
      setShowReferences(false)
      if (activeArtifact) fetchVersions(activeArtifact.id)
    }
    if (!nextShow) {
      setSelectedVersion(null)
      setCompareMode(false)
      setCompareVersions([])
      clearDiff()
    }
  }

  const handleToggleReferences = () => {
    const nextShow = !showReferences
    setShowReferences(nextShow)
    if (nextShow) {
      setShowHistory(false)
      if (activeArtifact) fetchReferences(activeArtifact.id)
    }
  }

  const handleReferenceClick = (channelId: string) => {
    setCurrentChannelById(channelId)
    // The message area will automatically update
  }

  const handleSelectTemplate = async (templateId: string) => {
    if (currentChannel && currentUser) {
      const created = await createArtifactFromTemplate(templateId, currentChannel.id, currentUser.id)
      if (created) {
        setShowTemplates(false)
      }
    }
  }

  const handleVersionClick = async (versionNum: number) => {
    if (!activeArtifact) return

    if (compareMode) {
      // Logic for comparing two versions
      let newCompare = [...compareVersions]
      if (newCompare.includes(versionNum)) {
        newCompare = newCompare.filter(v => v !== versionNum)
      } else {
        if (newCompare.length >= 2) newCompare.shift()
        newCompare.push(versionNum)
        newCompare.sort((a, b) => a - b)
      }
      setCompareVersions(newCompare)
      
      if (newCompare.length === 2) {
        fetchDiff(activeArtifact.id, newCompare[0], newCompare[1])
      } else {
        clearDiff()
      }
    } else {
      // Normal preview mode
      const detail = await fetchVersionDetail(activeArtifact.id, versionNum)
      if (detail) {
        setSelectedVersion(detail)
        setContent(detail.content)
        clearDiff()
      }
    }
  }

  const handleRestore = async () => {
    if (selectedVersion && activeArtifact) {
      await restoreVersion(activeArtifact.id, selectedVersion.version)
      setSelectedVersion(null)
      setShowHistory(false)
    }
  }

  if (!isCanvasOpen || !activeArtifact) return null

  const displayArtifact = selectedVersion || activeArtifact
  const isComparing = currentDiff !== null

  return (
    <div className="w-full h-full flex flex-col bg-white dark:bg-[#1a1d21] border-l shadow-2xl relative overflow-hidden animate-in slide-in-from-right duration-300">
      {/* Template Picker Overlay */}
      {showTemplates && activeArtifact?.id === "new-doc" && (
        <div className="absolute inset-0 z-50 bg-white dark:bg-[#1a1d21] flex flex-col">
          <header className="h-14 px-6 flex items-center justify-between border-b shrink-0">
            <div className="flex items-center gap-2">
              <Sparkles className="w-5 h-5 text-purple-600" />
              <h1 className="text-lg font-black tracking-tight uppercase">New Canvas</h1>
            </div>
            <Button variant="ghost" size="icon" className="rounded-full" onClick={closeCanvas}>
              <X className="w-5 h-5" />
            </Button>
          </header>
          
          <ScrollArea className="flex-1">
            <div className="p-10 max-w-4xl mx-auto space-y-10">
              <div className="space-y-1 text-center">
                <h2 className="text-3xl font-black tracking-tight">How would you like to start?</h2>
                <p className="text-muted-foreground">Select a template or start from a blank canvas.</p>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {isTemplatesLoading ? (
                  Array.from({ length: 4 }).map((_, i) => (
                    <div key={i} className="h-32 rounded-xl bg-muted animate-pulse" />
                  ))
                ) : (
                  <>
                    <button
                      onClick={() => { setShowTemplates(false); setIsEditing(true); }}
                      className="group p-6 text-left border-2 border-dashed rounded-2xl hover:border-purple-500/50 hover:bg-purple-500/5 transition-all flex flex-col justify-between h-40"
                    >
                      <div className="w-10 h-10 rounded-lg bg-muted flex items-center justify-center group-hover:bg-purple-100 dark:group-hover:bg-purple-900/20 group-hover:text-purple-600 transition-colors">
                        <PlusBadge className="w-5 h-5" />
                      </div>
                      <div>
                        <p className="font-black text-sm uppercase tracking-tight">Blank Document</p>
                        <p className="text-xs text-muted-foreground mt-1">Start with a clean slate</p>
                      </div>
                    </button>

                    {templates.map(template => (
                      <button
                        key={template.id}
                        onClick={() => handleSelectTemplate(template.id)}
                        className="group p-6 text-left border rounded-2xl hover:border-blue-500/50 hover:bg-blue-500/5 transition-all flex flex-col justify-between h-40"
                      >
                        <div className="w-10 h-10 rounded-lg bg-blue-500/10 flex items-center justify-center text-blue-600">
                          {template.type === 'code' ? <Code className="w-5 h-5" /> : <Type className="w-5 h-5" />}
                        </div>
                        <div>
                          <p className="font-black text-sm uppercase tracking-tight">{template.title}</p>
                          <p className="text-xs text-muted-foreground mt-1 line-clamp-2">{template.description}</p>
                        </div>
                      </button>
                    ))}
                  </>
                )}
              </div>
            </div>
          </ScrollArea>
        </div>
      )}
      {/* Header */}
      <header className="h-14 px-4 flex items-center justify-between border-b shrink-0 bg-muted/30">
        <div className="flex items-center gap-3 min-w-0">
          <div className="w-8 h-8 rounded-lg bg-blue-500/10 flex items-center justify-center shrink-0">
            {activeArtifact.type === 'code' ? <Code className="w-4 h-4 text-blue-600" /> : <Type className="w-4 h-4 text-blue-600" />}
          </div>
          <div className="flex flex-col min-w-0">
            <h3 className="font-bold text-sm truncate">{isComparing ? 'Comparison View' : displayArtifact.title}</h3>
            <div className="flex items-center gap-1.5">
              <span className="text-[10px] text-muted-foreground uppercase font-bold tracking-wider whitespace-nowrap">
                {isComparing ? `v${currentDiff.fromVersion} → v${currentDiff.toVersion}` : `v${displayArtifact.version} • ${activeArtifact.type}`}
              </span>
              {!isComparing && displayArtifact.updatedByUser && (
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
          {isComparing ? (
            <Button size="sm" variant="ghost" className="h-8 text-xs font-bold text-purple-600" onClick={() => { clearDiff(); setCompareVersions([]); }}>
              Exit Compare
            </Button>
          ) : selectedVersion ? (
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
            <Button size="sm" variant="ghost" className="h-8 font-bold text-[#1164a3]" onClick={() => setIsEditing(true)}>
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
          <Button 
            variant="ghost" 
            size="icon" 
            className={cn("h-7 w-7", showHistory && "text-purple-600 bg-purple-50 dark:bg-purple-900/20")} 
            onClick={handleToggleHistory}
            title="Version History"
          >
            <History className="w-3.5 h-3.5" />
          </Button>
          <Button 
            variant="ghost" 
            size="icon" 
            className={cn("h-7 w-7", showReferences && "text-blue-600 bg-blue-50 dark:bg-blue-900/20")} 
            onClick={handleToggleReferences}
            title="Backlinks"
          >
            <MessageSquare className="w-3.5 h-3.5" />
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
        {showHistory && (
          <div className="flex items-center gap-2 pl-1">
            <Button 
              variant="ghost" 
              size="sm" 
              className={cn("h-7 px-2 text-xs font-bold gap-1.5 transition-all", compareMode ? "text-purple-700 bg-purple-100" : "text-muted-foreground")}
              onClick={() => {
                setCompareMode(!compareMode)
                setCompareVersions([])
                clearDiff()
              }}
            >
              <GitCompare className="w-3.5 h-3.5" />
              {compareMode ? "Select two versions" : "Compare Versions"}
            </Button>
            {selectedVersion && !compareMode && (
              <Button variant="ghost" size="sm" className="h-7 px-2 text-xs text-muted-foreground" onClick={() => setSelectedVersion(null)}>
                <ChevronLeft className="w-3 h-3 mr-1" />
                Back to current
              </Button>
            )}
          </div>
        )}
        <div className="ml-auto flex items-center gap-1">
          <Button variant="ghost" size="icon" className="h-7 w-7"><Maximize2 className="w-3.5 h-3.5" /></Button>
          <Button variant="ghost" size="icon" className="h-7 w-7"><MoreVertical className="w-3.5 h-3.5" /></Button>
        </div>
      </div>

      {/* Main Content Area */}
      <div className="flex-1 min-h-0 relative flex">
        {/* Sidebar (History or References) */}
        {(showHistory || showReferences) && (
          <div className="w-64 border-r bg-muted/5 flex flex-col shrink-0">
            <header className="px-4 h-10 flex items-center justify-between border-b shrink-0">
              <span className="text-[10px] font-bold uppercase tracking-widest text-muted-foreground">
                {showHistory ? "Version History" : "Referencing Messages"}
              </span>
            </header>
            <ScrollArea className="flex-1">
              <div className="p-2 space-y-1">
                {showHistory ? (
                  isHistoryLoading ? (
                    <div className="flex items-center justify-center py-8">
                      <Loader2 className="w-4 h-4 animate-spin text-muted-foreground" />
                    </div>
                  ) : (
                    versions.map((v) => (
                      <button
                        key={v.version}
                        onClick={() => handleVersionClick(v.version)}
                        className={cn(
                          "w-full text-left p-3 rounded-lg transition-all group flex flex-col gap-1 border relative",
                          compareVersions.includes(v.version)
                            ? "bg-purple-100 border-purple-300 dark:bg-purple-900/40 dark:border-purple-700"
                            : (selectedVersion?.version === v.version || (!selectedVersion && v.version === activeArtifact.version && !compareMode))
                              ? "bg-purple-50 dark:bg-purple-900/20 border-purple-100 dark:border-purple-800"
                              : "hover:bg-muted/50 border-transparent"
                        )}
                      >
                        {compareVersions.includes(v.version) && (
                          <div className="absolute -left-1 top-1/2 -translate-y-1/2 w-2 h-4 bg-purple-500 rounded-r-full" />
                        )}
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
                  )
                ) : (
                  /* References List */
                  isReferencesLoading ? (
                    <div className="flex items-center justify-center py-8">
                      <Loader2 className="w-4 h-4 animate-spin text-muted-foreground" />
                    </div>
                  ) : references.length === 0 ? (
                    <div className="px-3 py-8 text-center text-[10px] text-muted-foreground italic leading-relaxed">
                      No messages referencing this artifact yet.
                    </div>
                  ) : (
                    references.map((ref, idx) => (
                      <button
                        key={idx}
                        onClick={() => handleReferenceClick(ref.channel?.id)}
                        className="w-full text-left p-3 rounded-lg transition-all border border-transparent hover:bg-muted/50 flex flex-col gap-1.5"
                      >
                        <div className="flex items-center justify-between">
                          <div className="flex items-center gap-1.5 min-w-0">
                            <UserAvatar src={ref.user?.avatar} name={ref.user?.name} className="h-3.5 w-3.5" />
                            <span className="text-xs font-bold truncate">{ref.user?.name}</span>
                          </div>
                          <span className="text-[8px] text-muted-foreground uppercase font-bold tracking-tighter">
                            #{ref.channel?.name}
                          </span>
                        </div>
                        <div className="text-[10px] text-foreground/70 line-clamp-2 italic leading-tight pl-5 border-l-2 border-muted" dangerouslySetInnerHTML={{ __html: ref.message?.content }} />
                      </button>
                    ))
                  )
                )}
              </div>
            </ScrollArea>
          </div>
        )}

        {/* Editor/Content Area / Diff View */}
        <div className="flex-1 min-h-0 relative flex flex-col">
          {(isLoading || isDiffLoading) ? (
            <div className="absolute inset-0 flex items-center justify-center bg-white/50 dark:bg-black/20 z-10">
              <div className="flex flex-col items-center gap-3">
                <Loader2 className="w-8 h-8 text-purple-500 animate-spin" />
                <span className="text-xs font-bold text-purple-600 animate-pulse uppercase tracking-widest">
                  {isDiffLoading ? "Calculating Differences..." : "Processing..."}
                </span>
              </div>
            </div>
          ) : null}

          {isComparing ? (
            <ArtifactDiffView diff={currentDiff} />
          ) : (
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
          )}
        </div>
      </div>

      {/* Footer / Context */}
      {!isComparing && (
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
      )}
    </div>
  )
}

function PlusBadge({ className }: { className?: string }) {
  return (
    <div className={cn("w-4 h-4 rounded-full border-2 border-current flex items-center justify-center font-bold text-[10px]", className)}>
      +
    </div>
  )
}
