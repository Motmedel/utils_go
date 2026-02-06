package body_parser

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	muxErrors "github.com/Motmedel/utils_go/pkg/http/mux/errors"
	processorPkg "github.com/Motmedel/utils_go/pkg/http/mux/types/processor"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	"github.com/Motmedel/utils_go/pkg/http/types/problem_detail"
)

func TestBodyParserFunction(t *testing.T) {
	t.Run("returns value from function", func(t *testing.T) {
		expected := "test-value"
		parser := BodyParserFunction[string](func(r *http.Request, body []byte) (string, *response_error.ResponseError) {
			return expected, nil
		})

		request := httptest.NewRequest(http.MethodPost, "/test", nil)
		result, responseError := parser.Parse(request, []byte("body"))

		if responseError != nil {
			t.Fatalf("unexpected response error: %v", responseError)
		}

		if result != expected {
			t.Errorf("got %q, expected %q", result, expected)
		}
	})

	t.Run("returns error from function", func(t *testing.T) {
		expectedError := errors.New("test error")
		parser := BodyParserFunction[string](func(r *http.Request, body []byte) (string, *response_error.ResponseError) {
			return "", &response_error.ResponseError{ServerError: expectedError}
		})

		request := httptest.NewRequest(http.MethodPost, "/test", nil)
		_, responseError := parser.Parse(request, []byte("body"))

		if responseError == nil {
			t.Fatal("expected response error, got nil")
		}

		if !errors.Is(responseError.ServerError, expectedError) {
			t.Errorf("got error %v, expected %v", responseError.ServerError, expectedError)
		}
	})

	t.Run("receives request and body", func(t *testing.T) {
		expectedBody := []byte("test body content")
		var receivedBody []byte
		var receivedRequest *http.Request

		parser := BodyParserFunction[string](func(r *http.Request, body []byte) (string, *response_error.ResponseError) {
			receivedRequest = r
			receivedBody = body
			return "ok", nil
		})

		request := httptest.NewRequest(http.MethodPost, "/test", nil)
		_, responseError := parser.Parse(request, expectedBody)

		if responseError != nil {
			t.Fatalf("unexpected response error: %v", responseError)
		}

		if receivedRequest == nil {
			t.Fatal("expected request to be passed to function")
		}

		if string(receivedBody) != string(expectedBody) {
			t.Errorf("got body %q, expected %q", receivedBody, expectedBody)
		}
	})
}

func TestNew(t *testing.T) {
	t.Run("creates parser from function", func(t *testing.T) {
		expected := 42
		parser := New(func(r *http.Request, body []byte) (int, *response_error.ResponseError) {
			return expected, nil
		})

		request := httptest.NewRequest(http.MethodPost, "/test", nil)
		result, responseError := parser.Parse(request, []byte("body"))

		if responseError != nil {
			t.Fatalf("unexpected response error: %v", responseError)
		}

		if result != expected {
			t.Errorf("got %d, expected %d", result, expected)
		}
	})
}

