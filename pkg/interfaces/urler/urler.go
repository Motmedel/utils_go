package urler

import "net/url"

type URLer interface {
	URL() *url.URL
}
