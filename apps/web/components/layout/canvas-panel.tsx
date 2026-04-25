"use client"

import { useState, useEffect, useRef, useMemo } from "react"
import { useUIStore } from "@/stores/ui-store"
import { useArtifactStore, ArtifactVersion } from "@/stores/artifact-store"
import { useChannelStore } from "@/stores/channel-store"
import { useUserStore } from "@/stores/user-store"
import {
  X, Maximize2, Minimize2, RotateCcw, Share2, Save, Wand2, History, MessageSquare,
  Copy, Code, Type, ExternalLink, MoreVertical, ChevronLeft, Loader2, GitCompare,
  Sparkles, Pencil, Link as LinkIcon, FileDown, FileText,
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { ScrollArea } from "@/components/ui/scroll-area"
import { cn } from "@/lib/utils"
import { UserAvatar } from "@/components/common/user-avatar"
import { formatDistanceToNow, format } from "date-fns"
import { ArtifactDiffView } from "./artifact-diff-view"
import { CanvasTipTapEditor, type CanvasEditorHandle } from "./canvas-tiptap-editor"
import { CanvasAIDock, type CanvasAIDockHandle, type CanvasAIDockLayout } from "./canvas-ai-dock"
import {
  DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem,
  DropdownMenuSeparator, DropdownMenuLabel,
} from "@/components/ui/dropdown-menu"
import { toast } from "sonner"
import { exportArtifactAsPDF, exportArtifactAsWord } from "@/lib/export-artifact"

export function CanvasPanel() {
  const {
    isCanvasOpen, closeCanvas, activeCanvasId,
    isCanvasEditing, setCanvasEditing,
    isCanvasMaximized, toggleCanvasMaximized,
  } = useUIStore()
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
    isTemplatesLoading,
    duplicateArtifact
  } = useArtifactStore()
  const { currentChannel, setCurrentChannelById } = useChannelStore()
  const { currentUser } = useUserStore()
  
  const [content, setContent] = useState("")
  const [isEditing, setIsEditing] = useState(false)

  // Inline title rename state. The title is shown as a label by default and
  // flips to a text input when clicked (or when the pencil is hit). `Enter`
  // commits, `Escape` cancels; blur commits as well so the workflow is
  // forgiving.
  const [isRenaming, setIsRenaming] = useState(false)
  const [titleDraft, setTitleDraft] = useState("")
  const titleInputRef = useRef<HTMLInputElement>(null)

  // Imperative handles for the rich editor + AI dock so the toolbar's "Ask AI"
  // button can focus the dock composer without lifting editor state up.
  const editorRef = useRef<CanvasEditorHandle>(null)
  const dockRef = useRef<CanvasAIDockHandle>(null)

  // Persisted AI-dock layout preference so it survives reloads / re-opens.
  // "bottom" (default) keeps the composer compact under the editor; "rail"
  // pulls it out to a full-height right sidebar — both editor + chat then
  // have their own dedicated scroll areas (the #5 request).
  const [aiDockLayout, setAIDockLayout] = useState<CanvasAIDockLayout>("bottom")
  useEffect(() => {
    if (typeof window === "undefined") return
    const saved = window.localStorage.getItem("canvas-ai-dock-layout")
    if (saved === "rail" || saved === "bottom") setAIDockLayout(saved)
  }, [])
  const handleLayoutChange = (next: CanvasAIDockLayout) => {
    setAIDockLayout(next)
    if (typeof window !== "undefined") {
      window.localStorage.setItem("canvas-ai-dock-layout", next)
    }
  }
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
          setCanvasEditing(false)
        }
        return
      }
      await updateArtifact(activeArtifact.id, { content })
      setIsEditing(false)
      setCanvasEditing(false)
    }
  }

  // Enter edit mode + broadcast to the workspace layout so it can widen the
  // right side to give the editor + AI chat dock a 33/33/33 split with the
  // message column (request #10).
  const handleEnterEdit = () => {
    setIsEditing(true)
    setCanvasEditing(true)
  }

  // ── Inline title rename ─────────────────────────────────────────────────
  const handleStartRename = () => {
    if (!activeArtifact || activeArtifact.id === "new-doc") return
    setTitleDraft(activeArtifact.title || "")
    setIsRenaming(true)
    // Defer to let the input mount before focusing.
    setTimeout(() => {
      titleInputRef.current?.focus()
      titleInputRef.current?.select()
    }, 0)
  }
  const handleCommitRename = async () => {
    if (!activeArtifact || activeArtifact.id === "new-doc") {
      setIsRenaming(false)
      return
    }
    const next = titleDraft.trim()
    const prev = activeArtifact.title || ""
    if (!next || next === prev) {
      setIsRenaming(false)
      return
    }
    await updateArtifact(activeArtifact.id, { title: next })
    setIsRenaming(false)
  }
  const handleCancelRename = () => {
    setIsRenaming(false)
    setTitleDraft("")
  }

  // ── Deep link + Open in new tab ─────────────────────────────────────────
  // Build a shareable URL that loads the workspace on this channel and
  // auto-opens the canvas via the `canvas` query param (the workspace page
  // reads it on mount).
  const canvasDeepLink = useMemo(() => {
    if (typeof window === "undefined" || !activeArtifact) return ""
    const url = new URL(window.location.origin + "/workspace")
    if (activeArtifact.channelId || currentChannel?.id) {
      url.searchParams.set("c", activeArtifact.channelId || currentChannel?.id || "")
    }
    url.searchParams.set("canvas", activeArtifact.id)
    return url.toString()
  }, [activeArtifact, currentChannel])

  const handleOpenInTab = () => {
    if (!canvasDeepLink) return
    window.open(canvasDeepLink, "_blank", "noopener,noreferrer")
  }

  const handleCopyLink = async () => {
    if (!canvasDeepLink) return
    try {
      await navigator.clipboard.writeText(canvasDeepLink)
      toast.success("Canvas link copied to clipboard")
    } catch {
      toast.error("Copy failed\u00a0— clipboard permission denied")
    }
  }

  const handleShareInChannel = () => {
    // Future: POST an inline canvas reference into the current channel's
    // composer. For now fall back to clipboard so the user can paste the
    // link into any conversation.
    handleCopyLink()
    toast.info("Paste the link into any channel or DM to share")
  }

  // ── Export ──────────────────────────────────────────────────────────────
  const handleExportWord = () => {
    if (!activeArtifact) return
    exportArtifactAsWord({ title: activeArtifact.title || "Untitled", html: content })
    toast.success("Exported as Word (.doc)")
  }
  const handleExportPDF = () => {
    if (!activeArtifact) return
    const ok = exportArtifactAsPDF({ title: activeArtifact.title || "Untitled", html: content })
    if (!ok) toast.error("Popup blocked — please allow popups for PDF export")
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

  const handleFork = async () => {
    if (selectedVersion && activeArtifact) {
      const created = await createArtifact({
        title: `${selectedVersion.title} (Fork)`,
        content: selectedVersion.content,
        type: activeArtifact.type,
        channelId: currentChannel?.id || activeArtifact.channelId,
      })
      if (created) {
        setSelectedVersion(null)
        setShowHistory(false)
      }
    }
  }

  const handleDuplicate = async () => {
    if (activeArtifact) {
      await duplicateArtifact(activeArtifact.id)
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
            {/* Inline rename (bug #1). Click the title or the pencil icon
                to edit; Enter / blur commits, Escape cancels. Disabled while
                comparing, previewing a version, or for unsaved placeholders. */}
            {isComparing ? (
              <h3 className="font-bold text-sm truncate">Comparison View</h3>
            ) : isRenaming ? (
              <input
                ref={titleInputRef}
                value={titleDraft}
                onChange={(e) => setTitleDraft(e.target.value)}
                onBlur={handleCommitRename}
                onKeyDown={(e) => {
                  if (e.key === "Enter") { e.preventDefault(); handleCommitRename() }
                  if (e.key === "Escape") { e.preventDefault(); handleCancelRename() }
                }}
                placeholder="Untitled"
                className="font-bold text-sm bg-transparent outline-none border-b border-purple-500/60 focus:border-purple-600 px-0.5 py-px w-full max-w-[420px]"
                maxLength={120}
              />
            ) : (
              <div className="flex items-center gap-1 group/title min-w-0">
                <button
                  type="button"
                  onClick={handleStartRename}
                  className="font-bold text-sm truncate text-left hover:text-purple-600 transition-colors"
                  title={activeArtifact.id === "new-doc" ? "Save the canvas before renaming" : "Click to rename"}
                  disabled={activeArtifact.id === "new-doc"}
                >
                  {displayArtifact.title || "Untitled"}
                </button>
                {activeArtifact.id !== "new-doc" && !selectedVersion && (
                  <Pencil
                    className="w-3 h-3 text-muted-foreground/40 opacity-0 group-hover/title:opacity-100 transition-opacity cursor-pointer hover:text-purple-600"
                    onClick={handleStartRename}
                  />
                )}
              </div>
            )}
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
            <div className="flex items-center gap-2">
              <Button size="sm" variant="outline" className="h-8 border-purple-200 dark:border-purple-800 text-purple-600 hover:bg-purple-50 dark:hover:bg-purple-900/20 font-bold" onClick={handleFork} disabled={isLoading}>
                <PlusBadge className="w-3.5 h-3.5 mr-2" />
                Fork as new
              </Button>
              <Button size="sm" variant="default" className="h-8 bg-purple-600 hover:bg-purple-700 text-white font-bold" onClick={handleRestore} disabled={isLoading}>
                <RotateCcw className="w-3.5 h-3.5 mr-2" />
                Restore this version
              </Button>
            </div>
          ) : isEditing ? (
            <Button size="sm" variant="default" className="h-8 bg-blue-600 hover:bg-blue-700 text-white" onClick={handleSave} disabled={isLoading}>
              <Save className="w-3.5 h-3.5 mr-2" />
              Save
            </Button>
          ) : (
            <Button size="sm" variant="ghost" className="h-8 font-bold text-[#1164a3]" onClick={handleEnterEdit}>
              Edit
            </Button>
          )}
          {/* Share dropdown (bug #5) */}
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon" className="h-8 w-8 rounded-full" title="Share">
                <Share2 className="w-4 h-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-56">
              <DropdownMenuLabel className="text-[9px] uppercase tracking-widest">Share canvas</DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem onClick={handleCopyLink}>
                <LinkIcon className="w-3.5 h-3.5 mr-2" />
                Copy link
              </DropdownMenuItem>
              <DropdownMenuItem onClick={handleShareInChannel}>
                <MessageSquare className="w-3.5 h-3.5 mr-2" />
                Share in channel…
              </DropdownMenuItem>
              <DropdownMenuItem onClick={handleOpenInTab}>
                <ExternalLink className="w-3.5 h-3.5 mr-2" />
                Open in new tab
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
          {/* Maximize / minimize (bug #6) */}
          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 rounded-full"
            onClick={toggleCanvasMaximized}
            title={isCanvasMaximized ? "Restore size" : "Maximize canvas"}
          >
            {isCanvasMaximized ? <Minimize2 className="w-4 h-4" /> : <Maximize2 className="w-4 h-4" />}
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
            <Button 
              variant="ghost" 
              size="sm" 
              className="h-7 px-2 text-xs font-medium"
              onClick={handleDuplicate}
              disabled={isLoading}
            >
              <Copy className="w-3 h-3 mr-1.5" />
              Duplicate
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
          <Button
            variant="ghost"
            size="icon"
            className="h-7 w-7"
            onClick={toggleCanvasMaximized}
            title={isCanvasMaximized ? "Restore size" : "Maximize canvas"}
          >
            {isCanvasMaximized ? <Minimize2 className="w-3.5 h-3.5" /> : <Maximize2 className="w-3.5 h-3.5" />}
          </Button>
          {/* MoreVertical menu (bug #7) — Duplicate / Export / Delete / Open-in-Tab */}
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon" className="h-7 w-7" title="More actions">
                <MoreVertical className="w-3.5 h-3.5" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-52">
              <DropdownMenuLabel className="text-[9px] uppercase tracking-widest">Actions</DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem onClick={handleDuplicate} disabled={isLoading}>
                <Copy className="w-3.5 h-3.5 mr-2" />
                Duplicate
              </DropdownMenuItem>
              <DropdownMenuItem onClick={handleOpenInTab}>
                <ExternalLink className="w-3.5 h-3.5 mr-2" />
                Open in new tab
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuLabel className="text-[9px] uppercase tracking-widest">Export</DropdownMenuLabel>
              <DropdownMenuItem onClick={handleExportWord}>
                <FileText className="w-3.5 h-3.5 mr-2" />
                Export as Word (.doc)
              </DropdownMenuItem>
              <DropdownMenuItem onClick={handleExportPDF}>
                <FileDown className="w-3.5 h-3.5 mr-2" />
                Export as PDF
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
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
                          <span
                            className="text-[9px] text-muted-foreground"
                            title={v.updatedAt ? safeFormatAbsolute(v.updatedAt) : undefined}
                          >
                            {v.updatedAt ? formatDistanceToNow(new Date(v.updatedAt), { addSuffix: true }) : '—'}
                          </span>
                        </div>
                        {/* Absolute date+time under the version label — always visible for clarity */}
                        {v.updatedAt && (
                          <span className="text-[9px] text-muted-foreground/70 tabular-nums">
                            {safeFormatAbsolute(v.updatedAt)}
                          </span>
                        )}
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
            <ArtifactDiffView
              diff={currentDiff}
              fromVersionMeta={versions.find(v => v.version === currentDiff.fromVersion)}
              toVersionMeta={versions.find(v => v.version === currentDiff.toVersion)}
            />
          ) : (() => {
            const dockVisible = isEditing && !selectedVersion && activeArtifact.type !== "code"
            // While the global edit mode is on, the workspace layout widens
            // the canvas panel to ~67% — so we also pin the dock to rail and
            // give it 50% of the canvas (yielding the 33/33/33 split with
            // the message column, request #10). Outside edit mode we honour
            // the user's saved bottom/rail preference.
            const forceRail = isCanvasEditing && dockVisible
            const useRail = dockVisible && (aiDockLayout === "rail" || forceRail)
            return (
              <div className={cn(
                "flex-1 min-h-0 flex",
                useRail ? "flex-row" : "flex-col",
              )}>
                <ScrollArea className={cn(useRail ? "flex-1 min-w-0" : "flex-1")}>
                  <div className="p-8 max-w-3xl mx-auto">
                    {isEditing && !selectedVersion ? (
                      activeArtifact.type === "code" ? (
                        <textarea
                          className="w-full min-h-[500px] bg-transparent border-none outline-none resize-none font-mono text-sm leading-relaxed"
                          value={content}
                          onChange={(e) => setContent(e.target.value)}
                          spellCheck={false}
                        />
                      ) : (
                        <CanvasTipTapEditor
                          ref={editorRef}
                          content={content}
                          onChange={setContent}
                          channelId={activeArtifact.channelId || currentChannel?.id || ""}
                          autoFocus
                          onRequestAIEdit={() => dockRef.current?.focusInput()}
                        />
                      )
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
                {/* Persistent AI chat dock — only while actively editing a
                    doc-type canvas. `layout="rail"` pulls it out as a full-height
                    sidebar so editor + chat each have dedicated scroll areas
                    (request #5). `layout="bottom"` keeps the compact bottom dock. */}
                {dockVisible && (
                  <div className={cn(
                    useRail
                      ? (forceRail ? "flex-1 min-w-0 border-l" : "w-[400px] shrink-0")
                      : "shrink-0",
                  )}> 
                    <CanvasAIDock
                      ref={dockRef}
                      editorRef={editorRef}
                      channelId={activeArtifact.channelId || currentChannel?.id || ""}
                      artifactId={activeArtifact.id}
                      artifactTitle={activeArtifact.title}
                      layout={aiDockLayout}
                      onLayoutChange={handleLayoutChange}
                    />
                  </div>
                )}
              </div>
            )
          })()}
        </div>
      </div>

      {/* Footer / Context */}
      {!isComparing && (
        <footer className="h-12 px-4 border-t flex items-center justify-between shrink-0 bg-muted/5">
          <div className="flex items-center gap-2 text-[10px] font-bold text-muted-foreground uppercase tracking-widest">
            <MessageSquare className="w-3 h-3" />
            Refers to #general discussion
          </div>
          <Button
            variant="ghost"
            size="sm"
            className="h-8 text-[10px] font-bold uppercase tracking-widest"
            onClick={handleOpenInTab}
            title="Open this canvas in a new browser tab"
          >
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

// Formats an ISO timestamp as "MMM d, HH:mm" safely. Empty string on bad input
// so the version-history row never renders "Invalid Date".
function safeFormatAbsolute(ts: string): string {
  try {
    const d = new Date(ts)
    if (isNaN(d.getTime())) return ""
    return format(d, "MMM d, HH:mm")
  } catch {
    return ""
  }
}
