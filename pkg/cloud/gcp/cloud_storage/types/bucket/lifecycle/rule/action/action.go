package action

type Action struct {
	Type         string `json:"type,omitempty"`
	StorageClass string `json:"storageClass,omitempty"`
}
