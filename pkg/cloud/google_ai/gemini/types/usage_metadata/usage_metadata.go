package usage_metadata

type UsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount,omitzero"`
	CandidatesTokenCount int `json:"candidatesTokenCount,omitzero"`
	ThoughtsTokenCount   int `json:"thoughtsTokenCount,omitzero"`
	TotalTokenCount      int `json:"totalTokenCount,omitzero"`
}
