package json_schema_body_parser

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Motmedel/utils_go/pkg/errors/types/nil_error"
	jsonschemaErrors "github.com/altshiftab/jsonschema/pkg/errors"
	"github.com/altshiftab/jsonschema/pkg/jsonschema"
)

type payload struct {
	Name string `json:"name"`
}

func newRequest(t *testing.T) *http.Request {
	t.Helper()
	return httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", nil)
}

func TestNew(t *testing.T) {
	t.Parallel()

	parser, err := New[payload]()
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	if parser == nil {
		t.Fatal("expected a non-nil parser")
	}
	if parser.schema == nil {
		t.Error("expected a non-nil schema")
	}
}

func TestNewWithSchema(t *testing.T) {
	t.Parallel()

	schema, err := jsonschema.NewFromType[payload]()
	if err != nil {
		t.Fatalf("new from type: %v", err)
	}

	testCases := []struct {
		name        string
		schema      *jsonschema.Schema
		expectError bool
	}{
		{name: "nil schema", schema: nil, expectError: true},
		{name: "valid schema", schema: schema, expectError: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			parser, err := NewWithSchema[payload](testCase.schema)
			if testCase.expectError {
				if err == nil {
					t.Fatal("expected an error")
				}
				if !errors.Is(err, jsonschemaErrors.ErrNilSchema) {
					t.Errorf("error = %v, want ErrNilSchema", err)
				}
				if parser != nil {
					t.Error("expected a nil parser")
				}
				return
			}

			if err != nil {
				t.Fatalf("new with schema: %v", err)
			}
			if parser == nil {
				t.Fatal("expected a non-nil parser")
			}
			if parser.schema != testCase.schema {
				t.Error("expected the provided schema to be stored")
			}
		})
	}
}

func TestParserParse(t *testing.T) {
	t.Parallel()

	parser, err := New[payload]()
	if err != nil {
		t.Fatalf("new: %v", err)
	}

	testCases := []struct {
		name                string
		body                string
		expectResponseError bool
		expectClientError   bool
		expectServerError   bool
		expectValidateError bool
		expectStatus        int
		expectName          string
	}{
		{
			name:       "valid body",
			body:       `{"name":"hello"}`,
			expectName: "hello",
		},
		{
			name:                "schema violation type mismatch",
			body:                `{"name":123}`,
			expectResponseError: true,
			expectClientError:   true,
			expectValidateError: true,
			expectStatus:        http.StatusUnprocessableEntity,
		},
		{
			name:                "schema violation missing required",
			body:                `{}`,
			expectResponseError: true,
			expectClientError:   true,
			expectValidateError: true,
			expectStatus:        http.StatusUnprocessableEntity,
		},
		{
			name:                "malformed json",
			body:                `{`,
			expectResponseError: true,
			expectServerError:   true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result, responseError := parser.Parse(newRequest(t), []byte(testCase.body))

			if !testCase.expectResponseError {
				if responseError != nil {
					t.Fatalf("unexpected response error: %+v", responseError)
				}
				if result.Name != testCase.expectName {
					t.Errorf("result name = %q, want %q", result.Name, testCase.expectName)
				}
				return
			}

			if responseError == nil {
				t.Fatal("expected a response error")
			}

			if testCase.expectClientError {
				if responseError.ClientError == nil {
					t.Error("expected a client error")
				}
				if responseError.ServerError != nil {
					t.Errorf("unexpected server error: %v", responseError.ServerError)
				}
			}

			if testCase.expectServerError {
				if responseError.ServerError == nil {
					t.Error("expected a server error")
				}
				if responseError.ClientError != nil {
					t.Errorf("unexpected client error: %v", responseError.ClientError)
				}
			}

			if testCase.expectValidateError {
				if _, ok := errors.AsType[*jsonschemaErrors.ValidateError](responseError.ClientError); !ok {
					t.Errorf("client error is not a *ValidateError: %v", responseError.ClientError)
				}
			}

			if testCase.expectStatus != 0 {
				problemDetail := responseError.ProblemDetail
				if problemDetail == nil {
					t.Fatal("expected a problem detail")
				}
				if problemDetail.Status != testCase.expectStatus {
					t.Errorf("status = %d, want %d", problemDetail.Status, testCase.expectStatus)
				}
			}
		})
	}
}

func TestParserParseSlice(t *testing.T) {
	t.Parallel()

	parser, err := New[[]payload]()
	if err != nil {
		t.Fatalf("new: %v", err)
	}

	result, responseError := parser.Parse(newRequest(t), []byte(`[{"name":"a"},{"name":"b"}]`))
	if responseError != nil {
		t.Fatalf("unexpected response error: %+v", responseError)
	}
	if len(result) != 2 {
		t.Fatalf("result length = %d, want 2", len(result))
	}
	if result[0].Name != "a" || result[1].Name != "b" {
		t.Errorf("result = %+v, want [{a} {b}]", result)
	}

	_, responseError = parser.Parse(newRequest(t), []byte(`[{"name":123}]`))
	if responseError == nil {
		t.Fatal("expected a response error for a schema violation")
	}
	if _, ok := errors.AsType[*jsonschemaErrors.ValidateError](responseError.ClientError); !ok {
		t.Errorf("client error is not a *ValidateError: %v", responseError.ClientError)
	}
	if problemDetail := responseError.ProblemDetail; problemDetail == nil {
		t.Fatal("expected a problem detail")
	} else if problemDetail.Status != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want %d", problemDetail.Status, http.StatusUnprocessableEntity)
	}
}

func TestParserParseNilSchema(t *testing.T) {
	t.Parallel()

	parser := &Parser[payload]{}

	_, responseError := parser.Parse(newRequest(t), []byte(`{"name":"hello"}`))
	if responseError == nil {
		t.Fatal("expected a response error")
	}
	if responseError.ServerError == nil {
		t.Fatal("expected a server error")
	}
	if !errors.Is(responseError.ServerError, jsonschemaErrors.ErrNilSchema) {
		t.Errorf("server error = %v, want ErrNilSchema", responseError.ServerError)
	}
}

func TestParserParseNilBodyParser(t *testing.T) {
	t.Parallel()

	schema, err := jsonschema.NewFromType[payload]()
	if err != nil {
		t.Fatalf("new from type: %v", err)
	}

	parser := &Parser[payload]{schema: schema}

	_, responseError := parser.Parse(newRequest(t), []byte(`{"name":"hello"}`))
	if responseError == nil {
		t.Fatal("expected a response error")
	}
	if responseError.ServerError == nil {
		t.Fatal("expected a server error")
	}

	nilError, ok := errors.AsType[*nil_error.Error](responseError.ServerError)
	if !ok {
		t.Fatalf("server error is not a *nil_error.Error: %v", responseError.ServerError)
	}
	if nilError.Field != "body parser" {
		t.Errorf("nil error field = %q, want %q", nilError.Field, "body parser")
	}
}
