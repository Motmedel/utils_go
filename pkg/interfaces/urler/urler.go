package urler

import (
	"errors"
	"net/url"
)

var (
	ErrNilUrler       = errors.New("nil urler")
	ErrNilStringUrler = errors.New("nil string urler")
)

type URLer interface {
	URL() *url.URL
}

type StringURLer interface {
	URL() string
}
