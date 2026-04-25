/**
 * Phase 69: Canvas File-Group Snapshot Helper
 * 
 * Extracts persisted file_ref blocks from a canvas document to create a 
 * trigger-time snapshot for AI analysis.
 */

export interface FileRefSnapshot {
  file_id: string
}

/**
 * Extracts all file_id references from TipTap HTML content.
 * 
 * For Phase 69, we look for <div data-type="file_ref" data-file-id="...">
 */
export function extractFileRefsFromHtml(html: string): FileRefSnapshot[] {
  if (!html) return []
  
  const fileRefs: FileRefSnapshot[] = []
  const parser = new DOMParser()
  const doc = parser.parseFromString(html, "text/html")
  
  const nodes = doc.querySelectorAll('div[data-type="file_ref"]')
  nodes.forEach((node) => {
    const fileId = node.getAttribute("data-file-id")
    if (fileId) {
      fileRefs.push({ file_id: fileId })
    }
  })
  
  return fileRefs
}

/**
 * Filters out duplicate file IDs to ensure the AI analyzes unique documents.
 */
export function getUniqueFileRefs(refs: FileRefSnapshot[]): FileRefSnapshot[] {
  const seen = new Set<string>()
  return refs.filter((ref) => {
    if (seen.has(ref.file_id)) return false
    seen.add(ref.file_id)
    return true
  })
}
