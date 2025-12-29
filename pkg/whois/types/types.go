package types

import "time"

type WhoisContext struct {
	Time            *time.Time
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
