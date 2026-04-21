package fileindex

import "fmt"

type OCRProvider interface {
	Name() string
	Extract(path string, assetName string) (OCRResult, error)
}

type MockOCRProvider struct{}

func (MockOCRProvider) Name() string {
	return "mock"
}

func (MockOCRProvider) Extract(_ string, assetName string) (OCRResult, error) {
	return OCRResult{
		Text:       fmt.Sprintf("Mock OCR text extracted from %s", assetName),
		Provider:   "mock",
		IsMock:     true,
		Confidence: "low",
	}, nil
}
