package asp

type Asp struct {
	Kind string `json:"kind,omitempty"`
	Etag string `json:"etag,omitempty"`

	CodeId int    `json:"codeId,omitempty"`
	Name   string `json:"name,omitempty"`

	CreationTime string `json:"creationTime,omitempty"`
	LastTimeUsed string `json:"lastTimeUsed,omitempty"`
}
