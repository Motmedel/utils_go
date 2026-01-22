package tokenapi

type Token interface {
	HeaderFields() map[string]any
	Claims() map[string]any
}
