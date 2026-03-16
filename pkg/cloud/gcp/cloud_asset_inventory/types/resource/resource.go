package resource

type Resource struct {
	Version              string         `json:"version,omitempty"`
	DiscoveryDocumentUri string         `json:"discoveryDocumentUri,omitempty"`
	DiscoveryName        string         `json:"discoveryName,omitempty"`
	ResourceUrl          string         `json:"resourceUrl,omitempty"`
	Parent               string         `json:"parent,omitempty"`
	Data                 map[string]any `json:"data,omitempty"`
	Location             string         `json:"location,omitempty"`
}
