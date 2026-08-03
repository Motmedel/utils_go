package middleware

import (
	stdcontext "context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddleware_IdentityPassesRequestThrough(t *testing.T) {
	t.Parallel()

	var identity Middleware = func(request *http.Request) *http.Request {
		return request
	}

	request := httptest.NewRequestWithContext(stdcontext.Background(), http.MethodGet, "/path", nil)
	if got := identity(request); got != request {
		t.Errorf("identity middleware returned a different request pointer")
	}
}

func TestMiddleware_TransformsRequest(t *testing.T) {
	t.Parallel()

	type keyType struct{}
	key := &keyType{}

	var enrich Middleware = func(request *http.Request) *http.Request {
		return request.WithContext(stdcontext.WithValue(request.Context(), key, "value"))
	}

	request := httptest.NewRequestWithContext(stdcontext.Background(), http.MethodGet, "/path", nil)
	got := enrich(request)

	if got.Context().Value(key) != "value" {
		t.Error("middleware did not attach the expected context value")
	}
}

func TestMiddleware_Composition(t *testing.T) {
	t.Parallel()

	var addHeaderA Middleware = func(request *http.Request) *http.Request {
		request.Header.Set("X-A", "a")
		return request
	}
	var addHeaderB Middleware = func(request *http.Request) *http.Request {
		request.Header.Set("X-B", "b")
		return request
	}

	request := httptest.NewRequestWithContext(stdcontext.Background(), http.MethodGet, "/path", nil)
	request = addHeaderB(addHeaderA(request))

	if request.Header.Get("X-A") != "a" {
		t.Error("first middleware header missing")
	}
	if request.Header.Get("X-B") != "b" {
		t.Error("second middleware header missing")
	}
}
