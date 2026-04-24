"use client"

// ── Canvas TipTap editor (Bug fix) ────────────────────────────────────────────
//
// Replaces the plain <textarea> previously used for document-type canvases.
// Emits HTML via `editor.getHTML()` so it stays wire-compatible with the
// existing `dangerouslySetInnerHTML` view-mode renderer and with the server's
// Artifact.Content field.
//
// Includes a fixed toolbar and a purple "AI Edit" button that opens a small
// prompt-driven dialog (see canvas-ai-edit-dialog.tsx). AI edits operate on
// the current selection when one exists, otherwise on the whole document.

import { useEffect, useRef, useState } from "react"
import { useEditor, EditorContent, Editor } from "@tiptap/react"
import StarterKit from "@tiptap/starter-kit"
import Placeholder from "@tiptap/extension-placeholder"
import TiptapLink from "@tiptap/extension-link"
import {
  Bold, Italic, Strikethrough, Heading1, Heading2, Heading3,
  List, ListOrdered, Quote, Code, Link as LinkIcon, Undo, Redo,
  Wand2,
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"
import { CanvasAIEditDialog } from "./canvas-ai-edit-dialog"

interface CanvasTipTapEditorProps {
  content: string
  onChange: (html: string) => void
  channelId: string
  autoFocus?: boolean
  readOnly?: boolean
  placeholder?: string
}

export function CanvasTipTapEditor({
  content,
  onChange,
  channelId,
  autoFocus = true,
  readOnly = false,
  placeholder = "Start writing, or press the AI button to draft something…",
}: CanvasTipTapEditorProps) {
  const [showAIDialog, setShowAIDialog] = useState(false)
  // Track the selection range to apply AI edit to (from, to). When the user
  // has no selection, these default to (0, doc.size) meaning "whole document".
  const aiTargetRef = useRef<{ from: number; to: number } | null>(null)

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
    ],
    content: content || "",
    editable: !readOnly,
    autofocus: autoFocus ? "end" : false,
    immediatelyRender: false,
    onUpdate: ({ editor }) => {
      onChange(editor.getHTML())
    },
    editorProps: {
      attributes: {
        class:
          "prose dark:prose-invert max-w-none focus:outline-none min-h-[500px] px-1",
      },
    },
  })

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
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [content])

  if (!editor) {
    return (
      <div className="min-h-[500px] text-sm text-muted-foreground italic">
        Loading editor…
      </div>
    )
  }

  const openAIEdit = () => {
    const { from, to } = editor.state.selection
    if (from === to) {
      aiTargetRef.current = { from: 0, to: editor.state.doc.content.size }
    } else {
      aiTargetRef.current = { from, to }
    }
    setShowAIDialog(true)
  }

  const applyAIResult = (html: string) => {
    if (!editor || !aiTargetRef.current) return
    const { from, to } = aiTargetRef.current
    editor
      .chain()
      .focus()
      .deleteRange({ from, to })
      .insertContentAt(from, html)
      .run()
    aiTargetRef.current = null
  }

  // Text captured for the dialog: selection text (if any) or whole doc text.
  const aiTargetText = (() => {
    if (!aiTargetRef.current) return editor.state.doc.textContent
    const { from, to } = aiTargetRef.current
    return editor.state.doc.textBetween(from, to, "\n\n")
  })()

  const hasSelection = editor.state.selection.from !== editor.state.selection.to

  return (
    <div className="flex flex-col gap-3">
      {!readOnly && (
        <Toolbar editor={editor} onAIEdit={openAIEdit} hasSelection={hasSelection} />
      )}
      <EditorContent editor={editor} />
      <CanvasAIEditDialog
        open={showAIDialog}
        onOpenChange={setShowAIDialog}
        channelId={channelId}
        targetText={aiTargetText}
        onApply={applyAIResult}
      />
    </div>
  )
}

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
          title={hasSelection ? "Rewrite selection with AI" : "Rewrite document with AI"}
        >
          <Wand2 className="w-3 h-3" />
          AI {hasSelection ? "Rewrite selection" : "Edit"}
        </Button>
      </div>
    </div>
  )
}
