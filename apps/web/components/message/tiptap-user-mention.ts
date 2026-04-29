import { Node, mergeAttributes } from "@tiptap/core"
import type { Editor } from "@tiptap/core"

export interface UserMentionAttrs {
  kind: "user"
  userId: string
  name: string
  userType?: "human" | "bot" | "ai"
  mentionText: string
}

declare module "@tiptap/core" {
  interface Commands<ReturnType> {
    userMention: {
      insertUserMention: (attrs: UserMentionAttrs) => ReturnType
    }
  }
}

const mentionTextOf = (attrs: Partial<UserMentionAttrs>) => {
  if (typeof attrs.mentionText === "string" && attrs.mentionText.trim().length > 0) {
    return attrs.mentionText
  }
  if (typeof attrs.name === "string" && attrs.name.trim().length > 0) {
    return `@${attrs.name}`
  }
  return "@mention"
}

const deleteSelectedMention = (editor: Editor) => {
  const { state } = editor
  const node = state.doc.nodeAt(state.selection.from)
  if (!node || node.type.name !== "userMention") return false
  editor.chain().focus().deleteRange({ from: state.selection.from, to: state.selection.to }).run()
  return true
}

export const UserMention = Node.create({
  name: "userMention",
  inline: true,
  group: "inline",
  atom: true,
  selectable: true,

  addAttributes() {
    return {
      kind: {
        default: "user",
        parseHTML: element => element.getAttribute("data-mention-kind") || "user",
      },
      userId: {
        default: "",
        parseHTML: element => element.getAttribute("data-mention-user-id") || "",
      },
      name: {
        default: "",
        parseHTML: element => element.getAttribute("data-mention-name") || element.textContent?.replace(/^@/, "") || "",
      },
      userType: {
        default: undefined,
        parseHTML: element => element.getAttribute("data-mention-user-type") || undefined,
      },
      mentionText: {
        default: "",
        parseHTML: element => element.textContent || "",
      },
    }
  },

  parseHTML() {
    return [{ tag: 'span[data-mention-kind="user"][data-mention-user-id]' }]
  },

  renderHTML({ HTMLAttributes }) {
    const mentionText = mentionTextOf(HTMLAttributes as Partial<UserMentionAttrs>)
    return [
      "span",
      mergeAttributes(HTMLAttributes, {
        "data-mention-kind": "user",
        "data-mention-user-id": HTMLAttributes.userId,
        "data-mention-name": HTMLAttributes.name,
        "data-mention-user-type": HTMLAttributes.userType,
        contenteditable: "false",
        class: "inline-flex items-center rounded-md border border-fuchsia-400/30 bg-fuchsia-500/10 px-1.5 py-0.5 font-semibold text-fuchsia-700 dark:text-fuchsia-300",
      }),
      mentionText,
    ]
  },

  renderText({ node }) {
    return mentionTextOf(node.attrs as Partial<UserMentionAttrs>)
  },

  addCommands() {
    return {
      insertUserMention:
        (attrs: UserMentionAttrs) =>
        ({ chain }) =>
          chain().insertContent([
            {
              type: this.name,
              attrs,
            },
            {
              type: "text",
              text: " ",
            },
          ]).run(),
    }
  },

  addKeyboardShortcuts() {
    return {
      Backspace: () => {
        const { selection } = this.editor.state
        if (!selection.empty) {
          return deleteSelectedMention(this.editor)
        }
        const nodeBefore = selection.$from.nodeBefore
        if (nodeBefore?.type.name !== this.name) return false
        const from = selection.$from.pos - nodeBefore.nodeSize
        this.editor.chain().focus().deleteRange({ from, to: selection.$from.pos }).run()
        return true
      },
      Delete: () => {
        const { selection } = this.editor.state
        if (!selection.empty) {
          return deleteSelectedMention(this.editor)
        }
        const nodeAfter = selection.$from.nodeAfter
        if (nodeAfter?.type.name !== this.name) return false
        this.editor.chain().focus().deleteRange({ from: selection.$from.pos, to: selection.$from.pos + nodeAfter.nodeSize }).run()
        return true
      },
    }
  },
})
