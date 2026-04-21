package fileindex

type ExtractionResult struct {
	Status         string
	Extractor      string
	ContentText    string
	ContentSummary string
	ErrorCode      string
	ErrorMessage   string
	NeedsOCR       bool
	OCRProvider    string
	OCRIsMock      bool
	Chunks         []Chunk
}

type Chunk struct {
	Text          string
	TokenEstimate int
	LocatorType   string
	LocatorValue  string
	Heading       string
}

type OCRResult struct {
	Text       string
	Provider   string
	IsMock     bool
	Confidence string
}
