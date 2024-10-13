package types

type WhoisContext struct {
	ServerAddress string
	ServerPort    int
	ClientAddress string
	ClientPort    int
	Transport     string
	RequestData   []byte
	ResponseData  []byte
}
