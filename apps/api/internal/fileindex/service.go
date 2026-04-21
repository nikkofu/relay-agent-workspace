package fileindex

import "github.com/nikkofu/relay-agent-workspace/api/internal/domain"

type Service struct {
	OCR OCRProvider
}

func NewService(ocr OCRProvider) *Service {
	return &Service{OCR: ocr}
}

func (s *Service) ExtractFile(path string, asset domain.FileAsset) ExtractionResult {
	ocr := s.OCR
	if ocr == nil {
		ocr = MockOCRProvider{}
	}
	return extractFile(path, asset, ocr)
}
