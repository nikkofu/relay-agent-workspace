"use client"

import { useState } from "react"
import {
  FileText, FileCode, FileImage, FileAudio, FileVideo, File,
  Star, MessageSquare, Share2, Download, ExternalLink,
  Tag, BookOpen, CheckCircle2, ChevronDown, ChevronUp, Loader2,
  Cpu, AlertTriangle, Search as SearchIcon, Link2,
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { cn } from "@/lib/utils"
import { MessageAttachment, FileAsset } from "@/types"
import { API_BASE_URL } from "@/lib/constants"
import Image from "next/image"
import { normalizeMessageAttachment, encodeFileDragPayload } from "@/lib/file-to-canvas"

interface FileAttachmentCardProps {
  attachment: MessageAttachment
  messageId?: string
}

function getFileIcon(mimeType?: string) {
  if (!mimeType) return File
  if (mimeType.startsWith("image/")) return FileImage
  if (mimeType.startsWith("audio/")) return FileAudio
  if (mimeType.startsWith("video/")) return FileVideo
  if (
    mimeType.includes("javascript") ||
    mimeType.includes("typescript") ||
    mimeType.includes("json") ||
    mimeType.includes("xml") ||
    mimeType.includes("code")
  ) return FileCode
  if (mimeType.includes("text/")) return FileText
  return File
}

function formatSize(bytes?: number): string | null {
  if (!bytes) return null
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

export function FileAttachmentCard({ attachment, messageId }: FileAttachmentCardProps) {
  const [expanded, setExpanded] = useState(false)
  const [lazyFiles, setLazyFiles] = useState<FileAsset[] | null>(null)
  const [isLoadingLazy, setIsLoadingLazy] = useState(false)

  const file = attachment.file || {}
  const preview = attachment.preview || {}

  const name = file.name || attachment.name
  const size = file.size || attachment.size
  const mimeType = file.type || attachment.mimeType
  const url = file.url || attachment.url

  const isWiki = file.source_kind === "wiki"
  const isReady = file.knowledge_state === "ready"
  const extractionStatus = file.extraction_status
  const isSearchable = file.is_searchable
  const isCitable = file.is_citable
  const isStarred = !!file.starred
  const commentCount = file.comment_count ?? 0
  const shareCount = file.share_count ?? 0
  const tags = file.tags || []
  const summary = file.summary

  const hasThumbnail = !!preview.thumbnail_url
  const FileIcon = getFileIcon(mimeType)

  const handleExpand = async () => {
    const next = !expanded
    setExpanded(next)
    if (next && !lazyFiles && messageId) {
      try {
        setIsLoadingLazy(true)
        const res = await fetch(`${API_BASE_URL}/messages/${messageId}/files`)
        if (res.ok) {
          const data = await res.json()
          setLazyFiles(data.files || [])
        }
      } catch {
        // silent fallback
      } finally {
        setIsLoadingLazy(false)
      }
    }
  }

  return (
    <div className={cn(
      "rounded-xl border bg-card shadow-sm overflow-hidden max-w-[360px] transition-all cursor-grab active:cursor-grabbing",
      isWiki && "border-violet-500/40 bg-violet-500/5 dark:bg-violet-950/20",
      isReady && !isWiki && "border-emerald-500/40 bg-emerald-500/5 dark:bg-emerald-950/20"
    )}
      draggable
      onDragStart={(e) => {
        const payload = {
          kind: "file-to-canvas" as const,
          file: normalizeMessageAttachment(attachment),
          source: "message" as const,
        };
        e.dataTransfer.setData("application/json", encodeFileDragPayload(payload));
        e.dataTransfer.effectAllowed = "copy";
        e.stopPropagation();
      }}
    >
      {/* Thumbnail */}
      {hasThumbnail && (
        <div className="relative h-32 w-full bg-muted overflow-hidden border-b">
          <Image
            src={preview.thumbnail_url!}
            alt={name || "attachment"}
            fill
            className="object-cover"
          />
        </div>
      )}

      {/* Main content row */}
      <div className="flex items-start gap-3 p-3">
        {/* Icon */}
        <div className={cn(
          "w-10 h-10 rounded-lg flex items-center justify-center shrink-0",
          hasThumbnail ? "w-8 h-8" : "",
          isWiki
            ? "bg-violet-500/10 text-violet-600"
            : isReady
            ? "bg-emerald-500/10 text-emerald-600"
            : "bg-blue-500/10 text-blue-600"
        )}>
          <FileIcon className={cn("w-5 h-5", hasThumbnail && "w-4 h-4")} />
        </div>

        {/* Info */}
        <div className="flex-1 min-w-0">
          <p className="text-sm font-semibold truncate leading-tight">{name}</p>

          <div className="flex items-center gap-1.5 mt-0.5 flex-wrap">
            {mimeType && (
              <span className="text-[10px] text-muted-foreground uppercase font-mono tracking-tight">
                {mimeType.split("/")[1] || mimeType}
              </span>
            )}
            {size && (
              <span className="text-[10px] text-muted-foreground">
                · {formatSize(size)}
              </span>
            )}
          </div>

          {/* Summary */}
          {summary && (
            <p className="text-[11px] text-muted-foreground mt-1 line-clamp-2 leading-snug">
              {summary}
            </p>
          )}

          {/* Badges + counters */}
          <div className="flex items-center gap-1.5 mt-1.5 flex-wrap">
            {isWiki && (
              <Badge
                variant="secondary"
                className="h-4 px-1.5 text-[9px] bg-violet-500/10 text-violet-700 dark:text-violet-400 border-violet-300 dark:border-violet-700 gap-0.5"
              >
                <BookOpen className="w-2.5 h-2.5" />
                Wiki
              </Badge>
            )}
            {isReady && (
              <Badge
                variant="secondary"
                className="h-4 px-1.5 text-[9px] bg-emerald-500/10 text-emerald-700 dark:text-emerald-400 border-emerald-300 dark:border-emerald-700 gap-0.5"
              >
                <CheckCircle2 className="w-2.5 h-2.5" />
                Ready
              </Badge>
            )}
            {isStarred && (
              <Star className="w-3 h-3 text-amber-500 fill-amber-500" />
            )}
            {extractionStatus === 'processing' && (
              <Badge variant="secondary" className="h-4 px-1.5 text-[9px] bg-amber-500/10 text-amber-700 dark:text-amber-400 border-amber-300 dark:border-amber-700 gap-0.5">
                <Loader2 className="w-2.5 h-2.5 animate-spin" />Indexing
              </Badge>
            )}
            {extractionStatus === 'ready' && (
              <Badge variant="secondary" className="h-4 px-1.5 text-[9px] bg-sky-500/10 text-sky-700 dark:text-sky-400 border-sky-300 dark:border-sky-700 gap-0.5">
                <Cpu className="w-2.5 h-2.5" />Indexed
              </Badge>
            )}
            {extractionStatus === 'failed' && (
              <Badge variant="secondary" className="h-4 px-1.5 text-[9px] bg-red-500/10 text-red-700 dark:text-red-400 border-red-300 dark:border-red-700 gap-0.5">
                <AlertTriangle className="w-2.5 h-2.5" />Failed
              </Badge>
            )}
            {extractionStatus === 'ocr_needed' && (
              <Badge variant="secondary" className="h-4 px-1.5 text-[9px] bg-orange-500/10 text-orange-700 dark:text-orange-400 border-orange-300 dark:border-orange-700 gap-0.5">
                <AlertTriangle className="w-2.5 h-2.5" />OCR Needed
              </Badge>
            )}
            {isSearchable && (
              <SearchIcon className="w-3 h-3 text-sky-500" />
            )}
            {isCitable && (
              <Link2 className="w-3 h-3 text-violet-500" />
            )}
            {commentCount > 0 && (
              <span className="flex items-center gap-0.5 text-[10px] text-muted-foreground">
                <MessageSquare className="w-3 h-3" />
                {commentCount}
              </span>
            )}
            {shareCount > 0 && (
              <span className="flex items-center gap-0.5 text-[10px] text-muted-foreground">
                <Share2 className="w-3 h-3" />
                {shareCount}
              </span>
            )}
          </div>
        </div>

        {/* Actions */}
        <div className="flex flex-col items-end gap-1 shrink-0">
          {url && (
            <Button variant="ghost" size="icon" className="h-6 w-6" asChild>
              <a
                href={url}
                target="_blank"
                rel="noopener noreferrer"
                download={name}
                title="Download"
              >
                <Download className="w-3 h-3" />
              </a>
            </Button>
          )}
          {preview.url && (
            <Button variant="ghost" size="icon" className="h-6 w-6" asChild>
              <a href={preview.url} target="_blank" rel="noopener noreferrer" title="Open preview">
                <ExternalLink className="w-3 h-3" />
              </a>
            </Button>
          )}
        </div>
      </div>

      {/* Tags row */}
      {tags.length > 0 && (
        <div className="flex flex-wrap gap-1 px-3 pb-2.5">
          {tags.map((tag) => (
            <span
              key={tag}
              className="flex items-center gap-0.5 text-[10px] text-muted-foreground bg-muted/60 px-1.5 py-0.5 rounded-full"
            >
              <Tag className="w-2.5 h-2.5" />
              {tag}
            </span>
          ))}
        </div>
      )}

      {/* Expand toggle — lazy-loads GET /api/v1/messages/:id/files */}
      {messageId && (
        <button
          className="w-full px-3 py-1.5 flex items-center justify-center gap-1 text-[10px] text-muted-foreground hover:text-foreground hover:bg-muted/40 transition-colors border-t"
          onClick={handleExpand}
        >
          {isLoadingLazy ? (
            <Loader2 className="w-3 h-3 animate-spin" />
          ) : expanded ? (
            <>
              <ChevronUp className="w-3 h-3" />
              Hide details
            </>
          ) : (
            <>
              <ChevronDown className="w-3 h-3" />
              More details
            </>
          )}
        </button>
      )}

      {/* Lazy-loaded file inspector */}
      {expanded && (
        <div className="px-3 pb-3 pt-2 border-t space-y-1.5">
          {lazyFiles === null && !isLoadingLazy && (
            <p className="text-[11px] text-muted-foreground">No additional file data.</p>
          )}
          {lazyFiles && lazyFiles.length === 0 && (
            <p className="text-[11px] text-muted-foreground">No files attached to this message.</p>
          )}
          {lazyFiles && lazyFiles.map((f) => (
            <div key={f.id} className="flex items-center gap-2 text-[11px]">
              <FileIcon className="w-3 h-3 shrink-0 text-muted-foreground" />
              <span className="font-medium truncate">{f.name}</span>
              {f.summary && (
                <span className="text-[10px] text-muted-foreground truncate hidden sm:inline">
                  — {f.summary}
                </span>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
