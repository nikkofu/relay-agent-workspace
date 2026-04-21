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

type CreateEntityInput struct {
	WorkspaceID string `json:"workspace_id"`
	Kind        string `json:"kind"`
	Title       string `json:"title"`
	Summary     string `json:"summary"`
	Status      string `json:"status"`
	OwnerUserID string `json:"owner_user_id"`
	SourceKind  string `json:"source_kind"`
	SourceRef   string `json:"source_ref"`
	Metadata    string `json:"metadata_json"`
}

type UpdateEntityInput struct {
	Kind        string `json:"kind"`
	Title       string `json:"title"`
	Summary     string `json:"summary"`
	Status      string `json:"status"`
	OwnerUserID string `json:"owner_user_id"`
	SourceKind  string `json:"source_kind"`
	SourceRef   string `json:"source_ref"`
	Metadata    string `json:"metadata_json"`
}

type AddEntityRefInput struct {
	RefKind  string `json:"ref_kind"`
	RefID    string `json:"ref_id"`
	Role     string `json:"role"`
	Metadata string `json:"metadata_json"`
}

type AddEntityLinkInput struct {
	WorkspaceID  string  `json:"workspace_id"`
	FromEntityID string  `json:"from_entity_id"`
	ToEntityID   string  `json:"to_entity_id"`
	Relation     string  `json:"relation"`
	Weight       float64 `json:"weight"`
	Metadata     string  `json:"metadata_json"`
}

type AddEntityEventInput struct {
	EventType   string `json:"event_type"`
	Title       string `json:"title"`
	Body        string `json:"body"`
	ActorUserID string `json:"actor_user_id"`
	SourceKind  string `json:"source_kind"`
	SourceRef   string `json:"source_ref"`
}

type GraphNode struct {
	ID    string `json:"id"`
	Kind  string `json:"kind"`
	Title string `json:"title"`
}

type GraphEdge struct {
	ID       string `json:"id"`
	From     string `json:"from"`
	To       string `json:"to"`
	Relation string `json:"relation"`
}

type EntityGraph struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}
