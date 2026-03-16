package resource_search_result

type ResourceSearchResult struct {
	Name                   string            `json:"name,omitempty"`
	AssetType              string            `json:"assetType,omitempty"`
	Project                string            `json:"project,omitempty"`
	Folders                []string          `json:"folders,omitempty"`
	Organization           string            `json:"organization,omitempty"`
	DisplayName            string            `json:"displayName,omitempty"`
	Description            string            `json:"description,omitempty"`
	Location               string            `json:"location,omitempty"`
	Labels                 map[string]string `json:"labels,omitempty"`
	NetworkTags            []string          `json:"networkTags,omitempty"`
	KmsKeys                []string          `json:"kmsKeys,omitempty"`
	CreateTime             string            `json:"createTime,omitempty"`
	UpdateTime             string            `json:"updateTime,omitempty"`
	State                  string            `json:"state,omitempty"`
	ParentFullResourceName string            `json:"parentFullResourceName,omitempty"`
	ParentAssetType        string            `json:"parentAssetType,omitempty"`
	TagKeys                []string          `json:"tagKeys,omitempty"`
	TagValues              []string          `json:"tagValues,omitempty"`
}
