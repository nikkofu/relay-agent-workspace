package fileindex

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/nikkofu/relay-agent-workspace/api/internal/domain"
)

var xmlWhitespace = regexp.MustCompile(`\s+`)

func extractFile(path string, asset domain.FileAsset, ocr OCRProvider) ExtractionResult {
	ext := strings.ToLower(filepath.Ext(asset.Name))
	if ext == "" {
		ext = strings.ToLower(filepath.Ext(path))
	}

	switch ext {
	case ".txt", ".md":
		return extractPlainText(path)
	case ".doc", ".xls", ".ppt":
		return ExtractionResult{
			Status:       "failed",
			Extractor:    "unsupported",
			ErrorCode:    "unsupported_legacy_format",
			ErrorMessage: "legacy Office formats are not supported",
		}
	case ".png", ".jpg", ".jpeg", ".webp":
		return extractImageWithOCR(path, asset, ocr)
	case ".docx":
		return extractZipXMLText(path, "docx", []string{"word/document.xml"}, "document", asset.Name)
	case ".xlsx":
		return extractZipXMLText(path, "xlsx", []string{"xl/sharedStrings.xml", "xl/worksheets/"}, "sheet", asset.Name)
	case ".pptx":
		return extractZipXMLText(path, "pptx", []string{"ppt/slides/"}, "slide", asset.Name)
	case ".pdf":
		return extractPDFText(path)
	default:
		if strings.HasPrefix(strings.ToLower(asset.ContentType), "text/") {
			return extractPlainText(path)
		}
		return ExtractionResult{
			Status:       "failed",
			Extractor:    "unsupported",
			ErrorCode:    "unsupported_format",
			ErrorMessage: fmt.Sprintf("unsupported file type: %s", ext),
		}
	}
}

func extractPlainText(path string) ExtractionResult {
	raw, err := os.ReadFile(path)
	if err != nil {
		return ExtractionResult{
			Status:       "failed",
			Extractor:    "plain_text",
			ErrorCode:    "read_failed",
			ErrorMessage: err.Error(),
		}
	}

	text := strings.TrimSpace(string(raw))
	if text == "" {
		return ExtractionResult{
			Status:       "failed",
			Extractor:    "plain_text",
			ErrorCode:    "empty_extractable_text",
			ErrorMessage: "file does not contain extractable text",
		}
	}

	return ExtractionResult{
		Status:         "ready",
		Extractor:      "plain_text",
		ContentText:    text,
		ContentSummary: summarizeText(text),
		Chunks:         chunkText(text, "document", "root", firstHeading(text)),
	}
}

func extractImageWithOCR(path string, asset domain.FileAsset, ocr OCRProvider) ExtractionResult {
	result, err := ocr.Extract(path, asset.Name)
	if err != nil {
		return ExtractionResult{
			Status:       "failed",
			Extractor:    "ocr",
			ErrorCode:    "ocr_failed",
			ErrorMessage: err.Error(),
			NeedsOCR:     true,
		}
	}

	text := strings.TrimSpace(result.Text)
	return ExtractionResult{
		Status:         "ready",
		Extractor:      "ocr-mock",
		ContentText:    text,
		ContentSummary: summarizeText(text),
		NeedsOCR:       true,
		OCRProvider:    result.Provider,
		OCRIsMock:      result.IsMock,
		Chunks:         chunkText(text, "document", "root", firstHeading(text)),
	}
}

func extractZipXMLText(path, extractor string, prefixes []string, locatorType string, assetName string) ExtractionResult {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return ExtractionResult{
			Status:       "failed",
			Extractor:    extractor,
			ErrorCode:    "archive_open_failed",
			ErrorMessage: err.Error(),
		}
	}
	defer reader.Close()

	parts := make([]string, 0)
	for _, file := range reader.File {
		if !matchesAnyPrefix(file.Name, prefixes) {
			continue
		}
		rc, err := file.Open()
		if err != nil {
			continue
		}
		raw, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			continue
		}
		text := extractXMLText(raw)
		if strings.TrimSpace(text) == "" {
			continue
		}
		parts = append(parts, text)
	}

	if len(parts) == 0 {
		return ExtractionResult{
			Status:       "failed",
			Extractor:    extractor,
			ErrorCode:    "empty_extractable_text",
			ErrorMessage: fmt.Sprintf("%s does not contain extractable text", assetName),
		}
	}

	text := strings.Join(parts, "\n\n")
	return ExtractionResult{
		Status:         "ready",
		Extractor:      extractor,
		ContentText:    text,
		ContentSummary: summarizeText(text),
		Chunks:         chunkText(text, locatorType, "root", firstHeading(text)),
	}
}

func extractPDFText(path string) ExtractionResult {
	// Fallback parser for text-based PDFs. This intentionally prefers deterministic,
	// best-effort extraction over perfect layout fidelity in Phase 44.
	raw, err := os.ReadFile(path)
	if err != nil {
		return ExtractionResult{
			Status:       "failed",
			Extractor:    "pdf",
			ErrorCode:    "read_failed",
			ErrorMessage: err.Error(),
		}
	}

	text := collapsePDFText(raw)
	if strings.TrimSpace(text) == "" {
		return ExtractionResult{
			Status:       "failed",
			Extractor:    "pdf",
			ErrorCode:    "empty_extractable_text",
			ErrorMessage: "pdf does not contain extractable text",
			NeedsOCR:     true,
		}
	}

	return ExtractionResult{
		Status:         "ready",
		Extractor:      "pdf",
		ContentText:    text,
		ContentSummary: summarizeText(text),
		Chunks:         chunkText(text, "document", "root", firstHeading(text)),
	}
}

func collapsePDFText(raw []byte) string {
	// Extremely lightweight PDF string extraction: pull strings from parentheses.
	// This is sufficient for simple text PDFs and keeps the initial implementation dependency-light.
	parts := make([]string, 0)
	inText := false
	escaped := false
	var current bytes.Buffer
	for _, b := range raw {
		switch {
		case escaped:
			current.WriteByte(b)
			escaped = false
		case b == '\\' && inText:
			escaped = true
		case b == '(':
			inText = true
			current.Reset()
		case b == ')' && inText:
			inText = false
			if s := strings.TrimSpace(current.String()); s != "" {
				parts = append(parts, s)
			}
		case inText:
			current.WriteByte(b)
		}
	}
	return xmlWhitespace.ReplaceAllString(strings.Join(parts, "\n"), " ")
}

func extractXMLText(raw []byte) string {
	decoder := xml.NewDecoder(bytes.NewReader(raw))
	var parts []string
	for {
		token, err := decoder.Token()
		if err != nil {
			break
		}
		switch value := token.(type) {
		case xml.CharData:
			if text := strings.TrimSpace(string(value)); text != "" {
				parts = append(parts, text)
			}
		}
	}
	return xmlWhitespace.ReplaceAllString(strings.Join(parts, " "), " ")
}

func matchesAnyPrefix(value string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return false
}

func summarizeText(text string) string {
	fields := strings.Fields(text)
	if len(fields) == 0 {
		return ""
	}
	if len(fields) > 30 {
		fields = fields[:30]
	}
	return strings.Join(fields, " ")
}

func firstHeading(text string) string {
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "#"), "#"))
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}
