package types

type TestEvent struct {
	Time    string  `json:"Time,omitempty"`
	Action  string  `json:"Action,omitempty"`
	Package string  `json:"Package,omitempty"`
	Test    string  `json:"Test,omitempty"`
	Elapsed float64 `json:"Elapsed,omitempty"`
	Output  string  `json:"Output,omitempty"`
}
