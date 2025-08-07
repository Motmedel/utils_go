package interfaces

type Signer interface {
	Sign(message []byte) (signature []byte, err error)
}

type Verifier interface {
	Verify(message []byte, signature []byte) error
}

type Method interface {
	Signer
	Verifier
	GetName() string
}