import { Node, mergeAttributes } from "@tiptap/core"
import { ReactNodeViewRenderer } from "@tiptap/react"
import { FileToCanvasFileCard } from "@/components/layout/canvas-file-card"

export const FileRef = Node.create({
  name: "file_ref",
  group: "block",
  atom: true,

  addAttributes() {
    return {
      file_id: { default: null },
      title: { default: "Untitled File" },
      mime_type: { default: "application/octet-stream" },
      size: { default: 0 },
      preview_url: { default: null },
    }
  },

  parseHTML() {
    return [
      {
        tag: 'div[data-type="file_ref"]',
        getAttributes: element => ({
          file_id: element.getAttribute('data-file-id'),
          title: element.getAttribute('data-title'),
          mime_type: element.getAttribute('data-mime-type'),
          size: parseInt(element.getAttribute('data-size') || '0', 10),
          preview_url: element.getAttribute('data-preview-url'),
        }),
      },
    ]
  },

  renderHTML({ HTMLAttributes }) {
    return [
      "div",
      mergeAttributes(HTMLAttributes, {
        "data-type": "file_ref",
        "data-file-id": HTMLAttributes.file_id,
        "data-title": HTMLAttributes.title,
        "data-mime-type": HTMLAttributes.mime_type,
        "data-size": HTMLAttributes.size,
        "data-preview-url": HTMLAttributes.preview_url,
      }),
    ]
  },

  addNodeView() {
    return ReactNodeViewRenderer(FileToCanvasFileCard)
  },
})
