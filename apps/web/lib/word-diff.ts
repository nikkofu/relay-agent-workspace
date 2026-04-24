// ── Word-level diff (Office Review style) ───────────────────────────────────
//
// Small, dependency-free Myers-style LCS diff over tokens (words +
// punctuation + whitespace). Returns a flat sequence of `DiffToken`s the UI
// can render inline with strikethrough for deletions and underline for
// insertions — exactly what Microsoft Word's Review pane shows.
//
// The diff is tokenized so whitespace and punctuation align naturally
// instead of the whole sentence being marked dirty on a single-word edit.

export type DiffTokenKind = "equal" | "insert" | "delete"

export interface DiffToken {
  kind: DiffTokenKind
  text: string
}

/** Run a word-level diff between two plain-text strings. */
export function diffWords(from: string, to: string): DiffToken[] {
  const a = tokenize(from)
  const b = tokenize(to)
  if (a.length === 0 && b.length === 0) return []
  if (a.length === 0) return b.map(text => ({ kind: "insert", text }))
  if (b.length === 0) return a.map(text => ({ kind: "delete", text }))

  // Classic LCS DP over tokens. Inputs here are user-written docs
  // (paragraphs, not megabytes) so O(N·M) is fine and avoids pulling a
  // diff library into the web bundle.
  const n = a.length
  const m = b.length
  const dp: number[][] = Array.from({ length: n + 1 }, () =>
    new Array<number>(m + 1).fill(0),
  )
  for (let i = 1; i <= n; i++) {
    for (let j = 1; j <= m; j++) {
      dp[i][j] = a[i - 1] === b[j - 1]
        ? dp[i - 1][j - 1] + 1
        : Math.max(dp[i - 1][j], dp[i][j - 1])
    }
  }

  // Walk back through the DP table, producing a reversed token stream.
  const out: DiffToken[] = []
  let i = n
  let j = m
  while (i > 0 && j > 0) {
    if (a[i - 1] === b[j - 1]) {
      out.push({ kind: "equal", text: a[i - 1] })
      i--; j--
    } else if (dp[i - 1][j] >= dp[i][j - 1]) {
      out.push({ kind: "delete", text: a[i - 1] })
      i--
    } else {
      out.push({ kind: "insert", text: b[j - 1] })
      j--
    }
  }
  while (i > 0) { out.push({ kind: "delete", text: a[i - 1] }); i-- }
  while (j > 0) { out.push({ kind: "insert", text: b[j - 1] }); j-- }

  out.reverse()
  return mergeAdjacent(out)
}

// ── Internal helpers ─────────────────────────────────────────────────────────

// Tokenize into: words, whitespace runs, single punctuation characters, or
// lone newlines. Keeping each class separate means a single-word swap
// inside a sentence does not paint the whole line as changed.
function tokenize(input: string): string[] {
  if (!input) return []
  const tokens: string[] = []
  const re = /(\n)|([ \t]+)|([A-Za-z0-9\u00C0-\uFFFF_]+)|([^\s])/g
  let match: RegExpExecArray | null
  while ((match = re.exec(input)) !== null) tokens.push(match[0])
  return tokens
}

// Collapse runs of same-kind tokens so downstream rendering is cheap +
// diff-pill chunks look natural in the UI ("brown fox" not "brown"," ","fox").
function mergeAdjacent(tokens: DiffToken[]): DiffToken[] {
  const out: DiffToken[] = []
  for (const t of tokens) {
    const last = out[out.length - 1]
    if (last && last.kind === t.kind) {
      last.text += t.text
    } else {
      out.push({ kind: t.kind, text: t.text })
    }
  }
  return out
}

/** Strip HTML markup to plain text so HTML-backed content (TipTap output)
 *  can be compared without `<p>`/`<br>` noise leaking into the diff. Falls
 *  back to a regex strip on non-DOM environments (SSR). */
export function htmlToPlainText(html: string): string {
  if (!html) return ""
  if (typeof document !== "undefined") {
    const tmp = document.createElement("div")
    tmp.innerHTML = html
    // Normalize paragraph breaks into actual newlines so the diff aligns
    // with the visual layout of the rendered doc.
    tmp.querySelectorAll("br").forEach(br => br.replaceWith("\n"))
    tmp.querySelectorAll("p, h1, h2, h3, h4, h5, h6, li, blockquote, pre").forEach(el => {
      el.append(document.createTextNode("\n"))
    })
    return (tmp.textContent || "").replace(/\n{3,}/g, "\n\n").trim()
  }
  return html
    .replace(/<br\s*\/?>/gi, "\n")
    .replace(/<\/(p|h[1-6]|li|blockquote|pre)>/gi, "\n")
    .replace(/<[^>]+>/g, "")
    .replace(/\n{3,}/g, "\n\n")
    .trim()
}
