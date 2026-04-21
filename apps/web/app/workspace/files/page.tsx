"use client"

import { useEffect, useState } from "react"
import { useFileStore } from "@/stores/file-store"
import { Search, Folder, FileIcon, Archive, Trash2, Download, Filter, MoreVertical, RefreshCcw, History, ShieldCheck, Loader2, Star, MessageSquare, Share2, BookOpen, Tag, Send, Globe, CheckCircle2, Edit2, X } from "lucide-react"
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
import type { FileComment, FileShare } from "@/types"

export default function FilesPage() {
  const { 
    files, archivedFiles, starredFiles,
    fetchFiles, fetchArchivedFiles, archiveFile, deleteFile,
    fetchFileAuditHistory, updateFileRetention, fetchFilePreview,
    fetchFileComments, createFileComment, toggleFileStar, fetchStarredFiles,
    shareFile, fetchFileShares, updateFileKnowledge,
  } = useFileStore()
  const { channels } = useChannelStore()
  const { users, fetchUsers } = useUserStore()
  
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

  const handleViewPreview = async (file: any) => {
    setSelectedFile(file)
    setIsPreviewLoading(true)
    setIsViewingPreview(true)
    setPreviewTab("details")
    setComments([])
    setShares([])
    setIsEditingKnowledge(false)
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
          <h1 className="text-lg font-black tracking-tight uppercase">Files & Assets</h1>
        </div>
        <div className="flex items-center gap-2">
          <Button 
            variant={showStarred ? "secondary" : "ghost"} 
            size="sm" 
            className="text-xs font-bold gap-2"
            onClick={() => { setShowStarred(!showStarred); setShowArchived(false) }}
          >
            <Star className={cn("w-3.5 h-3.5", showStarred && "fill-amber-400 text-amber-400")} />
            Starred
          </Button>
          <Button 
            variant={showArchived ? "secondary" : "ghost"} 
            size="sm" 
            className="text-xs font-bold gap-2"
            onClick={() => { setShowArchived(!showArchived); setShowStarred(false) }}
          >
            <Archive className="w-3.5 h-3.5" />
            {showArchived ? "Back to Active" : "Archive"}
          </Button>
        </div>
      </header>

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
