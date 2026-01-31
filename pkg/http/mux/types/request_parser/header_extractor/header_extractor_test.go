package header_extractor

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	motmedelHttpErrors "github.com/Motmedel/utils_go/pkg/http/errors"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/request_parser/header_extractor/header_extractor_config"
)

func TestNew(t *testing.T) {
	t.Run("empty name returns error", func(t *testing.T) {
		_, err := New("")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(err, ErrEmptyName) {
			t.Errorf("got error %v, expected %v", err, ErrEmptyName)
		}
	})

	t.Run("valid name creates parser", func(t *testing.T) {
		parser, err := New("X-Custom-Header")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if parser == nil {
			t.Fatal("expected non-nil parser")
		}

		if parser.Name != "X-Custom-Header" {
			t.Errorf("got name %q, expected %q", parser.Name, "X-Custom-Header")
		}
	})

	t.Run("with custom options", func(t *testing.T) {
		parser, err := New(
			"Authorization",
			header_extractor_config.WithProblemDetailStatusCode(http.StatusUnauthorized),
			header_extractor_config.WithProblemDetailMissingText("Missing authorization header."),
			header_extractor_config.WithProblemDetailMultipleText("Multiple authorization headers."),
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if parser == nil {
			t.Fatal("expected non-nil parser")
		}
	})
}

func TestParser_Parse(t *testing.T) {
	t.Run("nil request returns error", func(t *testing.T) {
		parser, err := New("X-Custom-Header")
		if err != nil {
			t.Fatalf("unexpected error creating parser: %v", err)
		}

		_, responseError := parser.Parse(nil)
		if responseError == nil {
			t.Fatal("expected response error, got nil")
		}

		if responseError.ServerError == nil {
			t.Fatal("expected server error, got nil")
		}

		if !errors.Is(responseError.ServerError, motmedelHttpErrors.ErrNilHttpRequest) {
			t.Errorf("got error %v, expected %v", responseError.ServerError, motmedelHttpErrors.ErrNilHttpRequest)
		}
	})

	t.Run("nil header returns error", func(t *testing.T) {
		parser, err := New("X-Custom-Header")
		if err != nil {
			t.Fatalf("unexpected error creating parser: %v", err)
		}

		request := &http.Request{Header: nil}
		_, responseError := parser.Parse(request)

		if responseError == nil {
			t.Fatal("expected response error, got nil")
		}

		if responseError.ServerError == nil {
			t.Fatal("expected server error, got nil")
		}

		if !errors.Is(responseError.ServerError, motmedelHttpErrors.ErrNilHttpRequestHeader) {
			t.Errorf("got error %v, expected %v", responseError.ServerError, motmedelHttpErrors.ErrNilHttpRequestHeader)
		}
	})

	t.Run("empty name in parser returns error", func(t *testing.T) {
		parser := &Parser{Name: ""}

		request := httptest.NewRequest(http.MethodGet, "/test", nil)
		_, responseError := parser.Parse(request)

		if responseError == nil {
			t.Fatal("expected response error, got nil")
		}

		if responseError.ServerError == nil {
			t.Fatal("expected server error, got nil")
		}

		if !errors.Is(responseError.ServerError, ErrEmptyName) {
			t.Errorf("got error %v, expected %v", responseError.ServerError, ErrEmptyName)
		}
	})

	t.Run("missing header returns client error", func(t *testing.T) {
		parser, err := New("X-Custom-Header")
		if err != nil {
			t.Fatalf("unexpected error creating parser: %v", err)
		}

		request := httptest.NewRequest(http.MethodGet, "/test", nil)
		_, responseError := parser.Parse(request)

		if responseError == nil {
			t.Fatal("expected response error, got nil")
		}

		if responseError.ClientError == nil {
			t.Fatal("expected client error, got nil")
		}

		if !errors.Is(responseError.ClientError, motmedelHttpErrors.ErrMissingHeader) {
			t.Errorf("got error %v, expected %v", responseError.ClientError, motmedelHttpErrors.ErrMissingHeader)
		}

		if responseError.ProblemDetail == nil {
			t.Fatal("expected problem detail, got nil")
		}

		expectedStatusCode := http.StatusBadRequest
		if responseError.ProblemDetail.Status != expectedStatusCode {
			t.Errorf("got status %d, expected %d", responseError.ProblemDetail.Status, expectedStatusCode)
		}

		extension := responseError.ProblemDetail.Extension
		if extension == nil {
			t.Fatal("expected problem detail extension, got nil")
		}

		headerValue, ok := extension["header"]
		if !ok {
			t.Fatal("expected 'header' key in extension")
		}

		if headerValue != "X-Custom-Header" {
			t.Errorf("got header %q, expected %q", headerValue, "X-Custom-Header")
		}
	})

	t.Run("multiple header values returns client error", func(t *testing.T) {
		parser, err := New("X-Custom-Header")
		if err != nil {
			t.Fatalf("unexpected error creating parser: %v", err)
		}

		request := httptest.NewRequest(http.MethodGet, "/test", nil)
		request.Header.Add("X-Custom-Header", "value1")
		request.Header.Add("X-Custom-Header", "value2")
		_, responseError := parser.Parse(request)

		if responseError == nil {
			t.Fatal("expected response error, got nil")
		}

		if responseError.ClientError == nil {
			t.Fatal("expected client error, got nil")
		}

		if !errors.Is(responseError.ClientError, motmedelHttpErrors.ErrMultipleHeaderValues) {
			t.Errorf("got error %v, expected %v", responseError.ClientError, motmedelHttpErrors.ErrMultipleHeaderValues)
		}

		if responseError.ProblemDetail == nil {
			t.Fatal("expected problem detail, got nil")
		}

		expectedStatusCode := http.StatusBadRequest
		if responseError.ProblemDetail.Status != expectedStatusCode {
			t.Errorf("got status %d, expected %d", responseError.ProblemDetail.Status, expectedStatusCode)
		}
	})

	t.Run("single header value returns value", func(t *testing.T) {
		parser, err := New("X-Custom-Header")
		if err != nil {
			t.Fatalf("unexpected error creating parser: %v", err)
		}

		expectedValue := "test-value"
		request := httptest.NewRequest(http.MethodGet, "/test", nil)
		request.Header.Set("X-Custom-Header", expectedValue)
		result, responseError := parser.Parse(request)

		if responseError != nil {
			t.Fatalf("unexpected response error: %v", responseError)
		}

		if result != expectedValue {
			t.Errorf("got %q, expected %q", result, expectedValue)
		}
	})

	t.Run("canonical header name matching", func(t *testing.T) {
		parser, err := New("content-type")
		if err != nil {
			t.Fatalf("unexpected error creating parser: %v", err)
		}

		expectedValue := "application/json"
		request := httptest.NewRequest(http.MethodGet, "/test", nil)
		request.Header.Set("Content-Type", expectedValue)
		result, responseError := parser.Parse(request)

		if responseError != nil {
			t.Fatalf("unexpected response error: %v", responseError)
		}

		if result != expectedValue {
			t.Errorf("got %q, expected %q", result, expectedValue)
		}
	})

	t.Run("custom status code for missing header", func(t *testing.T) {
		parser, err := New(
			"Authorization",
			header_extractor_config.WithProblemDetailStatusCode(http.StatusUnauthorized),
		)
		if err != nil {
			t.Fatalf("unexpected error creating parser: %v", err)
		}

		request := httptest.NewRequest(http.MethodGet, "/test", nil)
		_, responseError := parser.Parse(request)

		if responseError == nil {
			t.Fatal("expected response error, got nil")
		}

		if responseError.ProblemDetail == nil {
			t.Fatal("expected problem detail, got nil")
		}

		expectedStatusCode := http.StatusUnauthorized
		if responseError.ProblemDetail.Status != expectedStatusCode {
			t.Errorf("got status %d, expected %d", responseError.ProblemDetail.Status, expectedStatusCode)
		}
	})

	t.Run("custom text for missing header", func(t *testing.T) {
		expectedText := "Custom missing header message."
		parser, err := New(
			"X-Custom-Header",
			header_extractor_config.WithProblemDetailMissingText(expectedText),
		)
		if err != nil {
			t.Fatalf("unexpected error creating parser: %v", err)
		}

		request := httptest.NewRequest(http.MethodGet, "/test", nil)
		_, responseError := parser.Parse(request)

		if responseError == nil {
			t.Fatal("expected response error, got nil")
		}

		if responseError.ProblemDetail == nil {
			t.Fatal("expected problem detail, got nil")
		}

		if responseError.ProblemDetail.Detail != expectedText {
			t.Errorf("got detail %q, expected %q", responseError.ProblemDetail.Detail, expectedText)
		}
	})

	t.Run("custom text for multiple header values", func(t *testing.T) {
		expectedText := "Custom multiple headers message."
		parser, err := New(
			"X-Custom-Header",
			header_extractor_config.WithProblemDetailMultipleText(expectedText),
		)
		if err != nil {
			t.Fatalf("unexpected error creating parser: %v", err)
		}

		request := httptest.NewRequest(http.MethodGet, "/test", nil)
		request.Header.Add("X-Custom-Header", "value1")
		request.Header.Add("X-Custom-Header", "value2")
		_, responseError := parser.Parse(request)

		if responseError == nil {
			t.Fatal("expected response error, got nil")
		}

		if responseError.ProblemDetail == nil {
			t.Fatal("expected problem detail, got nil")
		}

		if responseError.ProblemDetail.Detail != expectedText {
			t.Errorf("got detail %q, expected %q", responseError.ProblemDetail.Detail, expectedText)
		}
	})

	t.Run("empty header value is valid", func(t *testing.T) {
		parser, err := New("X-Custom-Header")
		if err != nil {
			t.Fatalf("unexpected error creating parser: %v", err)
		}

		request := httptest.NewRequest(http.MethodGet, "/test", nil)
		request.Header.Set("X-Custom-Header", "")
		result, responseError := parser.Parse(request)

		if responseError != nil {
			t.Fatalf("unexpected response error: %v", responseError)
		}

		if result != "" {
			t.Errorf("got %q, expected empty string", result)
		}
	})
}
