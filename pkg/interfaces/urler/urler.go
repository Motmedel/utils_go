package urler

import (
	"errors"
	"net/url"
)

var ErrNilUrler = errors.New("nil urler")

type URLer interface {
	URL() *url.URL
}
