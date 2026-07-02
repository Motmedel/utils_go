package urler

import (
	"net/url"
)

type URLer interface {
	URL() *url.URL
}

type StringURLer interface {
	URL() string
}