func TestBodyParserWithProcessor(t *testing.T) {
	t.Run("nil body parser returns error", func(t *testing.T) {
		parser := &BodyParserWithProcessor[string, int]{
			BodyParser: nil,
			Processor: processorPkg.New(func(ctx context.Context, input string) (int, *response_error.ResponseError) {
				return len(input), nil
			}),
		}

		request := httptest.NewRequest(http.MethodPost, "/test", nil)
		_, responseError := parser.Parse(request, []byte("body"))

		if responseError == nil {
			t.Fatal("expected response error, got nil")
		}

		if responseError.ServerError == nil {
			t.Fatal("expected server error, got nil")
		}

		if !errors.Is(responseError.ServerError, muxErrors.ErrNilBodyParser) {
			t.Errorf("got error %v, expected %v", responseError.ServerError, muxErrors.ErrNilBodyParser)
		}
	})

	t.Run("nil processor returns error", func(t *testing.T) {
		parser := &BodyParserWithProcessor[string, int]{
			BodyParser: New(func(r *http.Request, body []byte) (string, *response_error.ResponseError) {
				return string(body), nil
			}),
			Processor: nil,
		}

		request := httptest.NewRequest(http.MethodPost, "/test", nil)
		_, responseError := parser.Parse(request, []byte("body"))

		if responseError == nil {
			t.Fatal("expected response error, got nil")
		}

		if responseError.ServerError == nil {
			t.Fatal("expected server error, got nil")
		}

		if !errors.Is(responseError.ServerError, muxErrors.ErrNilProcessor) {
			t.Errorf("got error %v, expected %v", responseError.ServerError, muxErrors.ErrNilProcessor)
		}
	})

	t.Run("body parser error is propagated", func(t *testing.T) {
		expectedDetail := "parser error"
		parser := &BodyParserWithProcessor[string, int]{
			BodyParser: New(func(r *http.Request, body []byte) (string, *response_error.ResponseError) {
				return "", &response_error.ResponseError{
					ProblemDetail: problem_detail.New(http.StatusBadRequest),
					ClientError:   errors.New(expectedDetail),
				}
			}),
			Processor: processorPkg.New(func(ctx context.Context, input string) (int, *response_error.ResponseError) {
				return len(input), nil
			}),
		}

		request := httptest.NewRequest(http.MethodPost, "/test", nil)
		_, responseError := parser.Parse(request, []byte("body"))

		if responseError == nil {
			t.Fatal("expected response error, got nil")
		}

		if responseError.ClientError == nil {
			t.Fatal("expected client error, got nil")
		}

		if responseError.ClientError.Error() != expectedDetail {
			t.Errorf("got error %q, expected %q", responseError.ClientError.Error(), expectedDetail)
		}
	})

	t.Run("processor error is propagated", func(t *testing.T) {
		expectedDetail := "processor error"
		parser := &BodyParserWithProcessor[string, int]{
			BodyParser: New(func(r *http.Request, body []byte) (string, *response_error.ResponseError) {
				return string(body), nil
			}),
			Processor: processorPkg.New(func(ctx context.Context, input string) (int, *response_error.ResponseError) {
				return 0, &response_error.ResponseError{
					ProblemDetail: problem_detail.New(http.StatusUnprocessableEntity),
					ClientError:   errors.New(expectedDetail),
				}
			}),
		}

		request := httptest.NewRequest(http.MethodPost, "/test", nil)
		_, responseError := parser.Parse(request, []byte("body"))

		if responseError == nil {
			t.Fatal("expected response error, got nil")
		}

		if responseError.ClientError == nil {
			t.Fatal("expected client error, got nil")
		}

		if responseError.ClientError.Error() != expectedDetail {
			t.Errorf("got error %q, expected %q", responseError.ClientError.Error(), expectedDetail)
		}
	})

	t.Run("successful parsing and processing", func(t *testing.T) {
		parser := &BodyParserWithProcessor[string, int]{
			BodyParser: New(func(r *http.Request, body []byte) (string, *response_error.ResponseError) {
				return string(body), nil
			}),
			Processor: processorPkg.New(func(ctx context.Context, input string) (int, *response_error.ResponseError) {
				return len(input), nil
			}),
		}

		request := httptest.NewRequest(http.MethodPost, "/test", nil)
		result, responseError := parser.Parse(request, []byte("hello"))

		if responseError != nil {
			t.Fatalf("unexpected response error: %v", responseError)
		}

		if result != 5 {
			t.Errorf("got %d, expected 5", result)
		}
	})

	t.Run("processor receives parsed value", func(t *testing.T) {
		expectedValue := "parsed-body-content"
		var receivedValue string

		parser := &BodyParserWithProcessor[string, string]{
			BodyParser: New(func(r *http.Request, body []byte) (string, *response_error.ResponseError) {
				return expectedValue, nil
			}),
			Processor: processorPkg.New(func(ctx context.Context, input string) (string, *response_error.ResponseError) {
				receivedValue = input
				return input + "-processed", nil
			}),
		}

		request := httptest.NewRequest(http.MethodPost, "/test", nil)
		result, responseError := parser.Parse(request, []byte("body"))

		if responseError != nil {
			t.Fatalf("unexpected response error: %v", responseError)
		}

		if receivedValue != expectedValue {
			t.Errorf("processor received %q, expected %q", receivedValue, expectedValue)
		}

		if result != expectedValue+"-processed" {
			t.Errorf("got %q, expected %q", result, expectedValue+"-processed")
		}
	})

	t.Run("context is passed to processor", func(t *testing.T) {
		type contextKey string
		key := contextKey("test-key")
		expectedContextValue := "context-value"

		parser := &BodyParserWithProcessor[string, string]{
			BodyParser: New(func(r *http.Request, body []byte) (string, *response_error.ResponseError) {
				return string(body), nil
			}),
			Processor: processorPkg.New(func(ctx context.Context, input string) (string, *response_error.ResponseError) {
				value := ctx.Value(key)
				if value == nil {
					return "", &response_error.ResponseError{ServerError: errors.New("context value not found")}
				}
				return value.(string), nil
			}),
		}

		request := httptest.NewRequest(http.MethodPost, "/test", nil)
		request = request.WithContext(context.WithValue(request.Context(), key, expectedContextValue))
		result, responseError := parser.Parse(request, []byte("body"))

		if responseError != nil {
			t.Fatalf("unexpected response error: %v", responseError)
		}

		if result != expectedContextValue {
			t.Errorf("got %q, expected %q", result, expectedContextValue)
		}
	})

	t.Run("body is passed to body parser", func(t *testing.T) {
		expectedBody := []byte("expected-body-content")
		var receivedBody []byte

		parser := &BodyParserWithProcessor[string, string]{
			BodyParser: New(func(r *http.Request, body []byte) (string, *response_error.ResponseError) {
				receivedBody = body
				return string(body), nil
			}),
			Processor: processorPkg.New(func(ctx context.Context, input string) (string, *response_error.ResponseError) {
				return input, nil
			}),
		}

		request := httptest.NewRequest(http.MethodPost, "/test", nil)
		_, responseError := parser.Parse(request, expectedBody)

		if responseError != nil {
			t.Fatalf("unexpected response error: %v", responseError)
		}

		if string(receivedBody) != string(expectedBody) {
			t.Errorf("got body %q, expected %q", receivedBody, expectedBody)
		}
	})
}

