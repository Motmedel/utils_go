package mux

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func vhostRequest(t *testing.T, host string) *http.Request {
	t.Helper()
	request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/path", nil)
	request.Host = host
	return request
}

func TestVhostMuxHandleRequest(t *testing.T) {
	t.Parallel()

	t.Run("nil vhost mux", func(t *testing.T) {
		t.Parallel()
		_, responseError := vhostMuxHandleRequest(nil, vhostRequest(t, "a.example.com"), httptest.NewRecorder())
		if responseError == nil || responseError.ServerError == nil {
			t.Fatalf("expected a server error, got %#v", responseError)
		}
	})

	t.Run("nil request", func(t *testing.T) {
		t.Parallel()
		_, responseError := vhostMuxHandleRequest(&VhostMux{}, nil, httptest.NewRecorder())
		if responseError == nil || responseError.ServerError == nil {
			t.Fatalf("expected a server error, got %#v", responseError)
		}
	})

	t.Run("nil host map", func(t *testing.T) {
		t.Parallel()
		_, responseError := vhostMuxHandleRequest(&VhostMux{}, vhostRequest(t, "a.example.com"), httptest.NewRecorder())
		if responseError == nil || responseError.ServerError == nil {
			t.Fatalf("expected a server error, got %#v", responseError)
		}
	})

	t.Run("unknown host is 421", func(t *testing.T) {
		t.Parallel()
		vhostMux := &VhostMux{HostToSpecification: map[string]*VhostMuxSpecification{}}
		_, responseError := vhostMuxHandleRequest(vhostMux, vhostRequest(t, "unknown.example.com"), httptest.NewRecorder())
		if responseError == nil || responseError.ProblemDetail == nil || responseError.ProblemDetail.Status != http.StatusMisdirectedRequest {
			t.Fatalf("expected 421, got %#v", responseError)
		}
	})

	t.Run("nil specification", func(t *testing.T) {
		t.Parallel()
		vhostMux := &VhostMux{HostToSpecification: map[string]*VhostMuxSpecification{"a.example.com": nil}}
		_, responseError := vhostMuxHandleRequest(vhostMux, vhostRequest(t, "a.example.com"), httptest.NewRecorder())
		if responseError == nil || responseError.ServerError == nil {
			t.Fatalf("expected a server error, got %#v", responseError)
		}
	})

	t.Run("redirect", func(t *testing.T) {
		t.Parallel()
		vhostMux := &VhostMux{HostToSpecification: map[string]*VhostMuxSpecification{
			"r.example.com": {RedirectTo: "https://new.example.com"},
		}}
		response, responseError := vhostMuxHandleRequest(vhostMux, vhostRequest(t, "r.example.com"), httptest.NewRecorder())
		if responseError != nil {
			t.Fatalf("unexpected error: %#v", responseError)
		}
		if response == nil || response.StatusCode != http.StatusMovedPermanently {
			t.Fatalf("expected a 301, got %#v", response)
		}
		var location string
		for _, header := range response.Headers {
			if header != nil && header.Name == "Location" {
				location = header.Value
			}
		}
		if !strings.HasPrefix(location, "https://new.example.com") {
			t.Fatalf("unexpected Location: %q", location)
		}
	})

	t.Run("delegates to the inner mux", func(t *testing.T) {
		t.Parallel()
		handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})
		vhostMux := &VhostMux{HostToSpecification: map[string]*VhostMuxSpecification{
			"m.example.com": {Mux: handler},
		}}
		recorder := httptest.NewRecorder()
		response, responseError := vhostMuxHandleRequest(vhostMux, vhostRequest(t, "m.example.com"), recorder)
		if responseError != nil || response != nil {
			t.Fatalf("expected the inner mux to handle it, got response %#v error %#v", response, responseError)
		}
		if recorder.Code != http.StatusTeapot {
			t.Fatalf("expected the inner mux to write 418, got %d", recorder.Code)
		}
	})

	t.Run("unusable specification", func(t *testing.T) {
		t.Parallel()
		vhostMux := &VhostMux{HostToSpecification: map[string]*VhostMuxSpecification{"u.example.com": {}}}
		_, responseError := vhostMuxHandleRequest(vhostMux, vhostRequest(t, "u.example.com"), httptest.NewRecorder())
		if responseError == nil || responseError.ServerError == nil {
			t.Fatalf("expected a server error, got %#v", responseError)
		}
	})
}

func TestVhostMux_ServeHTTP(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("inner"))
	})
	vhostMux := &VhostMux{HostToSpecification: map[string]*VhostMuxSpecification{
		"app.example.com": {Mux: handler},
	}}

	recorder := httptest.NewRecorder()
	vhostMux.ServeHTTP(recorder, vhostRequest(t, "app.example.com"))

	if recorder.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200", recorder.Code)
	}
	if recorder.Body.String() != "inner" {
		t.Fatalf("body = %q, want inner", recorder.Body.String())
	}
}

func TestVhostMux_PatchHttpServer(t *testing.T) {
	t.Parallel()

	t.Run("nil server is a no-op", func(t *testing.T) {
		t.Parallel()
		(&VhostMux{}).PatchHttpServer(nil)
	})

	t.Run("patches handler and certificate resolver", func(t *testing.T) {
		t.Parallel()

		certificate := &tls.Certificate{}
		vhostMux := &VhostMux{HostToSpecification: map[string]*VhostMuxSpecification{
			"secure.example.com": {Certificate: certificate},
		}}
		server := &http.Server{ReadHeaderTimeout: time.Second}
		vhostMux.PatchHttpServer(server)

		if server.Handler == nil {
			t.Error("expected the server handler to be set")
		}
		if server.TLSConfig == nil || server.TLSConfig.GetCertificate == nil {
			t.Fatal("expected GetCertificate to be set")
		}

		got, err := server.TLSConfig.GetCertificate(&tls.ClientHelloInfo{ServerName: "secure.example.com"})
		if err != nil {
			t.Fatalf("get certificate: %v", err)
		}
		if got != certificate {
			t.Error("expected the configured certificate")
		}

		if got, err := server.TLSConfig.GetCertificate(nil); err != nil || got != nil {
			t.Errorf("nil client hello: got %#v err %v", got, err)
		}

		if got, err := server.TLSConfig.GetCertificate(&tls.ClientHelloInfo{ServerName: "unknown.example.com"}); err != nil || got != nil {
			t.Errorf("unknown host: got %#v err %v", got, err)
		}
	})

	t.Run("nil host map yields a certificate error", func(t *testing.T) {
		t.Parallel()
		server := &http.Server{ReadHeaderTimeout: time.Second}
		(&VhostMux{}).PatchHttpServer(server)
		if _, err := server.TLSConfig.GetCertificate(&tls.ClientHelloInfo{ServerName: "x"}); err == nil {
			t.Fatal("expected an error for a nil host map")
		}
	})
}
