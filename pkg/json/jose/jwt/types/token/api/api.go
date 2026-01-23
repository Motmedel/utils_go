package api

type Token interface {
	HeaderFields() map[string]any
	Claims() map[string]any
}
