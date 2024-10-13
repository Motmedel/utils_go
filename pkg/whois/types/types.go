package types

type WhoisContext struct {
	ServerAddress   string
	ServerIpAddress string
	ServerPort      int
	ClientAddress   string
	ClientIpAddress string
	ClientPort      int
	Transport       string
	RequestData     []byte
	ResponseData    []byte
}
