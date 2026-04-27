package sign_blob_request

type Request struct {
	Payload   string   `json:"payload"`
	Delegates []string `json:"delegates,omitempty"`
}
