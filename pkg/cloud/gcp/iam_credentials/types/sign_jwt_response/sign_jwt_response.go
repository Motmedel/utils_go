package sign_jwt_response

type Response struct {
	KeyId     string `json:"keyId"`
	SignedJwt string `json:"signedJwt"`
}
