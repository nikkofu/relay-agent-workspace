"use client"

import { useCallback, useEffect, useRef, useState } from "react"
import { useFileStore } from "@/stores/file-store"
import { Search, Folder, FileIcon, Archive, Trash2, Download, Filter, MoreVertical, RefreshCcw, History, ShieldCheck, Loader2, Star, MessageSquare, Share2, BookOpen, Tag, Send, Globe, CheckCircle2, Edit2, X, Cpu, AlertTriangle, AlignLeft, Layers, Link2, FileSearch, RefreshCw } from "lucide-react"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Badge } from "@/components/ui/badge"
import { Textarea } from "@/components/ui/textarea"
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs"
import Image from "next/image"
import { 
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger 
} from "@/components/ui/dropdown-menu"
import { format, formatDistanceToNow } from "date-fns"
import { useChannelStore } from "@/stores/channel-store"
import { useUserStore } from "@/stores/user-store"
import { 
  Select, 
  SelectContent, 
  SelectItem, 
  SelectTrigger, 
  SelectValue 
} from "@/components/ui/select"
import { 
  Dialog, 
  DialogContent, 
  DialogHeader, 
  DialogTitle, 
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog"
import { cn } from "@/lib/utils"
import { Label } from "@/components/ui/label"
import type { FileComment, FileShare, FileChunk, FileSearchResult, CitationEvidence } from "@/types"
import { CitationCard } from "@/components/citation/citation-card"
import { normalizeFilesPageFile, encodeFileDragPayload } from "@/lib/file-to-canvas"

export default function FilesPage() {
  const { 
    files, archivedFiles, starredFiles,
    fetchFiles, fetchArchivedFiles, archiveFile, deleteFile,
    fetchFileAuditHistory, updateFileRetention, fetchFilePreview,
    fetchFileComments, createFileComment, toggleFileStar, fetchStarredFiles,
    shareFile, fetchFileShares, updateFileKnowledge,
    fetchFileExtraction, rebuildFileExtraction, fetchFileExtractedContent,
    fetchFileChunks, fetchFileCitations, searchFiles,
  } = useFileStore()
  const { channels } = useChannelStore()
  const { users, fetchUsers } = useUserStore()
  const openedQueryFileIdRef = useRef<string | null>(null)
  
  const [q, setQ] = useState("")
  const [uploaderId, setUploaderId] = useState("all")
  const [contentType, setContentType] = useState("all")
  const [showArchived, setShowArchived] = useState(false)
  const [showStarred, setShowStarred] = useState(false)

  // Governance & Preview State
  const [isViewingAudit, setIsViewingAudit] = useState(false)
  const [auditLogs, setAuditLogs] = useState<any[]>([])
  const [selectedFile, setSelectedFile] = useState<any>(null)
  const [isUpdatingRetention, setIsUpdatingRetention] = useState(false)
  const [retentionDays, setRetentionDays] = useState("0")
  const [isViewingPreview, setIsViewingPreview] = useState(false)
  const [filePreview, setFilePreview] = useState<any>(null)
  const [isPreviewLoading, setIsPreviewLoading] = useState(false)
  const [previewTab, setPreviewTab] = useState("details")

  // Comments state
  const [comments, setComments] = useState<FileComment[]>([])
  const [commentInput, setCommentInput] = useState("")
  const [isPostingComment, setIsPostingComment] = useState(false)

  // Shares state
  const [shares, setShares] = useState<FileShare[]>([])
  const [isShareDialogOpen, setIsShareDialogOpen] = useState(false)
  const [shareChannelId, setShareChannelId] = useState("")
  const [shareComment, setShareComment] = useState("")
  const [isSharingFile, setIsSharingFile] = useState(false)

  // Content search state
  const [isContentSearch, setIsContentSearch] = useState(false)
  const [contentSearchQ, setContentSearchQ] = useState("")
  const [contentSearchResults, setContentSearchResults] = useState<FileSearchResult[]>([])
  const [isSearchingContent, setIsSearchingContent] = useState(false)

  // Indexing / extraction state
  const [extraction, setExtraction] = useState<any>(null)
  const [extractedText, setExtractedText] = useState<string | null>(null)
  const [chunks, setChunks] = useState<FileChunk[]>([])
  const [citations, setCitations] = useState<CitationEvidence[]>([])
  const [isLoadingExtraction, setIsLoadingExtraction] = useState(false)
  const [isRebuildingExtraction, setIsRebuildingExtraction] = useState(false)
  const [showExtractedText, setShowExtractedText] = useState(false)
  const [showChunks, setShowChunks] = useState(false)
  const [showCitations, setShowCitations] = useState(false)

  // Knowledge state
  const [isEditingKnowledge, setIsEditingKnowledge] = useState(false)
  const [knowledgeState, setKnowledgeState] = useState("")
  const [sourceKind, setSourceKind] = useState("")
  const [knowledgeSummary, setKnowledgeSummary] = useState("")
  const [knowledgeTags, setKnowledgeTags] = useState("")
  const [isSavingKnowledge, setIsSavingKnowledge] = useState(false)

  useEffect(() => {
    fetchUsers()
  }, [fetchUsers])

  useEffect(() => {
    if (showStarred) {
      fetchStarredFiles()
      return
    }
    const params: any = { q }
    if (uploaderId !== "all") params.uploaderId = uploaderId
    if (contentType !== "all") params.contentType = contentType

    if (showArchived) {
      fetchArchivedFiles(params)
    } else {
      fetchFiles(params)
    }
  }, [showArchived, showStarred, q, uploaderId, contentType, fetchFiles, fetchArchivedFiles, fetchStarredFiles])

  const activeFiles = showStarred ? starredFiles : showArchived ? archivedFiles : files
  const contentTypes = Array.from(new Set(activeFiles.map(f => f.type).filter(Boolean)))

  const handleViewAudit = async (file: any) => {
    setSelectedFile(file)
    const logs = await fetchFileAuditHistory(file.id)
    setAuditLogs(logs)
    setIsViewingAudit(true)
  }

  const handleViewPreview = useCallback(async (file: any) => {
    setSelectedFile(file)
    setIsPreviewLoading(true)
    setIsViewingPreview(true)
    setPreviewTab("details")
    setComments([])
    setShares([])
    setIsEditingKnowledge(false)
    setExtraction(null)
    setExtractedText(null)
    setChunks([])
    setCitations([])
    setShowExtractedText(false)
    setShowChunks(false)
    setShowCitations(false)
    const preview = await fetchFilePreview(file.id)
    setFilePreview(preview)
    setIsPreviewLoading(false)
    const [fetchedComments, fetchedShares] = await Promise.all([
      fetchFileComments(file.id),
      fetchFileShares(file.id),
    ])
    setComments(fetchedComments)
    setShares(fetchedShares)
    setKnowledgeState(file.knowledge_state || "")
    setSourceKind(file.source_kind || "")
    setKnowledgeSummary(file.summary || "")
    setKnowledgeTags((file.tags || []).join(", "))
  }, [fetchFilePreview, fetchFileComments, fetchFileShares])

  useEffect(() => {
    const queryFileId = new URLSearchParams(window.location.search).get("id")
    if (!queryFileId || openedQueryFileIdRef.current === queryFileId) return
    const file = activeFiles.find(f => f.id === queryFileId)
    if (!file) return
    openedQueryFileIdRef.current = queryFileId
    handleViewPreview(file)
  }, [activeFiles, handleViewPreview])

  const handleLoadExtraction = async (fileId: string) => {
    if (extraction) return
    setIsLoadingExtraction(true)
    const ext = await fetchFileExtraction(fileId)
    setExtraction(ext)
    setIsLoadingExtraction(false)
  }

  const handleRebuildExtraction = async (fileId: string) => {
    setIsRebuildingExtraction(true)
    await rebuildFileExtraction(fileId)
    setIsRebuildingExtraction(false)
    setExtraction(null)
    handleLoadExtraction(fileId)
  }

  const handleLoadExtractedText = async (fileId: string) => {
    const result = await fetchFileExtractedContent(fileId)
    setExtractedText(result?.text || null)
    setShowExtractedText(true)
  }

  const handleLoadChunks = async (fileId: string) => {
    const result = await fetchFileChunks(fileId)
    setChunks(result)
    setShowChunks(true)
  }

  const handleLoadCitations = async (fileId: string) => {
    const result = await fetchFileCitations(fileId)
    setCitations(result)
    setShowCitations(true)
  }

  const handleContentSearch = async (value: string) => {
    setContentSearchQ(value)
    if (!value.trim()) { setContentSearchResults([]); return }
    setIsSearchingContent(true)
    const results = await searchFiles(value.trim())
    setContentSearchResults(results)
    setIsSearchingContent(false)
  }

  const handlePostComment = async () => {
    if (!selectedFile || !commentInput.trim()) return
    setIsPostingComment(true)
    const newComment = await createFileComment(selectedFile.id, commentInput.trim())
    if (newComment) {
      setComments(prev => [...prev, newComment])
      setCommentInput("")
    }
    setIsPostingComment(false)
  }

  const handleShareFile = async () => {
    if (!selectedFile || !shareChannelId) return
    setIsSharingFile(true)
    const share = await shareFile(selectedFile.id, shareChannelId, shareComment || undefined)
    if (share) {
      setShares(prev => [...prev, share])
      setShareChannelId("")
      setShareComment("")
      setIsShareDialogOpen(false)
    }
    setIsSharingFile(false)
  }

  const handleSaveKnowledge = async () => {
    if (!selectedFile) return
    setIsSavingKnowledge(true)
    const tags = knowledgeTags.split(",").map(t => t.trim()).filter(Boolean)
    const updated = await updateFileKnowledge(selectedFile.id, {
      knowledge_state: knowledgeState || undefined,
      source_kind: sourceKind || undefined,
      summary: knowledgeSummary || undefined,
      tags: tags.length > 0 ? tags : undefined,
    })
    if (updated) setSelectedFile({ ...selectedFile, ...updated })
    setIsEditingKnowledge(false)
    setIsSavingKnowledge(false)
  }

  const handleUpdateRetention = async () => {
    if (selectedFile) {
      await updateFileRetention(selectedFile.id, parseInt(retentionDays))
      setIsUpdatingRetention(false)
    }
  }

  return (
    <div className="flex-1 flex flex-col bg-white dark:bg-[#1a1d21] h-full overflow-hidden">
      <header className="h-14 px-6 flex items-center border-b shrink-0 justify-between">
        <div className="flex items-center gap-2">
          <Folder className="w-5 h-5 text-blue-600" />
          <h1 className="text-lg font-black tracking-tight uppercase">Files &amp; Assets</h1>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant={isContentSearch ? "secondary" : "ghost"}
            size="sm"
            className="text-xs font-bold gap-2"
            onClick={() => { setIsContentSearch(!isContentSearch); setContentSearchQ(""); setContentSearchResults([]) }}
          >
            <FileSearch className={cn("w-3.5 h-3.5", isContentSearch && "text-sky-500")} />
            Content Search
          </Button>
          <Button 
            variant={showStarred ? "secondary" : "ghost"} 
            size="sm" 
            className="text-xs font-bold gap-2"
            onClick={() => { setShowStarred(!showStarred); setShowArchived(false); setIsContentSearch(false) }}
          >
            <Star className={cn("w-3.5 h-3.5", showStarred && "fill-amber-400 text-amber-400")} />
            Starred
          </Button>
          <Button 
            variant={showArchived ? "secondary" : "ghost"} 
            size="sm" 
            className="text-xs font-bold gap-2"
            onClick={() => { setShowArchived(!showArchived); setShowStarred(false); setIsContentSearch(false) }}
          >
            <Archive className="w-3.5 h-3.5" />
            {showArchived ? "Back to Active" : "Archive"}
          </Button>
        </div>
      </header>

      {/* Content Search Panel */}
      {isContentSearch && (
        <div className="p-4 border-b bg-sky-500/5 space-y-3">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-sky-500" />
            <Input
              placeholder="Search file content, extracted text, summaries..."
              className="pl-9 border-sky-300 dark:border-sky-700"
              value={contentSearchQ}
              onChange={e => handleContentSearch(e.target.value)}
            />
            {isSearchingContent && <Loader2 className="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 animate-spin text-sky-500" />}
          </div>
          {contentSearchQ && (
            <div className="space-y-2">
              {contentSearchResults.length === 0 && !isSearchingContent && (
                <p className="text-xs text-muted-foreground italic text-center py-2">No indexed file content matches found.</p>
              )}
              {contentSearchResults.map(r => (
                <div key={r.id} className="bg-white dark:bg-[#1a1d21] rounded-lg border p-3 space-y-1.5 cursor-pointer hover:bg-sky-500/5 transition-colors"
                  onClick={() => handleViewPreview(r)}
                  draggable
                  onDragStart={(e) => {
                    const filePayload = normalizeFilesPageFile(r)
                    if (!filePayload) {
                      e.preventDefault()
                      return
                    }
                    const payload = {
                      kind: "file-to-canvas" as const,
                      file: filePayload,
                      source: "search" as const,
                    };
                    e.dataTransfer.setData("application/json", encodeFileDragPayload(payload));
                    e.dataTransfer.effectAllowed = "copy";
                    // Avoid text selection during drag
                    e.stopPropagation();
                  }}
                >
                  <div className="flex items-center gap-2">
                    <FileIcon className="w-4 h-4 text-sky-600 shrink-0" />
                    <span className="text-sm font-bold truncate">{r.name}</span>
                    {r.extraction_status === 'ready' && (
                      <Badge className="ml-auto text-[9px] h-4 px-1.5 bg-sky-500/10 text-sky-700 dark:text-sky-400 border-sky-300 gap-0.5 shrink-0">
                        <Cpu className="w-2.5 h-2.5" />Indexed
                      </Badge>
                    )}
                  </div>
                  {r.snippet && (
                    <p className="text-[11px] text-muted-foreground leading-relaxed bg-muted/30 rounded px-2 py-1 font-mono">
                      ...{r.snippet}...
                    </p>
                  )}
                  {r.match_reason && (
                    <p className="text-[10px] text-sky-600 font-bold uppercase tracking-wider">{r.match_reason}</p>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Filter Bar */}
      <div className="p-4 border-b bg-muted/10 flex flex-wrap gap-3">
        <div className="relative flex-1 min-w-[200px]">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <Input 
            placeholder={showArchived ? "Search archive..." : "Search active files..."} 
            className="pl-9 bg-white dark:bg-[#1a1d21]"
            value={q}
            onChange={(e) => setQ(e.target.value)}
          />
        </div>

        <Select value={uploaderId} onValueChange={setUploaderId}>
          <SelectTrigger className="w-[160px] bg-white dark:bg-[#1a1d21]">
            <SelectValue placeholder="Uploader" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Uploaders</SelectItem>
            {users.map(u => (
              <SelectItem key={u.id} value={u.id}>{u.name}</SelectItem>
            ))}
          </SelectContent>
        </Select>

        <Select value={contentType} onValueChange={setContentType}>
          <SelectTrigger className="w-[140px] bg-white dark:bg-[#1a1d21]">
            <SelectValue placeholder="File Type" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Types</SelectItem>
            {contentTypes.map(type => (
              <SelectItem key={type} value={type}>{type.toUpperCase()}</SelectItem>
            ))}
          </SelectContent>
        </Select>

        <Button variant="ghost" size="icon" onClick={() => {
          setQ("")
          setUploaderId("all")
          setContentType("all")
        }}>
          <Filter className="w-4 h-4" />
        </Button>
      </div>

      <ScrollArea className="flex-1">
        <div className="p-0">
          <div className="min-w-full divide-y divide-border">
            {activeFiles.map(file => (
              <div 
                key={file.id}
                className="group flex items-center justify-between p-4 hover:bg-muted/30 transition-colors cursor-pointer"
                onClick={() => handleViewPreview(file)}
                draggable
                onDragStart={(e) => {
                  const filePayload = normalizeFilesPageFile(file)
                  if (!filePayload) {
                    e.preventDefault()
                    return
                  }
                  const payload = {
                    kind: "file-to-canvas" as const,
                    file: filePayload,
                    source: "files" as const,
                  };
                  e.dataTransfer.setData("application/json", encodeFileDragPayload(payload));
                  e.dataTransfer.effectAllowed = "copy";
                  // Avoid text selection during drag
                  e.stopPropagation();
                }}
              >
                <div className="flex items-center gap-4 min-w-0">
                  <div className="w-10 h-10 rounded-lg bg-blue-500/10 flex items-center justify-center text-blue-600 shrink-0">
                    <FileIcon className="w-5 h-5" />
                  </div>
                  <div className="min-w-0">
                    <div className="flex items-center gap-1.5 flex-wrap">
                      <h3 className="font-bold text-sm truncate">{file.name}</h3>
                      {file.extraction_status === 'processing' && <Loader2 className="w-3 h-3 text-amber-500 animate-spin shrink-0" />}
                      {file.extraction_status === 'ready' && file.is_searchable && <Cpu className="w-3 h-3 text-sky-500 shrink-0" />}
                      {file.extraction_status === 'failed' && <AlertTriangle className="w-3 h-3 text-red-500 shrink-0" />}
                      {file.extraction_status === 'ocr_needed' && <AlertTriangle className="w-3 h-3 text-orange-500 shrink-0" />}
                      {file.is_citable && <Link2 className="w-3 h-3 text-violet-500 shrink-0" />}
                    </div>
                    <p className="text-[10px] text-muted-foreground uppercase font-black tracking-tighter">
                      {file.type} • {(file.size / 1024).toFixed(1)} KB • {format(new Date(file.createdAt), 'MMM d, yyyy')}
                    </p>
                  </div>
                </div>

                <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                  <Button
                    variant="ghost" size="icon" className="h-8 w-8"
                    onClick={(e) => { e.stopPropagation(); toggleFileStar(file.id) }}
                    title={file.starred ? "Unstar" : "Star"}
                  >
                    <Star className={cn("w-4 h-4", file.starred && "fill-amber-400 text-amber-400")} />
                  </Button>
                  {(file.comment_count || 0) > 0 && (
                    <span className="text-[10px] text-muted-foreground flex items-center gap-0.5">
                      <MessageSquare className="w-3 h-3" />{file.comment_count}
                    </span>
                  )}
                  <Button variant="ghost" size="icon" className="h-8 w-8" asChild onClick={(e) => e.stopPropagation()}>
                    <a href={file.url} target="_blank" rel="noopener noreferrer">
                      <Download className="w-4 h-4" />
                    </a>
                  </Button>
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild onClick={(e) => e.stopPropagation()}>
                      <Button variant="ghost" size="icon" className="h-8 w-8">
                        <MoreVertical className="w-4 h-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      <DropdownMenuItem onClick={(e) => { e.stopPropagation(); handleViewPreview(file); }}>
                        <FileIcon className="w-3.5 h-3.5 mr-2" />
                        <span>Preview &amp; Collaborate</span>
                      </DropdownMenuItem>
                      <DropdownMenuItem onClick={(e) => { e.stopPropagation(); handleViewAudit(file); }}>
                        <History className="w-3.5 h-3.5 mr-2" />
                        <span>Audit History</span>
                      </DropdownMenuItem>
                      <DropdownMenuItem onClick={(e) => { 
                        e.stopPropagation();
                        setSelectedFile(file);
                        setIsUpdatingRetention(true);
                      }}>
                        <ShieldCheck className="w-3.5 h-3.5 mr-2" />
                        <span>Retention Policy</span>
                      </DropdownMenuItem>
                      <DropdownMenuItem onClick={(e) => { e.stopPropagation(); archiveFile(file.id, !showArchived); }}>
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
                      <DropdownMenuItem className="text-red-600" onClick={(e) => { e.stopPropagation(); deleteFile(file.id); }}>
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

      {/* Audit History Dialog */}
      <Dialog open={isViewingAudit} onOpenChange={setIsViewingAudit}>
        <DialogContent className="sm:max-w-[500px]">
          <DialogHeader>
            <DialogTitle className="text-xl font-black flex items-center gap-2">
              <History className="w-5 h-5 text-purple-600" />
              Audit Log: {selectedFile?.name}
            </DialogTitle>
            <DialogDescription className="text-xs">
              View the complete history of actions performed on this file.
            </DialogDescription>
          </DialogHeader>
          <ScrollArea className="max-h-[400px] mt-4 pr-4">
            <div className="space-y-4">
              {auditLogs.map((log) => (
                <div key={log.id} className="flex gap-3 text-xs">
                  <div className="w-1.5 h-1.5 rounded-full bg-blue-500 mt-1.5 shrink-0" />
                  <div className="flex-1">
                    <p className="font-bold">
                      {log.user?.name || "System"} <span className="font-normal text-muted-foreground">performed</span> {log.action.replace('_', ' ')}
                    </p>
                    <p className="text-[10px] text-muted-foreground mt-0.5">{format(new Date(log.occurredAt), 'PPp')}</p>
                    {log.metadata && (
                      <pre className="mt-1 p-2 bg-muted rounded text-[9px] overflow-x-auto">
                        {JSON.stringify(log.metadata, null, 2)}
                      </pre>
                    )}
                  </div>
                </div>
              ))}
              {auditLogs.length === 0 && (
                <p className="text-center py-10 text-muted-foreground italic">No audit history found.</p>
              )}
            </div>
          </ScrollArea>
        </DialogContent>
      </Dialog>

      {/* Retention Policy Dialog */}
      <Dialog open={isUpdatingRetention} onOpenChange={setIsUpdatingRetention}>
        <DialogContent className="sm:max-w-[400px]">
          <DialogHeader>
            <DialogTitle className="text-xl font-black">Retention Policy</DialogTitle>
            <DialogDescription className="text-xs">
              Configure how long this file will be retained in the workspace.
            </DialogDescription>
          </DialogHeader>
          <div className="py-6 space-y-4">
            <p className="text-xs text-muted-foreground leading-relaxed">
              Define how long this file should be kept before it is automatically archived or deleted.
            </p>
            <div className="space-y-2">
              <Label className="text-[10px] font-bold uppercase tracking-wider">Storage Duration</Label>
              <Select value={retentionDays} onValueChange={setRetentionDays}>
                <SelectTrigger>
                  <SelectValue placeholder="Select duration" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="0">Indefinite (Forever)</SelectItem>
                  <SelectItem value="30">30 Days</SelectItem>
                  <SelectItem value="90">90 Days</SelectItem>
                  <SelectItem value="365">1 Year</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setIsUpdatingRetention(false)}>Cancel</Button>
            <Button className="bg-[#3f0e40] text-white" onClick={handleUpdateRetention}>Update Policy</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* File Preview/Collab Dialog */}
      <Dialog open={isViewingPreview} onOpenChange={(open) => { setIsViewingPreview(open); if (!open) setIsEditingKnowledge(false) }}>
        <DialogContent className="sm:max-w-[680px] p-0 overflow-hidden bg-white dark:bg-[#1a1d21] flex flex-col max-h-[90vh]">
          <DialogHeader className="sr-only">
            <DialogTitle>File Details</DialogTitle>
            <DialogDescription>File collaboration and details.</DialogDescription>
          </DialogHeader>

          {/* Preview visual */}
          <div className="h-48 bg-muted flex items-center justify-center relative shrink-0">
            {isPreviewLoading ? (
              <div className="flex flex-col items-center gap-2">
                <Loader2 className="w-8 h-8 animate-spin text-purple-600" />
                <p className="text-xs font-bold uppercase tracking-widest text-muted-foreground">Loading...</p>
              </div>
            ) : filePreview?.preview_url ? (
              <Image src={filePreview.preview_url} alt={selectedFile?.name || 'File Preview'} fill className="object-contain" />
            ) : (
              <div className="w-16 h-16 rounded-2xl bg-blue-500/10 flex items-center justify-center text-blue-600">
                <FileIcon className="w-8 h-8" />
              </div>
            )}
          </div>

          {/* Header row */}
          <div className="px-5 pt-4 pb-2 border-b flex items-start gap-3 shrink-0">
            <div className="flex-1 min-w-0">
              <h2 className="text-base font-black tracking-tight truncate">{selectedFile?.name}</h2>
              <div className="flex items-center gap-2 mt-0.5 flex-wrap">
                <Badge variant="secondary" className="text-[9px] uppercase font-black px-1.5 h-4">
                  {filePreview?.content_type || selectedFile?.type}
                </Badge>
                <span className="text-[10px] text-muted-foreground font-bold">
                  {selectedFile && (selectedFile.size / 1024).toFixed(1)} KB
                </span>
                {selectedFile?.source_kind === 'wiki' && (
                  <Badge className="text-[9px] h-4 px-1.5 bg-emerald-500/10 text-emerald-700 dark:text-emerald-400 border-emerald-500/20 font-bold gap-1">
                    <BookOpen className="w-2.5 h-2.5" /> Wiki
                  </Badge>
                )}
                {selectedFile?.knowledge_state === 'ready' && (
                  <Badge className="text-[9px] h-4 px-1.5 bg-blue-500/10 text-blue-700 dark:text-blue-400 border-blue-500/20 font-bold gap-1">
                    <CheckCircle2 className="w-2.5 h-2.5" /> Ready
                  </Badge>
                )}
              </div>
            </div>
            <div className="flex items-center gap-1 shrink-0">
              <Button
                variant="ghost" size="icon" className="h-8 w-8"
                onClick={() => selectedFile && toggleFileStar(selectedFile.id).then(starred => {
                  if (starred !== null) setSelectedFile((f: any) => ({ ...f, starred }))
                })}
                title={selectedFile?.starred ? "Unstar" : "Star"}
              >
                <Star className={cn("w-4 h-4", selectedFile?.starred && "fill-amber-400 text-amber-400")} />
              </Button>
              <Button variant="ghost" size="icon" className="h-8 w-8" title="Share to channel" onClick={() => setIsShareDialogOpen(true)}>
                <Share2 className="w-4 h-4" />
              </Button>
              <Button asChild size="sm" className="bg-[#3f0e40] text-white hover:bg-[#3f0e40]/90 h-8 text-xs">
                <a href={selectedFile?.url} target="_blank" rel="noopener noreferrer">
                  <Download className="w-3.5 h-3.5 mr-1.5" /> Download
                </a>
              </Button>
            </div>
          </div>

          {/* Tabs */}
          <Tabs value={previewTab} onValueChange={setPreviewTab} className="flex flex-col flex-1 min-h-0">
            <TabsList className="mx-5 mt-3 mb-0 h-8 shrink-0 bg-muted/40 w-fit">
              <TabsTrigger value="details" className="text-xs h-7 px-3">Details</TabsTrigger>
              <TabsTrigger value="comments" className="text-xs h-7 px-3 gap-1">
                <MessageSquare className="w-3 h-3" />
                Comments{comments.length > 0 && <span className="ml-1 text-[10px] font-black">{comments.length}</span>}
              </TabsTrigger>
              <TabsTrigger value="shares" className="text-xs h-7 px-3 gap-1">
                <Share2 className="w-3 h-3" />
                Shares{shares.length > 0 && <span className="ml-1 text-[10px] font-black">{shares.length}</span>}
              </TabsTrigger>
              <TabsTrigger value="knowledge" className="text-xs h-7 px-3 gap-1">
                <BookOpen className="w-3 h-3" />
                Knowledge
              </TabsTrigger>
              <TabsTrigger
                value="indexing"
                className="text-xs h-7 px-3 gap-1"
                onClick={() => selectedFile && handleLoadExtraction(selectedFile.id)}
              >
                <Cpu className="w-3 h-3" />
                Indexing
              </TabsTrigger>
            </TabsList>

            <ScrollArea className="flex-1">
              {/* Details tab */}
              <TabsContent value="details" className="p-5 m-0 space-y-4">
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-1">
                    <p className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">Uploaded By</p>
                    <p className="text-sm font-bold truncate">{filePreview?.uploader?.name || "Unknown User"}</p>
                  </div>
                  <div className="space-y-1">
                    <p className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">Uploaded On</p>
                    <p className="text-sm font-bold">{selectedFile && format(new Date(selectedFile.createdAt), 'PPp')}</p>
                  </div>
                  {filePreview?.expires_at && (
                    <div className="space-y-1">
                      <p className="text-[10px] font-black uppercase tracking-widest text-red-600">Retention Expiry</p>
                      <p className="text-sm font-bold">{format(new Date(filePreview.expires_at), 'PPp')}</p>
                    </div>
                  )}
                  <div className="space-y-1">
                    <p className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">Preview Kind</p>
                    <p className="text-sm font-bold uppercase">{filePreview?.preview_kind || "Standard"}</p>
                  </div>
                  <div className="space-y-1">
                    <p className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">Comments</p>
                    <p className="text-sm font-bold">{comments.length}</p>
                  </div>
                  <div className="space-y-1">
                    <p className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">Shared</p>
                    <p className="text-sm font-bold">{shares.length} time{shares.length !== 1 ? 's' : ''}</p>
                  </div>
                </div>
                {(selectedFile?.tags?.length > 0) && (
                  <div className="space-y-1.5 pt-3 border-t border-dashed">
                    <p className="text-[10px] font-black uppercase tracking-widest text-muted-foreground flex items-center gap-1">
                      <Tag className="w-3 h-3" /> Tags
                    </p>
                    <div className="flex flex-wrap gap-1.5">
                      {(selectedFile.tags as string[]).map((tag: string) => (
                        <Badge key={tag} variant="secondary" className="text-[10px] font-bold px-2 h-5">{tag}</Badge>
                      ))}
                    </div>
                  </div>
                )}
              </TabsContent>

              {/* Comments tab */}
              <TabsContent value="comments" className="p-5 m-0 space-y-4">
                <div className="space-y-3">
                  {comments.length === 0 && (
                    <p className="text-center py-6 text-xs text-muted-foreground italic">No comments yet. Be the first!</p>
                  )}
                  {comments.map(c => (
                    <div key={c.id} className="flex gap-3 text-xs">
                      <div className="w-6 h-6 rounded-full bg-blue-500/10 flex items-center justify-center text-blue-600 font-black text-[10px] shrink-0">
                        {(c.user?.name || 'U').charAt(0).toUpperCase()}
                      </div>
                      <div className="flex-1 bg-muted/40 rounded-xl px-3 py-2">
                        <div className="flex items-center gap-2 mb-1">
                          <span className="font-bold">{c.user?.name || 'User'}</span>
                          <span className="text-muted-foreground text-[10px]">{formatDistanceToNow(new Date(c.created_at), { addSuffix: true })}</span>
                        </div>
                        <p className="text-foreground/90 leading-relaxed">{c.content}</p>
                      </div>
                    </div>
                  ))}
                </div>
                <div className="flex gap-2 pt-2 border-t border-dashed">
                  <Textarea
                    placeholder="Add a comment..."
                    className="min-h-[60px] resize-none text-xs"
                    value={commentInput}
                    onChange={e => setCommentInput(e.target.value)}
                    onKeyDown={e => { if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) handlePostComment() }}
                  />
                  <Button
                    size="icon" className="h-9 w-9 shrink-0 self-end bg-[#3f0e40] text-white hover:bg-[#3f0e40]/90"
                    disabled={!commentInput.trim() || isPostingComment}
                    onClick={handlePostComment}
                  >
                    {isPostingComment ? <Loader2 className="w-4 h-4 animate-spin" /> : <Send className="w-4 h-4" />}
                  </Button>
                </div>
              </TabsContent>

              {/* Shares tab */}
              <TabsContent value="shares" className="p-5 m-0 space-y-3">
                {shares.length === 0 && (
                  <p className="text-center py-6 text-xs text-muted-foreground italic">This file hasn't been shared yet.</p>
                )}
                {shares.map(s => (
                  <div key={s.id} className="bg-muted/30 rounded-xl p-3 text-xs space-y-1.5">
                    <div className="flex items-center gap-2">
                      <div className="w-5 h-5 rounded-full bg-purple-500/10 flex items-center justify-center text-purple-600 font-black text-[9px] shrink-0">
                        {(s.actor?.name || 'U').charAt(0).toUpperCase()}
                      </div>
                      <span className="font-bold">{s.actor?.name || 'User'}</span>
                      <span className="text-muted-foreground">shared to</span>
                      <span className="font-bold">#{channels.find(c => c.id === s.channel_id)?.name || s.channel_id}</span>
                      <span className="ml-auto text-muted-foreground text-[10px]">{formatDistanceToNow(new Date(s.created_at), { addSuffix: true })}</span>
                    </div>
                    {s.comment && <p className="text-muted-foreground pl-7 italic">{s.comment}</p>}
                  </div>
                ))}
                <Button
                  variant="outline" size="sm" className="w-full text-xs font-bold gap-2 mt-2"
                  onClick={() => setIsShareDialogOpen(true)}
                >
                  <Share2 className="w-3.5 h-3.5" /> Share to Channel
                </Button>
              </TabsContent>

              {/* Knowledge tab */}
              <TabsContent value="knowledge" className="p-5 m-0 space-y-4">
                {!isEditingKnowledge ? (
                  <div className="space-y-4">
                    <div className="grid grid-cols-2 gap-4">
                      <div className="space-y-1">
                        <p className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">Source Kind</p>
                        <p className="text-sm font-bold capitalize">{selectedFile?.source_kind || <span className="text-muted-foreground italic text-xs">Not set</span>}</p>
                      </div>
                      <div className="space-y-1">
                        <p className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">Knowledge State</p>
                        <p className="text-sm font-bold capitalize">{selectedFile?.knowledge_state || <span className="text-muted-foreground italic text-xs">Not set</span>}</p>
                      </div>
                    </div>
                    {selectedFile?.summary && (
                      <div className="space-y-1.5">
                        <p className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">Summary</p>
                        <p className="text-xs text-foreground/80 leading-relaxed bg-muted/30 rounded-xl px-3 py-2">{selectedFile.summary}</p>
                      </div>
                    )}
                    {(selectedFile?.tags?.length > 0) && (
                      <div className="space-y-1.5">
                        <p className="text-[10px] font-black uppercase tracking-widest text-muted-foreground flex items-center gap-1"><Tag className="w-3 h-3" /> Tags</p>
                        <div className="flex flex-wrap gap-1.5">
                          {(selectedFile.tags as string[]).map((tag: string) => (
                            <Badge key={tag} variant="secondary" className="text-[10px] font-bold px-2 h-5">{tag}</Badge>
                          ))}
                        </div>
                      </div>
                    )}
                    {(!selectedFile?.source_kind && !selectedFile?.knowledge_state && !selectedFile?.summary) && (
                      <p className="text-center py-4 text-xs text-muted-foreground italic">No knowledge metadata. Add some to make this file retrievable.</p>
                    )}
                    <Button variant="outline" size="sm" className="text-xs font-bold gap-2" onClick={() => setIsEditingKnowledge(true)}>
                      <Edit2 className="w-3.5 h-3.5" /> Edit Metadata
                    </Button>
                  </div>
                ) : (
                  <div className="space-y-4">
                    <div className="grid grid-cols-2 gap-4">
                      <div className="space-y-1.5">
                        <Label className="text-[10px] font-black uppercase tracking-wider">Source Kind</Label>
                        <Select value={sourceKind} onValueChange={setSourceKind}>
                          <SelectTrigger className="h-8 text-xs">
                            <SelectValue placeholder="Select kind" />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="">None</SelectItem>
                            <SelectItem value="wiki"><div className="flex items-center gap-2"><BookOpen className="w-3.5 h-3.5" />Wiki</div></SelectItem>
                            <SelectItem value="doc"><div className="flex items-center gap-2"><FileIcon className="w-3.5 h-3.5" />Document</div></SelectItem>
                            <SelectItem value="report"><div className="flex items-center gap-2"><Globe className="w-3.5 h-3.5" />Report</div></SelectItem>
                            <SelectItem value="reference"><div className="flex items-center gap-2"><Tag className="w-3.5 h-3.5" />Reference</div></SelectItem>
                          </SelectContent>
                        </Select>
                      </div>
                      <div className="space-y-1.5">
                        <Label className="text-[10px] font-black uppercase tracking-wider">Knowledge State</Label>
                        <Select value={knowledgeState} onValueChange={setKnowledgeState}>
                          <SelectTrigger className="h-8 text-xs">
                            <SelectValue placeholder="Select state" />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="">None</SelectItem>
                            <SelectItem value="draft">Draft</SelectItem>
                            <SelectItem value="review">In Review</SelectItem>
                            <SelectItem value="ready">Ready</SelectItem>
                            <SelectItem value="archived">Archived</SelectItem>
                          </SelectContent>
                        </Select>
                      </div>
                    </div>
                    <div className="space-y-1.5">
                      <Label className="text-[10px] font-black uppercase tracking-wider">Summary</Label>
                      <Textarea
                        placeholder="Describe what this file contains..."
                        className="min-h-[70px] resize-none text-xs"
                        value={knowledgeSummary}
                        onChange={e => setKnowledgeSummary(e.target.value)}
                      />
                    </div>
                    <div className="space-y-1.5">
                      <Label className="text-[10px] font-black uppercase tracking-wider">Tags (comma-separated)</Label>
                      <Input
                        placeholder="e.g. onboarding, api, design"
                        className="h-8 text-xs"
                        value={knowledgeTags}
                        onChange={e => setKnowledgeTags(e.target.value)}
                      />
                    </div>
                    <div className="flex gap-2">
                      <Button
                        size="sm" className="bg-[#3f0e40] text-white hover:bg-[#3f0e40]/90 text-xs gap-2"
                        disabled={isSavingKnowledge}
                        onClick={handleSaveKnowledge}
                      >
                        {isSavingKnowledge ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : <CheckCircle2 className="w-3.5 h-3.5" />}
                        Save
                      </Button>
                      <Button variant="ghost" size="sm" className="text-xs" onClick={() => setIsEditingKnowledge(false)}>
                        <X className="w-3.5 h-3.5 mr-1" /> Cancel
                      </Button>
                    </div>
                  </div>
                )}
              </TabsContent>
              {/* Indexing tab */}
              <TabsContent value="indexing" className="p-5 m-0 space-y-4">
                {isLoadingExtraction && (
                  <div className="flex items-center gap-2 py-6 text-xs text-muted-foreground">
                    <Loader2 className="w-4 h-4 animate-spin" /> Loading indexing data...
                  </div>
                )}
                {extraction && (
                  <>
                    {/* Status card */}
                    <div className="rounded-xl border p-3 space-y-3 bg-muted/20">
                      <div className="flex items-center justify-between">
                        <p className="text-[10px] font-black uppercase tracking-widest text-muted-foreground">Extraction Status</p>
                        <Button
                          variant="ghost" size="sm" className="h-6 text-[10px] gap-1 text-muted-foreground"
                          disabled={isRebuildingExtraction}
                          onClick={() => selectedFile && handleRebuildExtraction(selectedFile.id)}
                        >
                          {isRebuildingExtraction ? <Loader2 className="w-3 h-3 animate-spin" /> : <RefreshCw className="w-3 h-3" />}
                          Rebuild
                        </Button>
                      </div>
                      <div className="flex items-center gap-2 flex-wrap">
                        {extraction.status === 'processing' && <Badge className="gap-1 bg-amber-500/10 text-amber-700 border-amber-300 text-[10px]"><Loader2 className="w-3 h-3 animate-spin" />Processing</Badge>}
                        {extraction.status === 'ready' && <Badge className="gap-1 bg-sky-500/10 text-sky-700 border-sky-300 text-[10px]"><Cpu className="w-3 h-3" />Indexed</Badge>}
                        {extraction.status === 'failed' && <Badge className="gap-1 bg-red-500/10 text-red-700 border-red-300 text-[10px]"><AlertTriangle className="w-3 h-3" />Failed</Badge>}
                        {extraction.status === 'ocr_needed' && <Badge className="gap-1 bg-orange-500/10 text-orange-700 border-orange-300 text-[10px]"><AlertTriangle className="w-3 h-3" />OCR Needed</Badge>}
                        {extraction.is_searchable && <Badge className="gap-1 bg-sky-500/10 text-sky-700 border-sky-300 text-[10px]"><Search className="w-3 h-3" />Searchable</Badge>}
                        {extraction.is_citable && <Badge className="gap-1 bg-violet-500/10 text-violet-700 border-violet-300 text-[10px]"><Link2 className="w-3 h-3" />Citable</Badge>}
                        {extraction.needs_ocr && <Badge className="gap-1 bg-orange-500/10 text-orange-700 border-orange-300 text-[10px]"><AlertTriangle className="w-3 h-3" />{extraction.ocr_is_mock ? 'Mock OCR' : extraction.ocr_provider || 'OCR'}</Badge>}
                      </div>
                      {extraction.last_indexed_at && (
                        <p className="text-[10px] text-muted-foreground">Last indexed: {format(new Date(extraction.last_indexed_at), 'PPp')}</p>
                      )}
                      {extraction.content_summary && (
                        <p className="text-[11px] text-foreground/80 leading-relaxed bg-muted/40 rounded-lg px-3 py-2">{extraction.content_summary}</p>
                      )}
                    </div>

                    {/* Extracted text */}
                    <div className="space-y-2">
                      <Button variant="outline" size="sm" className="text-[10px] font-bold gap-1.5 h-7"
                        onClick={() => selectedFile && (showExtractedText ? setShowExtractedText(false) : handleLoadExtractedText(selectedFile.id))}
                      >
                        <AlignLeft className="w-3 h-3" />
                        {showExtractedText ? 'Hide' : 'Show'} Extracted Text
                      </Button>
                      {showExtractedText && (
                        <pre className="text-[10px] text-foreground/80 leading-relaxed bg-muted/30 rounded-lg p-3 max-h-40 overflow-y-auto whitespace-pre-wrap break-words font-mono">
                          {extractedText || '(No extracted text available)'}
                        </pre>
                      )}
                    </div>

                    {/* Chunks */}
                    <div className="space-y-2">
                      <Button variant="outline" size="sm" className="text-[10px] font-bold gap-1.5 h-7"
                        onClick={() => selectedFile && (showChunks ? setShowChunks(false) : handleLoadChunks(selectedFile.id))}
                      >
                        <Layers className="w-3 h-3" />
                        {showChunks ? 'Hide' : 'Show'} Chunks {chunks.length > 0 && `(${chunks.length})`}
                      </Button>
                      {showChunks && (
                        <div className="space-y-2 max-h-40 overflow-y-auto">
                          {chunks.length === 0 && <p className="text-xs text-muted-foreground italic">No chunks available.</p>}
                          {chunks.map((chunk, i) => (
                            <div key={chunk.id || i} className="bg-muted/30 rounded-lg p-2 text-[10px] font-mono">
                              <p className="text-muted-foreground mb-1 font-sans font-bold">Chunk {chunk.chunk_index + 1}{chunk.char_count ? ` · ${chunk.char_count} chars` : ''}{chunk.token_estimate ? ` · ~${chunk.token_estimate} tokens` : ''}</p>
                              <p className="text-foreground/80 line-clamp-3 leading-relaxed break-words">{chunk.content}</p>
                            </div>
                          ))}
                        </div>
                      )}
                    </div>

                    {/* Citations */}
                    <div className="space-y-2">
                      <Button variant="outline" size="sm" className="text-[10px] font-bold gap-1.5 h-7"
                        onClick={() => selectedFile && (showCitations ? setShowCitations(false) : handleLoadCitations(selectedFile.id))}
                      >
                        <Link2 className="w-3 h-3" />
                        {showCitations ? 'Hide' : 'Show'} Citations {citations.length > 0 && `(${citations.length})`}
                      </Button>
                      {showCitations && (
                        <div className="space-y-2 max-h-52 overflow-y-auto">
                          {citations.length === 0 && <p className="text-xs text-muted-foreground italic">No citations yet.</p>}
                          {citations.map((cit, i) => (
                            <CitationCard key={cit.id || i} citation={cit} compact />
                          ))}
                        </div>
                      )}
                    </div>
                  </>
                )}
                {!isLoadingExtraction && !extraction && (
                  <div className="py-8 text-center space-y-2">
                    <Cpu className="w-8 h-8 text-muted-foreground/20 mx-auto" />
                    <p className="text-xs text-muted-foreground italic">No extraction data found. Try rebuilding.</p>
                    <Button variant="outline" size="sm" className="text-[10px] font-bold gap-1.5 h-7"
                      disabled={isRebuildingExtraction}
                      onClick={() => selectedFile && handleRebuildExtraction(selectedFile.id)}
                    >
                      {isRebuildingExtraction ? <Loader2 className="w-3 h-3 animate-spin" /> : <RefreshCw className="w-3 h-3" />}
                      Rebuild Extraction
                    </Button>
                  </div>
                )}
              </TabsContent>
            </ScrollArea>
          </Tabs>
        </DialogContent>
      </Dialog>

      {/* Share to Channel Dialog */}
      <Dialog open={isShareDialogOpen} onOpenChange={setIsShareDialogOpen}>
        <DialogContent className="sm:max-w-[400px]">
          <DialogHeader>
            <DialogTitle className="text-base font-black flex items-center gap-2">
              <Share2 className="w-4 h-4 text-purple-600" /> Share File to Channel
            </DialogTitle>
            <DialogDescription className="text-xs">Share <strong>{selectedFile?.name}</strong> into a channel with an optional message.</DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-2">
            <div className="space-y-1.5">
              <Label className="text-[10px] font-black uppercase tracking-wider">Channel</Label>
              <Select value={shareChannelId} onValueChange={setShareChannelId}>
                <SelectTrigger className="h-9 text-xs">
                  <SelectValue placeholder="Select channel..." />
                </SelectTrigger>
                <SelectContent>
                  {channels.filter(c => !c.isArchived).map(c => (
                    <SelectItem key={c.id} value={c.id}>#{c.name}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-1.5">
              <Label className="text-[10px] font-black uppercase tracking-wider">Message (optional)</Label>
              <Textarea
                placeholder="Add a note about this file..."
                className="min-h-[70px] resize-none text-xs"
                value={shareComment}
                onChange={e => setShareComment(e.target.value)}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="ghost" size="sm" onClick={() => setIsShareDialogOpen(false)}>Cancel</Button>
            <Button
              size="sm" className="bg-[#3f0e40] text-white hover:bg-[#3f0e40]/90 text-xs gap-2"
              disabled={!shareChannelId || isSharingFile}
              onClick={handleShareFile}
            >
              {isSharingFile ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : <Share2 className="w-3.5 h-3.5" />}
              Share File
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
