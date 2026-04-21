package fileindex

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nikkofu/relay-agent-workspace/api/internal/domain"
)

func TestExtractPlainTextContent(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "notes.txt")
	if err := os.WriteFile(path, []byte("Launch checklist\n\nShip search"), 0o644); err != nil {
		t.Fatalf("failed to write temp text file: %v", err)
	}

	svc := NewService(MockOCRProvider{})
	result := svc.ExtractFile(path, domain.FileAsset{
		ID:          "file-1",
		Name:        "notes.txt",
		StoragePath: filepath.Base(path),
		ContentType: "text/plain",
	})

	if result.Status != "ready" {
		t.Fatalf("expected ready extraction status, got %#v", result)
	}
	if result.Extractor != "plain_text" {
		t.Fatalf("expected plain_text extractor, got %#v", result)
	}
	if !strings.Contains(result.ContentText, "Launch checklist") {
		t.Fatalf("expected extracted content, got %#v", result)
	}
	if len(result.Chunks) == 0 {
		t.Fatalf("expected text extraction chunks, got %#v", result)
	}
}

func TestExtractMarkdownContent(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "brief.md")
	if err := os.WriteFile(path, []byte("# Launch\n\n## Risks\n\nWatch latency"), 0o644); err != nil {
		t.Fatalf("failed to write temp markdown file: %v", err)
	}

	svc := NewService(MockOCRProvider{})
	result := svc.ExtractFile(path, domain.FileAsset{
		ID:          "file-1",
		Name:        "brief.md",
		StoragePath: filepath.Base(path),
		ContentType: "text/markdown",
	})

	if result.Status != "ready" || result.Extractor != "plain_text" {
		t.Fatalf("expected markdown extraction to use plain_text path, got %#v", result)
	}
	if !strings.Contains(result.ContentText, "## Risks") {
		t.Fatalf("expected markdown content text, got %#v", result)
	}
}

func TestExtractUnsupportedLegacyOfficeReturnsFailure(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "legacy.doc")
	if err := os.WriteFile(path, []byte("legacy"), 0o644); err != nil {
		t.Fatalf("failed to write temp legacy file: %v", err)
	}

	svc := NewService(MockOCRProvider{})
	result := svc.ExtractFile(path, domain.FileAsset{
		ID:          "file-1",
		Name:        "legacy.doc",
		StoragePath: filepath.Base(path),
		ContentType: "application/msword",
	})

	if result.Status != "failed" || result.ErrorCode != "unsupported_legacy_format" {
		t.Fatalf("expected unsupported legacy failure, got %#v", result)
	}
}

func TestExtractImageUsesMockOCRProvider(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "diagram.png")
	if err := os.WriteFile(path, []byte("fake-image"), 0o644); err != nil {
		t.Fatalf("failed to write temp image file: %v", err)
	}

	svc := NewService(MockOCRProvider{})
	result := svc.ExtractFile(path, domain.FileAsset{
		ID:          "file-1",
		Name:        "diagram.png",
		StoragePath: filepath.Base(path),
		ContentType: "image/png",
	})

	if result.Status != "ready" || result.OCRProvider != "mock" || !result.OCRIsMock {
		t.Fatalf("expected mock OCR extraction, got %#v", result)
	}
	if !strings.Contains(result.ContentText, "Mock OCR text extracted from diagram.png") {
		t.Fatalf("expected mock OCR content, got %#v", result)
	}
}

func TestChunkTextPreservesLocatorMetadata(t *testing.T) {
	chunks := chunkText("Heading\n\nBody paragraph\n\nSecond block", "page", "3", "Heading")
	if len(chunks) < 2 {
		t.Fatalf("expected multiple chunks, got %#v", chunks)
	}
	if chunks[0].LocatorType != "page" || chunks[0].LocatorValue != "3" || chunks[0].Heading != "Heading" {
		t.Fatalf("expected locator metadata to be preserved, got %#v", chunks[0])
	}
	if chunks[0].TokenEstimate == 0 {
		t.Fatalf("expected token estimate to be populated, got %#v", chunks[0])
	}
}

