"use client"

// ── Canvas TipTap editor ──────────────────────────────────────────────────────
//
// Rich-text editor for document-type canvases. Emits HTML via
// `editor.getHTML()` so it stays wire-compatible with the existing
// `dangerouslySetInnerHTML` view-mode renderer and with the server's
// Artifact.Content field.
//
// Exposes an imperative handle (see CanvasEditorHandle) so sibling components
// — most notably the Canvas AI Dock — can read the current selection and
// apply AI-produced HTML without lifting the editor state all the way up.
//
// The toolbar's purple "AI" button no longer opens a modal dialog; it calls
// the optional `onRequestAIEdit` prop (the dock focuses its composer so the
// user can keep their cursor / selection context intact).

import { forwardRef, useEffect, useImperativeHandle } from "react"
import { useEditor, EditorContent, Editor } from "@tiptap/react"
import StarterKit from "@tiptap/starter-kit"
import Placeholder from "@tiptap/extension-placeholder"
import TiptapLink from "@tiptap/extension-link"
import { FileRef } from "@/lib/tiptap-file-ref"
import { decodeFileDragPayload } from "@/lib/file-to-canvas"
import {
  Bold, Italic, Strikethrough, Heading1, Heading2, Heading3,
  List, ListOrdered, Quote, Code, Link as LinkIcon, Undo, Redo,
  Wand2,
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"

/** Range inside the ProseMirror document: [from, to) positions. */
export interface EditorRange { from: number; to: number }

// Imperative surface exposed to sibling components (canvas-ai-dock.tsx) via
// a ref. Keeps the AI composer out of the editor's internal state while
// still letting it read selections and apply AI output deterministically.
export interface CanvasEditorHandle {
  /** Plain text of the current selection, or "" when nothing is selected. */
  getSelectionText: () => string
  /** Plain text of the entire document. */
  getDocText: () => string
  /** True when the selection is non-empty. */
  hasSelection: () => boolean
  /** Current selection range, even when empty (from === to). */
  getSelectionRange: () => EditorRange | null
  /** Replace the current selection with HTML; if no selection, replace the whole
   *  doc. Returns the range of the newly-inserted content, or null on failure. */
  applyHtmlToSelection: (html: string) => EditorRange | null
  /** Replace the whole document with HTML. Returns the new full range. */
  applyHtmlToDoc: (html: string) => EditorRange | null
  /** Insert HTML at the current cursor position (no deletion). Returns the
   *  range of the newly-inserted content. */
  insertHtmlAtCursor: (html: string) => EditorRange | null
  /** Select a range (highlights it visually) and scroll it into view. */
  highlightRange: (range: EditorRange) => void
  /** Collapse the current selection (removes highlight). */
  clearSelection: () => void
  /** Move focus back into the editor. */
  focus: () => void
}

interface CanvasTipTapEditorProps {
  content: string
  onChange: (html: string) => void
  channelId: string
  autoFocus?: boolean
  readOnly?: boolean
  placeholder?: string
  /** Invoked when the toolbar's AI button is pressed. Parent typically focuses
   *  the Canvas AI dock composer here. */
  onRequestAIEdit?: () => void
}

export const CanvasTipTapEditor = forwardRef<CanvasEditorHandle, CanvasTipTapEditorProps>(
  function CanvasTipTapEditor(
    { content, onChange, autoFocus = true, readOnly = false,
      placeholder = "Start writing, or press ⌘K / the AI button to draft with AI…",
      onRequestAIEdit },
    ref,
  ) {
  const editor = useEditor({
    extensions: [
      StarterKit.configure({
        heading: { levels: [1, 2, 3] },
        // Disable the built-in Link mark — we add the extension below.
        link: false,
      }),
      Placeholder.configure({ placeholder }),
      TiptapLink.configure({
        openOnClick: false,
        HTMLAttributes: { class: "text-blue-600 underline cursor-pointer" },
      }),
      FileRef,
    ],
    content: content || "",
    editable: !readOnly,
    autofocus: autoFocus ? "end" : false,
    immediatelyRender: false,
    editorProps: {
      attributes: {
        class:
          "prose dark:prose-invert max-w-none focus:outline-none min-h-[500px] px-1",
      },
      handleDrop: (view, event, _slice, moved) => {
        if (!moved && event.dataTransfer) {
          const json = event.dataTransfer.getData("application/json")
          const payload = decodeFileDragPayload(json)
          if (payload) {
            const { file } = payload
            const coordinates = view.posAtCoords({ left: event.clientX, top: event.clientY })
            const pos = coordinates ? coordinates.pos : view.state.selection.head
            
            view.dispatch(
              view.state.tr.insert(
                pos,
                view.state.schema.nodes.file_ref.create({
                  file_id: file.id,
                  title: file.title,
                  mime_type: file.mime_type,
                  size: file.size,
                  preview_url: file.preview_url,
                })
              )
            )
            return true
          }
        }
        return false
      },
    },
    onUpdate: ({ editor }) => {
      onChange(editor.getHTML())
    },
  })

  // Expose imperative commands to parent via ref so the AI dock can talk to
  // the editor without us lifting editor state. All methods no-op safely
  // while the editor is still initializing.
  useImperativeHandle(ref, () => ({
    getSelectionText: () => {
      if (!editor) return ""
      const { from, to } = editor.state.selection
      if (from === to) return ""
      return editor.state.doc.textBetween(from, to, "\n\n")
    },
    getDocText: () => editor?.state.doc.textContent ?? "",
    hasSelection: () => {
      if (!editor) return false
      return editor.state.selection.from !== editor.state.selection.to
    },
    getSelectionRange: () => {
      if (!editor) return null
      const { from, to } = editor.state.selection
      return { from, to }
    },
    applyHtmlToSelection: (html: string) => {
      if (!editor) return null
      const { from, to } = editor.state.selection
      if (from === to) {
        // No selection — replace the whole document. The new full range is
        // [0, docSize) after setContent.
        editor.chain().focus().setContent(html).run()
        const size = editor.state.doc.content.size
        return { from: 0, to: size }
      }
      editor
        .chain()
        .focus()
        .deleteRange({ from, to })
        .insertContentAt(from, html)
        .run()
      // After insertContentAt, the selection sits at the end of the inserted
      // content. Compute the new [from, inserted-end) range.
      const newTo = editor.state.selection.to
      return { from, to: newTo }
    },
    applyHtmlToDoc: (html: string) => {
      if (!editor) return null
      editor.chain().focus().setContent(html).run()
      const size = editor.state.doc.content.size
      return { from: 0, to: size }
    },
    insertHtmlAtCursor: (html: string) => {
      if (!editor) return null
      const from = editor.state.selection.to
      editor.chain().focus().insertContent(html).run()
      const to = editor.state.selection.to
      return { from, to }
    },
    highlightRange: (range: EditorRange) => {
      if (!editor) return
      const size = editor.state.doc.content.size
      const from = Math.max(0, Math.min(range.from, size))
      const to = Math.max(from, Math.min(range.to, size))
      editor
        .chain()
        .focus()
        .setTextSelection({ from, to })
        .scrollIntoView()
        .run()
    },
    clearSelection: () => {
      if (!editor) return
      const pos = editor.state.selection.to
      editor.chain().focus().setTextSelection(pos).run()
    },
    focus: () => { editor?.commands.focus() },
  }), [editor])

  // Keep external content changes in sync (e.g. after save/version restore).
  // We compare against the current HTML first so the cursor position is
  // preserved during normal typing — the effect only runs when the parent
  // explicitly replaces the content.
  useEffect(() => {
    if (!editor) return
    const current = editor.getHTML()
    if (content !== current) {
      editor.commands.setContent(content || "")
    }
  }, [content, editor])

  if (!editor) {
    return (
      <div className="min-h-[500px] text-sm text-muted-foreground italic">
        Loading editor…
      </div>
    )
  }

  const hasSelection = editor.state.selection.from !== editor.state.selection.to

  return (
    <div className="flex flex-col gap-3">
      {!readOnly && (
        <Toolbar
          editor={editor}
          onAIEdit={() => onRequestAIEdit?.()}
          hasSelection={hasSelection}
        />
      )}
      <EditorContent editor={editor} />
    </div>
  )
})

// ── Toolbar ──────────────────────────────────────────────────────────────────

interface ToolbarProps {
  editor: Editor
  onAIEdit: () => void
  hasSelection: boolean
}

function Toolbar({ editor, onAIEdit, hasSelection }: ToolbarProps) {
  const btn = (active: boolean) =>
    cn(
      "h-7 w-7 rounded hover:bg-muted transition-colors",
      active && "bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300",
    )

  const handleLink = () => {
    const prev = editor.getAttributes("link").href as string | undefined
    const input = window.prompt("URL (leave blank to remove)", prev ?? "https://")
    if (input === null) return
    if (input === "") {
      editor.chain().focus().extendMarkRange("link").unsetLink().run()
    } else {
      editor.chain().focus().extendMarkRange("link").setLink({ href: input }).run()
    }
  }

  return (
    <div className="flex flex-wrap items-center gap-0.5 sticky top-0 z-10 bg-white/90 dark:bg-[#1a1d21]/90 backdrop-blur border-b py-1.5 px-1 -mx-1">
      <Button
        variant="ghost"
        size="icon"
        className={btn(editor.isActive("bold"))}
        onClick={() => editor.chain().focus().toggleBold().run()}
        title="Bold (Cmd+B)"
      >
        <Bold className="w-3.5 h-3.5" />
      </Button>
      <Button
        variant="ghost"
        size="icon"
        className={btn(editor.isActive("italic"))}
        onClick={() => editor.chain().focus().toggleItalic().run()}
        title="Italic (Cmd+I)"
      >
        <Italic className="w-3.5 h-3.5" />
      </Button>
      <Button
        variant="ghost"
        size="icon"
        className={btn(editor.isActive("strike"))}
        onClick={() => editor.chain().focus().toggleStrike().run()}
        title="Strikethrough"
      >
        <Strikethrough className="w-3.5 h-3.5" />
      </Button>
      <div className="w-px h-4 bg-border mx-1" />
      <Button
        variant="ghost"
        size="icon"
        className={btn(editor.isActive("heading", { level: 1 }))}
        onClick={() => editor.chain().focus().toggleHeading({ level: 1 }).run()}
        title="Heading 1"
      >
        <Heading1 className="w-3.5 h-3.5" />
      </Button>
      <Button
        variant="ghost"
        size="icon"
        className={btn(editor.isActive("heading", { level: 2 }))}
        onClick={() => editor.chain().focus().toggleHeading({ level: 2 }).run()}
        title="Heading 2"
      >
        <Heading2 className="w-3.5 h-3.5" />
      </Button>
      <Button
        variant="ghost"
        size="icon"
        className={btn(editor.isActive("heading", { level: 3 }))}
        onClick={() => editor.chain().focus().toggleHeading({ level: 3 }).run()}
        title="Heading 3"
      >
        <Heading3 className="w-3.5 h-3.5" />
      </Button>
      <div className="w-px h-4 bg-border mx-1" />
      <Button
        variant="ghost"
        size="icon"
        className={btn(editor.isActive("bulletList"))}
        onClick={() => editor.chain().focus().toggleBulletList().run()}
        title="Bullet list"
      >
        <List className="w-3.5 h-3.5" />
      </Button>
      <Button
        variant="ghost"
        size="icon"
        className={btn(editor.isActive("orderedList"))}
        onClick={() => editor.chain().focus().toggleOrderedList().run()}
        title="Numbered list"
      >
        <ListOrdered className="w-3.5 h-3.5" />
      </Button>
      <Button
        variant="ghost"
        size="icon"
        className={btn(editor.isActive("blockquote"))}
        onClick={() => editor.chain().focus().toggleBlockquote().run()}
        title="Quote"
      >
        <Quote className="w-3.5 h-3.5" />
      </Button>
      <Button
        variant="ghost"
        size="icon"
        className={btn(editor.isActive("codeBlock"))}
        onClick={() => editor.chain().focus().toggleCodeBlock().run()}
        title="Code block"
      >
        <Code className="w-3.5 h-3.5" />
      </Button>
      <Button
        variant="ghost"
        size="icon"
        className={btn(editor.isActive("link"))}
        onClick={handleLink}
        title="Link"
      >
        <LinkIcon className="w-3.5 h-3.5" />
      </Button>
      <div className="w-px h-4 bg-border mx-1" />
      <Button
        variant="ghost"
        size="icon"
        className="h-7 w-7 rounded hover:bg-muted"
        onClick={() => editor.chain().focus().undo().run()}
        disabled={!editor.can().undo()}
        title="Undo"
      >
        <Undo className="w-3.5 h-3.5" />
      </Button>
      <Button
        variant="ghost"
        size="icon"
        className="h-7 w-7 rounded hover:bg-muted"
        onClick={() => editor.chain().focus().redo().run()}
        disabled={!editor.can().redo()}
        title="Redo"
      >
        <Redo className="w-3.5 h-3.5" />
      </Button>

      <div className="ml-auto flex items-center gap-1">
        <Button
          size="sm"
          variant="outline"
          className="h-7 gap-1.5 text-xs font-bold border-purple-500/30 text-purple-700 dark:text-purple-300 hover:bg-purple-500/10"
          onClick={onAIEdit}
          title={hasSelection ? "Rewrite selection with AI (⌘K)" : "Ask AI (⌘K)"}
        >
          <Wand2 className="w-3 h-3" />
          {hasSelection ? "AI Rewrite" : "Ask AI"}
          <kbd className="ml-1 text-[9px] px-1 py-0 bg-muted/60 rounded font-mono">⌘K</kbd>
        </Button>
      </div>
    </div>
  )
}
