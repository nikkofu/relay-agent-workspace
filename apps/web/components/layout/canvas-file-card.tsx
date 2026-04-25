"use client"

import { NodeViewWrapper, NodeViewProps } from "@tiptap/react"
import { FileIcon, FileText, FileImage, FileAudio, FileVideo, Download, ExternalLink } from "lucide-react"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"

function getFileIcon(mimeType?: string) {
  if (!mimeType) return FileIcon
  if (mimeType.startsWith("image/")) return FileImage
  if (mimeType.startsWith("audio/")) return FileAudio
  if (mimeType.startsWith("video/")) return FileVideo
  if (mimeType.includes("pdf") || mimeType.includes("text/")) return FileText
  return FileIcon
}

function formatSize(bytes?: number): string {
  if (!bytes) return "0 B"
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

export function FileToCanvasFileCard(props: NodeViewProps) {
  const { node, selected } = props
  const { file_id, title, mime_type, size, preview_url } = node.attrs

  const Icon = getFileIcon(mime_type)

  return (
    <NodeViewWrapper className={cn(
      "my-4 p-3 bg-muted/40 border-2 rounded-xl flex items-center gap-3 transition-all group",
      selected ? "border-sky-500 bg-sky-500/5 ring-1 ring-sky-500" : "border-transparent hover:border-muted-foreground/20 hover:bg-muted/60"
    )}>
      <div className="w-10 h-10 rounded-lg bg-background border flex items-center justify-center shrink-0 shadow-sm">
        <Icon className="w-5 h-5 text-muted-foreground" />
      </div>

      <div className="flex-1 min-w-0">
        <p className="text-sm font-bold truncate leading-tight">{title}</p>
        <p className="text-[10px] text-muted-foreground uppercase font-black tracking-tighter mt-0.5">
          {mime_type.split("/")[1] || mime_type} • {formatSize(size)}
        </p>
      </div>

      <div className="flex items-center gap-1 shrink-0 opacity-0 group-hover:opacity-100 transition-opacity">
        <Button variant="ghost" size="icon" className="h-8 w-8" asChild>
          <a href={`/api/v1/files/${file_id}/download`} download={title}>
            <Download className="w-4 h-4" />
          </a>
        </Button>
        <Button variant="ghost" size="icon" className="h-8 w-8" asChild>
          <a href={preview_url || `/workspace/files?id=${file_id}`} target="_blank" rel="noopener noreferrer">
            <ExternalLink className="w-4 h-4" />
          </a>
        </Button>
      </div>
    </NodeViewWrapper>
  )
}
