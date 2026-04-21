package knowledge

type Citation struct {
	ID           string  `json:"id"`
	EvidenceKind string  `json:"evidence_kind"`
	SourceKind   string  `json:"source_kind"`
	SourceRef    string  `json:"source_ref"`
	RefKind      string  `json:"ref_kind"`
	Locator      string  `json:"locator,omitempty"`
	Snippet      string  `json:"snippet"`
	Title        string  `json:"title,omitempty"`
	Score        float64 `json:"score"`
	EntityID     string  `json:"entity_id,omitempty"`
	EntityTitle  string  `json:"entity_title,omitempty"`
}

type LookupParams struct {
	Query        string
	ChannelID    string
	EntityID     string
	Limit        int
	IncludeKinds map[string]bool
}
