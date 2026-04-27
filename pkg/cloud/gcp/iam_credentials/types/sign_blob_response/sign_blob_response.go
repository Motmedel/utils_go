package sign_blob_response

type Response struct {
	KeyId      string `json:"keyId"`
	SignedBlob string `json:"signedBlob"`
}
