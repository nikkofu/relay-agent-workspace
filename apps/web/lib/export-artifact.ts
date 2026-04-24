// ── Artifact Export Helpers ──────────────────────────────────────────────────
//
// Client-side export of a document-type artifact to .doc (Word) and .pdf.
// Intentionally zero-dependency: we already render TipTap HTML in the canvas
// so we reuse it.
//
//   • Word: a Microsoft Word-compatible HTML blob saved with a .doc
//     extension. Word (and Google Docs / Pages) open it as a rich document.
//     This is the widely-used "Altchunk HTML" trick — lighter than pulling
//     in docx.js.
//
//   • PDF: we open a same-origin child window, write a minimal HTML
//     document with the artifact content, and trigger `window.print()` so
//     the user's OS print dialog lets them "Save as PDF". All modern
//     browsers render this correctly and it avoids shipping a 200+ kB PDF
//     library in the web bundle.

/** Download a .doc (Word-compatible HTML) file for the given artifact. */
export function exportArtifactAsWord(args: {
  title: string
  html: string
}): void {
  const { title, html } = args
  const safeTitle = title.trim() || "Untitled"
  const doc = buildWordHtml(safeTitle, html)
  const blob = new Blob(["\uFEFF", doc], {
    type: "application/msword;charset=utf-8",
  })
  triggerDownload(blob, `${sanitizeFilename(safeTitle)}.doc`)
}

/** Open a print-ready window for the artifact and call `window.print()` so
 *  the user can "Save as PDF" from the browser print dialog. Returns `true`
 *  on success, `false` when the popup was blocked. */
export function exportArtifactAsPDF(args: {
  title: string
  html: string
}): boolean {
  const { title, html } = args
  const safeTitle = title.trim() || "Untitled"
  const printWin = window.open("", "_blank", "noopener,noreferrer,width=900,height=1100")
  if (!printWin) return false

  printWin.document.open()
  printWin.document.write(buildPdfHtml(safeTitle, html))
  printWin.document.close()

  // Give the browser a tick to render + load web fonts before invoking print.
  printWin.addEventListener("load", () => {
    setTimeout(() => {
      try {
        printWin.focus()
        printWin.print()
      } catch {
        // ignore — some browsers block print from opener; user can still Cmd+P.
      }
    }, 150)
  })
  return true
}

// ── Internal builders ────────────────────────────────────────────────────────

function buildWordHtml(title: string, bodyHtml: string): string {
  // Word's HTML-import parser wants a `<html xmlns:office>` envelope with a
  // ProgId meta so it knows to treat the file as a Word document rather
  // than a web page. The CSS is kept minimal + print-safe so Word doesn't
  // fight us on layout.
  return `<!DOCTYPE html>
<html xmlns:o="urn:schemas-microsoft-com:office:office"
      xmlns:w="urn:schemas-microsoft-com:office:word"
      xmlns="http://www.w3.org/TR/REC-html40">
<head>
  <meta charset="utf-8" />
  <meta name="ProgId" content="Word.Document" />
  <meta name="Generator" content="Microsoft Word 15" />
  <meta name="Originator" content="Microsoft Word 15" />
  <title>${escapeHtml(title)}</title>
  <style>
    @page { margin: 2.54cm; }
    body { font-family: 'Calibri', 'Segoe UI', sans-serif; font-size: 11pt; line-height: 1.5; color: #1a1a1a; }
    h1 { font-size: 22pt; font-weight: 700; margin: 0 0 8pt; }
    h2 { font-size: 16pt; font-weight: 700; margin: 16pt 0 6pt; }
    h3 { font-size: 13pt; font-weight: 700; margin: 14pt 0 4pt; }
    p { margin: 0 0 8pt; }
    ul, ol { margin: 0 0 8pt 20pt; padding: 0; }
    li { margin: 0 0 4pt; }
    blockquote { margin: 0 0 8pt; padding: 6pt 12pt; border-left: 3pt solid #6b46c1; color: #444; font-style: italic; }
    code { font-family: 'Consolas', 'Courier New', monospace; background: #f4f4f5; padding: 1pt 4pt; border-radius: 2pt; }
    pre { font-family: 'Consolas', 'Courier New', monospace; background: #f4f4f5; padding: 8pt; border-radius: 2pt; }
    a { color: #1164a3; text-decoration: underline; }
    hr { border: none; border-top: 1pt solid #d4d4d8; margin: 16pt 0; }
  </style>
</head>
<body>
  <h1>${escapeHtml(title)}</h1>
  ${bodyHtml}
</body>
</html>`
}

function buildPdfHtml(title: string, bodyHtml: string): string {
  // Same HTML-first strategy as Word, but targeted at the browser's print
  // pipeline. The print button auto-clicks once the window finishes loading.
  return `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <title>${escapeHtml(title)}</title>
  <style>
    @page { size: A4; margin: 20mm 18mm; }
    html, body { background: #ffffff; color: #111827; }
    body {
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Inter', system-ui, sans-serif;
      font-size: 11pt; line-height: 1.55;
      margin: 0; padding: 24pt 32pt;
      max-width: 780px; margin-left: auto; margin-right: auto;
    }
    h1 { font-size: 26pt; font-weight: 800; letter-spacing: -0.5pt; margin: 0 0 4pt; }
    h2 { font-size: 18pt; font-weight: 700; margin: 18pt 0 6pt; }
    h3 { font-size: 14pt; font-weight: 700; margin: 14pt 0 4pt; }
    p { margin: 0 0 8pt; }
    ul, ol { margin: 0 0 10pt 24pt; padding: 0; }
    li { margin: 0 0 4pt; }
    blockquote { margin: 0 0 10pt; padding: 8pt 14pt; border-left: 3pt solid #a855f7; color: #4b5563; font-style: italic; background: #faf5ff; border-radius: 2pt; }
    code { font-family: 'SF Mono', 'Consolas', monospace; background: #f4f4f5; padding: 1pt 5pt; border-radius: 3pt; font-size: 10pt; }
    pre { font-family: 'SF Mono', 'Consolas', monospace; background: #f8fafc; padding: 10pt 12pt; border-radius: 4pt; overflow-x: auto; font-size: 10pt; line-height: 1.4; border: 1pt solid #e4e4e7; }
    a { color: #1164a3; text-decoration: underline; }
    hr { border: none; border-top: 1pt solid #e4e4e7; margin: 18pt 0; }
    .subtitle { font-size: 9pt; color: #6b7280; letter-spacing: 0.3pt; margin-bottom: 18pt; text-transform: uppercase; font-weight: 700; }
    @media print {
      body { padding: 0; }
      .subtitle { display: none; }
    }
  </style>
</head>
<body>
  <h1>${escapeHtml(title)}</h1>
  <div class="subtitle">Exported from Relay Agent Workspace · ${new Date().toLocaleString()}</div>
  ${bodyHtml}
</body>
</html>`
}

function sanitizeFilename(name: string): string {
  return (name || "document")
    .replace(/[/\\?%*:|"<>]/g, "-")
    .replace(/\s+/g, " ")
    .trim()
    .slice(0, 80) || "document"
}

function escapeHtml(s: string): string {
  return s
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
}

function triggerDownload(blob: Blob, filename: string): void {
  const url = URL.createObjectURL(blob)
  const a = document.createElement("a")
  a.href = url
  a.download = filename
  a.rel = "noopener"
  document.body.appendChild(a)
  a.click()
  setTimeout(() => {
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  }, 0)
}
