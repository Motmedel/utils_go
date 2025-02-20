package firewall

import (
	"github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	"net/http"
)

type Verdict int

const (
	VerdictAccept Verdict = iota
	VerdictDrop
	VerdictReject
)

type Configuration struct {
	Handler func(*http.Request) (Verdict, *response_error.ResponseError)
}
