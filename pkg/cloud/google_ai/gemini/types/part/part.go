package part

type Part struct {
	Text    string `json:"text,omitempty"`
	Thought bool   `json:"thought,omitzero"`
}