func TestNewWithProcessor(t *testing.T) {
	t.Run("creates parser with processor", func(t *testing.T) {
		bodyParser := New(func(r *http.Request, body []byte) (string, *response_error.ResponseError) {
			return string(body), nil
		})
		processor := processorPkg.New(func(ctx context.Context, input string) (int, *response_error.ResponseError) {
			return len(input), nil
		})

		parser := NewWithProcessor(bodyParser, processor)

		if parser == nil {
			t.Fatal("expected non-nil parser")
		}

		request := httptest.NewRequest(http.MethodPost, "/test", nil)
		result, responseError := parser.Parse(request, []byte("test"))

		if responseError != nil {
			t.Fatalf("unexpected response error: %v", responseError)
		}

		if result != 4 {
			t.Errorf("got %d, expected 4", result)
		}
	})

	t.Run("request is available in body parser", func(t *testing.T) {
		var receivedMethod string

		bodyParser := New(func(r *http.Request, body []byte) (string, *response_error.ResponseError) {
			receivedMethod = r.Method
			return string(body), nil
		})
		processor := processorPkg.New(func(ctx context.Context, input string) (string, *response_error.ResponseError) {
			return input, nil
		})

		parser := NewWithProcessor(bodyParser, processor)

		request := httptest.NewRequest(http.MethodPut, "/test", nil)
		_, responseError := parser.Parse(request, []byte("body"))

		if responseError != nil {
			t.Fatalf("unexpected response error: %v", responseError)
		}

		if receivedMethod != http.MethodPut {
			t.Errorf("got method %q, expected %q", receivedMethod, http.MethodPut)
		}
	})
}