func TestExtractDOCXContent(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "launch.docx")
	writeZipFile(t, path, map[string]string{
		"word/document.xml": `<document><body><p><t>Launch Plan</t></p><p><t>Owner Nikko</t></p></body></document>`,
	})

	svc := NewService(MockOCRProvider{})
	result := svc.ExtractFile(path, domain.FileAsset{
		ID:          "file-1",
		Name:        "launch.docx",
		StoragePath: filepath.Base(path),
		ContentType: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	})

	if result.Status != "ready" || result.Extractor != "docx" {
		t.Fatalf("expected docx extraction, got %#v", result)
	}
	if !strings.Contains(result.ContentText, "Launch Plan") {
		t.Fatalf("expected docx text, got %#v", result)
	}
}

func TestExtractXLSXContent(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "roadmap.xlsx")
	writeZipFile(t, path, map[string]string{
		"xl/sharedStrings.xml":     `<sst><si><t>Quarter</t></si><si><t>Launch</t></si></sst>`,
		"xl/worksheets/sheet1.xml": `<worksheet><sheetData><row><c><v>0</v></c><c><v>1</v></c></row></sheetData></worksheet>`,
	})

	svc := NewService(MockOCRProvider{})
	result := svc.ExtractFile(path, domain.FileAsset{
		ID:          "file-1",
		Name:        "roadmap.xlsx",
		StoragePath: filepath.Base(path),
		ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	})

	if result.Status != "ready" || result.Extractor != "xlsx" {
		t.Fatalf("expected xlsx extraction, got %#v", result)
	}
	if !strings.Contains(result.ContentText, "Quarter") {
		t.Fatalf("expected xlsx text, got %#v", result)
	}
}

func TestExtractPPTXContent(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "deck.pptx")
	writeZipFile(t, path, map[string]string{
		"ppt/slides/slide1.xml": `<slide><cSld><spTree><sp><txBody><p><r><t>Launch Review</t></r></p></txBody></sp></spTree></cSld></slide>`,
	})

	svc := NewService(MockOCRProvider{})
	result := svc.ExtractFile(path, domain.FileAsset{
		ID:          "file-1",
		Name:        "deck.pptx",
		StoragePath: filepath.Base(path),
		ContentType: "application/vnd.openxmlformats-officedocument.presentationml.presentation",
	})

	if result.Status != "ready" || result.Extractor != "pptx" {
		t.Fatalf("expected pptx extraction, got %#v", result)
	}
	if !strings.Contains(result.ContentText, "Launch Review") {
		t.Fatalf("expected pptx text, got %#v", result)
	}
}

func TestExtractPDFContent(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "brief.pdf")
	if err := os.WriteFile(path, []byte("%PDF-1.4\nBT (Launch Brief) Tj ET\nBT (Owner Nikko) Tj ET"), 0o644); err != nil {
		t.Fatalf("failed to write temp pdf file: %v", err)
	}

	svc := NewService(MockOCRProvider{})
	result := svc.ExtractFile(path, domain.FileAsset{
		ID:          "file-1",
		Name:        "brief.pdf",
		StoragePath: filepath.Base(path),
		ContentType: "application/pdf",
	})

	if result.Status != "ready" || result.Extractor != "pdf" {
		t.Fatalf("expected pdf extraction, got %#v", result)
	}
	if !strings.Contains(result.ContentText, "Launch Brief") {
		t.Fatalf("expected pdf text, got %#v", result)
	}
}

func writeZipFile(t *testing.T, path string, files map[string]string) {
	t.Helper()

	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create zip file: %v", err)
	}
	defer f.Close()

	writer := zip.NewWriter(f)
	for name, contents := range files {
		part, err := writer.Create(name)
		if err != nil {
			t.Fatalf("failed to create zip part %s: %v", name, err)
		}
		if _, err := bytes.NewBufferString(contents).WriteTo(part); err != nil {
			t.Fatalf("failed to write zip part %s: %v", name, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close zip writer: %v", err)
	}
}
