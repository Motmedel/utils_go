package json_body_parser

import (
	"net/http"
	"testing"
)

type payload struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestParse(t *testing.T) {
	t.Parallel()

	parser := New[payload]()

	t.Run("valid body", func(t *testing.T) {
		t.Parallel()
		result, responseError := parser.Parse(nil, []byte(`{"name":"alice","age":30}`))
		if responseError != nil {
			t.Fatalf("unexpected error: %#v", responseError)
		}
		if result.Name != "alice" || result.Age != 30 {
			t.Fatalf("unexpected result: %#v", result)
		}
	})

	t.Run("type mismatch is 422", func(t *testing.T) {
		t.Parallel()
		_, responseError := parser.Parse(nil, []byte(`{"age":"not-a-number"}`))
		if responseError == nil || responseError.ProblemDetail == nil {
			t.Fatalf("expected a problem detail, got %#v", responseError)
		}
		if responseError.ProblemDetail.Status != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", responseError.ProblemDetail.Status)
		}
	})

	t.Run("malformed body is a 400 client error", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			name string
			body []byte
		}{
			{name: "garbage", body: []byte(`{malformed`)},
			{name: "truncated", body: []byte(`{"name":"alice"`)},
			{name: "empty", body: []byte(``)},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				t.Parallel()
				_, responseError := parser.Parse(nil, testCase.body)
				if responseError == nil || responseError.ProblemDetail == nil {
					t.Fatalf("expected a problem detail, got %#v", responseError)
				}
				if responseError.ClientError == nil {
					t.Fatalf("expected a client error, got %#v", responseError)
				}
				if responseError.ServerError != nil {
					t.Fatalf("expected no server error, got %#v", responseError.ServerError)
				}
				if responseError.ProblemDetail.Status != http.StatusBadRequest {
					t.Fatalf("expected 400, got %d", responseError.ProblemDetail.Status)
				}
			})
		}
	})
}
