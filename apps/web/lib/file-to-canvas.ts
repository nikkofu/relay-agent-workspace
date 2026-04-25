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
export function normalizeFilesPageFile(file: any): NormalizedFilePayload {
  return {
    id: file.id,
    title: file.name || file.title,
    mime_type: file.content_type || file.mime_type,
    size: file.size_bytes || file.size,
    preview_url: file.preview_url || `/api/v1/files/${file.id}/preview`,
  }
}

/**
 * Normalizes a file object from a message attachment metadata.
 */
export function normalizeMessageAttachment(attachment: any): NormalizedFilePayload {
  // Message metadata often nests the file object or keeps it flat.
  const f = attachment.file || attachment
  return {
    id: f.id,
    title: f.title || f.name,
    mime_type: f.mime_type || f.contentType,
    size: f.size || f.sizeBytes,
    preview_url: f.preview_url || `/api/v1/files/${f.id}/preview`,
  }
}

/**
 * Normalizes a search result hit.
 */
export function normalizeSearchFile(result: any): NormalizedFilePayload {
  const f = result.file || result
  return {
    id: f.id,
    title: f.title || f.name,
    mime_type: f.mime_type || f.content_type,
    size: f.size || f.size_bytes,
    preview_url: f.preview_url || `/api/v1/files/${f.id}/preview`,
  }
}

/**
 * Type guard for the shared drag payload.
 */
export function isFileToCanvasDragPayload(data: any): data is FileToCanvasDragPayload {
  return (
    data &&
    data.kind === "file-to-canvas" &&
    data.file &&
    typeof data.file.id === "string" &&
    typeof data.file.title === "string"
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
