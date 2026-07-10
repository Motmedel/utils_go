package token

type Token struct {
	Kind string `json:"kind,omitempty"`
	Etag string `json:"etag,omitempty"`

	ClientId    string   `json:"clientId,omitempty"`
	DisplayText string   `json:"displayText,omitempty"`
	UserKey     string   `json:"userKey,omitempty"`
	Scopes      []string `json:"scopes,omitempty"`

	Anonymous bool `json:"anonymous,omitempty"`
	NativeApp bool `json:"nativeApp,omitempty"`
}
