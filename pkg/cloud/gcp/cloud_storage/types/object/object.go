package object

type Object struct {
	Kind            string            `json:"kind,omitempty"`
	Id              string            `json:"id,omitempty"`
	SelfLink        string            `json:"selfLink,omitempty"`
	Name            string            `json:"name,omitempty"`
	Bucket          string            `json:"bucket,omitempty"`
	Generation      string            `json:"generation,omitempty"`
	Metageneration  string            `json:"metageneration,omitempty"`
	ContentType     string            `json:"contentType,omitempty"`
	TimeCreated     string            `json:"timeCreated,omitempty"`
	Updated         string            `json:"updated,omitempty"`
	StorageClass    string            `json:"storageClass,omitempty"`
	Size            string            `json:"size,omitempty"`
	Md5Hash         string            `json:"md5Hash,omitempty"`
	MediaLink       string            `json:"mediaLink,omitempty"`
	Crc32c          string            `json:"crc32c,omitempty"`
	Etag            string            `json:"etag,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}
