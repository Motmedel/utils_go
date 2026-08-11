package resource

import "encoding/json/jsontext"

type Resource struct {
	Version              string         `json:"version,omitempty"`
	DiscoveryDocumentUri string         `json:"discoveryDocumentUri,omitempty"`
	DiscoveryName        string         `json:"discoveryName,omitempty"`
	ResourceUrl          string         `json:"resourceUrl,omitempty"`
	Parent               string         `json:"parent,omitempty"`
	Data                 jsontext.Value `json:"data,omitempty"`
	Location             string         `json:"location,omitempty"`
}
