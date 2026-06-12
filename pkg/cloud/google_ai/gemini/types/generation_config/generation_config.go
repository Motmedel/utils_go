package generation_config

import (
	"github.com/Motmedel/utils_go/pkg/cloud/google_ai/gemini/types/thinking_config"
)

type GenerationConfig struct {
	Temperature     *float64 `json:"temperature,omitempty"`
	TopP            *float64 `json:"topP,omitempty"`
	TopK            *int     `json:"topK,omitempty"`
	CandidateCount  int      `json:"candidateCount,omitempty"`
	MaxOutputTokens int      `json:"maxOutputTokens,omitempty"`
	StopSequences   []string `json:"stopSequences,omitempty"`
	// ResponseMimeType must be "application/json" for ResponseSchema or
	// ResponseJsonSchema to take effect.
	ResponseMimeType string `json:"responseMimeType,omitempty"`
	// ResponseSchema is the OpenAPI-subset schema format native to the Gemini API.
	ResponseSchema any `json:"responseSchema,omitempty"`
	// ResponseJsonSchema accepts standard JSON Schema (Gemini 2.5 and later).
	ResponseJsonSchema any                             `json:"responseJsonSchema,omitempty"`
	ThinkingConfig     *thinking_config.ThinkingConfig `json:"thinkingConfig,omitempty"`
}
