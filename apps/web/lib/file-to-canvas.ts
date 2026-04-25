/**
 * Phase 68: File-to-Canvas Drag & Drop Contract
 * 
 * Defines the shared types and normalization helpers for dragging files
 * from various workspace surfaces (Files page, message attachments, search)
 * and dropping them into a canvas.
 */

export interface NormalizedFilePayload {
  id: string
  title: string
  mime_type: string
  size: number
  preview_url?: string
}

export interface FileToCanvasDragPayload {
  kind: "file-to-canvas"
  file: NormalizedFilePayload
  source: "files" | "message" | "search"
}

/**
 * The persisted block shape inside a canvas (TipTap node / Artifact content).
 */
export interface FileRefBlock {
  type: "file_ref"
  file_id: string
  title: string
  mime_type: string
  size: number
  preview_url?: string
}

/**
 * Normalizes a raw file object from the Files page or file-store.
 */
function normalizeFileLike(file: any): NormalizedFilePayload | null {
  if (!file || typeof file !== "object") return null
  const id = file.id
  const title = file.title || file.name || file.filename
  const mime_type = file.mime_type || file.content_type || file.type || file.mimeType || "application/octet-stream"
  const size = file.size ?? file.size_bytes ?? file.sizeBytes ?? 0
  if (typeof id !== "string" || !id.trim()) return null
  if (typeof title !== "string" || !title.trim()) return null
  if (typeof mime_type !== "string" || !mime_type.trim()) return null
  if (typeof size !== "number" || Number.isNaN(size) || size < 0) return null
  return {
    id,
    title,
    mime_type,
    size,
    preview_url: typeof file.preview_url === "string" && file.preview_url
      ? file.preview_url
      : `/api/v1/files/${id}/preview`,
  }
}

export function normalizeFilesPageFile(file: any): NormalizedFilePayload | null {
  return normalizeFileLike(file)
}

/**
 * Normalizes a file object from a message attachment metadata.
 */
export function normalizeMessageAttachment(attachment: any): NormalizedFilePayload | null {
  const f = attachment.file || attachment
  return normalizeFileLike(f)
}

/**
 * Normalizes a search result hit.
 */
export function normalizeSearchFile(result: any): NormalizedFilePayload | null {
  const f = result.file || result
  return normalizeFileLike(f)
}

/**
 * Type guard for the shared drag payload.
 */
export function isFileToCanvasDragPayload(data: any): data is FileToCanvasDragPayload {
  return (
    data &&
    data.kind === "file-to-canvas" &&
    (data.source === "files" || data.source === "message" || data.source === "search") &&
    data.file &&
    typeof data.file.id === "string" &&
    data.file.id.trim().length > 0 &&
    typeof data.file.title === "string" &&
    data.file.title.trim().length > 0 &&
    typeof data.file.mime_type === "string" &&
    data.file.mime_type.trim().length > 0 &&
    typeof data.file.size === "number" &&
    !Number.isNaN(data.file.size) &&
    data.file.size >= 0 &&
    (data.file.preview_url === undefined || typeof data.file.preview_url === "string")
  )
}

/**
 * Encodes the payload for DataTransfer.
 */
export function encodeFileDragPayload(payload: FileToCanvasDragPayload): string {
  return JSON.stringify(payload)
}

/**
 * Decodes the payload from DataTransfer. Returns null if invalid.
 */
export function decodeFileDragPayload(json: string): FileToCanvasDragPayload | null {
  try {
    const data = JSON.parse(json)
    if (isFileToCanvasDragPayload(data)) return data
    return null
  } catch {
    return null
  }
}
