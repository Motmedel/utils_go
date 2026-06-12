package usage_metadata

type UsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount,omitempty"`
	CandidatesTokenCount int `json:"candidatesTokenCount,omitempty"`
	ThoughtsTokenCount   int `json:"thoughtsTokenCount,omitempty"`
	TotalTokenCount      int `json:"totalTokenCount,omitempty"`
}
