package query_extractor

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	motmedelHttpErrors "github.com/Motmedel/utils_go/pkg/http/errors"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/request_parser/query_extractor/query_extractor_config"
	motmedelReflectErrors "github.com/Motmedel/utils_go/pkg/reflect/errors"
)

type basicQuery struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type optionalQuery struct {
	Required string `json:"required"`
	Optional string `json:"optional,omitempty"`
}

type intQuery struct {
	Count   int    `json:"count"`
	Size    int64  `json:"size"`
	Small   int8   `json:"small"`
	Medium  int16  `json:"medium"`
	Large   int32  `json:"large"`
	Huge    int64  `json:"huge"`
	Default int    `json:"default,omitempty"`
}

type uintQuery struct {
	Count   uint   `json:"count"`
	Size    uint64 `json:"size"`
	Small   uint8  `json:"small"`
	Medium  uint16 `json:"medium"`
	Large   uint32 `json:"large"`
	Huge    uint64 `json:"huge"`
	Default uint   `json:"default,omitempty"`
}

type floatQuery struct {
	Rate    float64 `json:"rate"`
	Ratio   float32 `json:"ratio"`
	Default float64 `json:"default,omitempty"`
}

type boolQuery struct {
	Enabled bool `json:"enabled"`
	Active  bool `json:"active,omitempty"`
}

type sliceQuery struct {
	Tags []string `json:"tags"`
	IDs  []int    `json:"ids"`
}

type arrayQuery struct {
	Coords [3]int `json:"coords"`
}

type byteSliceQuery struct {
	Data []byte `json:"data"`
}

type skipFieldQuery struct {
	Name    string `json:"name"`
	Skipped string `json:"-"`
}

type unexportedQuery struct {
	Name    string
	private string
}

type pointerQuery struct {
	Name  string  `json:"name"`
	Value *string `json:"value"`
}

type omitZeroQuery struct {
	Required string `json:"required"`
	Optional string `json:"optional,omitzero"`
}

func TestNew(t *testing.T) {
	t.Run("creates parser with default config", func(t *testing.T) {
		parser := New[basicQuery]()
		if parser == nil {
			t.Fatal("expected non-nil parser")
		}
	})

	t.Run("creates parser with options", func(t *testing.T) {
		parser := New[basicQuery](query_extractor_config.WithAllowAdditionalParameters(true))
		if parser == nil {
			t.Fatal("expected non-nil parser")
		}
	})
}

func TestParser_Parse(t *testing.T) {
	t.Run("nil request returns error", func(t *testing.T) {
		parser := New[basicQuery]()
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

	t.Run("nil URL returns error", func(t *testing.T) {
		parser := New[basicQuery]()
		request := &http.Request{URL: nil}
		_, responseError := parser.Parse(request)

		if responseError == nil {
			t.Fatal("expected response error, got nil")
		}

		if responseError.ServerError == nil {
			t.Fatal("expected server error, got nil")
		}

		if !errors.Is(responseError.ServerError, motmedelHttpErrors.ErrNilHttpRequestUrl) {
			t.Errorf("got error %v, expected %v", responseError.ServerError, motmedelHttpErrors.ErrNilHttpRequestUrl)
		}
	})

	t.Run("non-struct type returns error", func(t *testing.T) {
		parser := New[string]()
		request := httptest.NewRequest(http.MethodGet, "/test", nil)
		_, responseError := parser.Parse(request)

		if responseError == nil {
			t.Fatal("expected response error, got nil")
		}

		if responseError.ServerError == nil {
			t.Fatal("expected server error, got nil")
		}

		if !errors.Is(responseError.ServerError, motmedelReflectErrors.ErrNotStruct) {
			t.Errorf("got error %v, expected %v", responseError.ServerError, motmedelReflectErrors.ErrNotStruct)
		}
	})

	t.Run("malformed query returns error", func(t *testing.T) {
		parser := New[basicQuery]()
		request := httptest.NewRequest(http.MethodGet, "/test?%", nil)
		_, responseError := parser.Parse(request)

		if responseError == nil {
			t.Fatal("expected response error, got nil")
		}

		if responseError.ProblemDetail == nil {
			t.Fatal("expected problem detail, got nil")
		}

		expectedStatusCode := http.StatusBadRequest
		if responseError.ProblemDetail.Status != expectedStatusCode {
			t.Errorf("got status %d, expected %d", responseError.ProblemDetail.Status, expectedStatusCode)
		}

		expectedDetail := "Malformed query."
		if responseError.ProblemDetail.Detail != expectedDetail {
			t.Errorf("got detail %q, expected %q", responseError.ProblemDetail.Detail, expectedDetail)
		}
	})

	t.Run("missing required parameter returns error", func(t *testing.T) {
		parser := New[basicQuery]()
		request := httptest.NewRequest(http.MethodGet, "/test?name=test", nil)
		_, responseError := parser.Parse(request)

		if responseError == nil {
			t.Fatal("expected response error, got nil")
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

		errorsValue, ok := extension["errors"]
		if !ok {
			t.Fatal("expected 'errors' key in extension")
		}

		errorsList, ok := errorsValue.([]string)
		if !ok {
			t.Fatal("expected errors to be []string")
		}

		found := false
		for _, err := range errorsList {
			if err == "missing parameter: value" {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("expected error about missing 'value' parameter, got %v", errorsList)
		}
	})

	t.Run("successful basic query parsing", func(t *testing.T) {
		parser := New[basicQuery]()
		request := httptest.NewRequest(http.MethodGet, "/test?name=test&value=hello", nil)
		result, responseError := parser.Parse(request)

		if responseError != nil {
			t.Fatalf("unexpected response error: %v", responseError)
		}

		if result.Name != "test" {
			t.Errorf("got name %q, expected %q", result.Name, "test")
		}

		if result.Value != "hello" {
			t.Errorf("got value %q, expected %q", result.Value, "hello")
		}
	})

	t.Run("optional parameter can be omitted", func(t *testing.T) {
		parser := New[optionalQuery]()
		request := httptest.NewRequest(http.MethodGet, "/test?required=test", nil)
		result, responseError := parser.Parse(request)

		if responseError != nil {
			t.Fatalf("unexpected response error: %v", responseError)
		}

		if result.Required != "test" {
			t.Errorf("got required %q, expected %q", result.Required, "test")
		}

		if result.Optional != "" {
			t.Errorf("got optional %q, expected empty string", result.Optional)
		}
	})

	t.Run("omitzero parameter can be omitted", func(t *testing.T) {
		parser := New[omitZeroQuery]()
		request := httptest.NewRequest(http.MethodGet, "/test?required=test", nil)
		result, responseError := parser.Parse(request)

		if responseError != nil {
			t.Fatalf("unexpected response error: %v", responseError)
		}

		if result.Required != "test" {
			t.Errorf("got required %q, expected %q", result.Required, "test")
		}

		if result.Optional != "" {
			t.Errorf("got optional %q, expected empty string", result.Optional)
		}
	})

	t.Run("integer parsing", func(t *testing.T) {
		parser := New[intQuery]()
		request := httptest.NewRequest(http.MethodGet, "/test?count=42&size=9223372036854775807&small=127&medium=32767&large=2147483647&huge=9223372036854775807", nil)
		result, responseError := parser.Parse(request)

		if responseError != nil {
			t.Fatalf("unexpected response error: %v", responseError)
		}

		if result.Count != 42 {
			t.Errorf("got count %d, expected %d", result.Count, 42)
		}

		if result.Size != 9223372036854775807 {
			t.Errorf("got size %d, expected %d", result.Size, int64(9223372036854775807))
		}

		if result.Small != 127 {
			t.Errorf("got small %d, expected %d", result.Small, 127)
		}

		if result.Medium != 32767 {
			t.Errorf("got medium %d, expected %d", result.Medium, 32767)
		}

		if result.Large != 2147483647 {
			t.Errorf("got large %d, expected %d", result.Large, 2147483647)
		}
	})

	t.Run("invalid integer returns error", func(t *testing.T) {
		parser := New[intQuery]()
		request := httptest.NewRequest(http.MethodGet, "/test?count=notanumber&size=1&small=1&medium=1&large=1&huge=1", nil)
		_, responseError := parser.Parse(request)

		if responseError == nil {
			t.Fatal("expected response error, got nil")
		}

		if responseError.ProblemDetail == nil {
			t.Fatal("expected problem detail, got nil")
		}
	})

	t.Run("unsigned integer parsing", func(t *testing.T) {
		parser := New[uintQuery]()
		request := httptest.NewRequest(http.MethodGet, "/test?count=42&size=18446744073709551615&small=255&medium=65535&large=4294967295&huge=18446744073709551615", nil)
		result, responseError := parser.Parse(request)

		if responseError != nil {
			t.Fatalf("unexpected response error: %v", responseError)
		}

		if result.Count != 42 {
			t.Errorf("got count %d, expected %d", result.Count, 42)
		}

		if result.Small != 255 {
			t.Errorf("got small %d, expected %d", result.Small, 255)
		}

		if result.Medium != 65535 {
			t.Errorf("got medium %d, expected %d", result.Medium, 65535)
		}

		if result.Large != 4294967295 {
			t.Errorf("got large %d, expected %d", result.Large, 4294967295)
		}
	})

	t.Run("invalid unsigned integer returns error", func(t *testing.T) {
		parser := New[uintQuery]()
		request := httptest.NewRequest(http.MethodGet, "/test?count=-1&size=1&small=1&medium=1&large=1&huge=1", nil)
		_, responseError := parser.Parse(request)

		if responseError == nil {
			t.Fatal("expected response error, got nil")
		}

		if responseError.ProblemDetail == nil {
			t.Fatal("expected problem detail, got nil")
		}
	})

	t.Run("float parsing", func(t *testing.T) {
		parser := New[floatQuery]()
		request := httptest.NewRequest(http.MethodGet, "/test?rate=3.14159&ratio=2.5", nil)
		result, responseError := parser.Parse(request)

		if responseError != nil {
			t.Fatalf("unexpected response error: %v", responseError)
		}

		if result.Rate != 3.14159 {
			t.Errorf("got rate %f, expected %f", result.Rate, 3.14159)
		}

		if result.Ratio != 2.5 {
			t.Errorf("got ratio %f, expected %f", result.Ratio, 2.5)
		}
	})

	t.Run("invalid float returns error", func(t *testing.T) {
		parser := New[floatQuery]()
		request := httptest.NewRequest(http.MethodGet, "/test?rate=notafloat&ratio=1.0", nil)
		_, responseError := parser.Parse(request)

		if responseError == nil {
			t.Fatal("expected response error, got nil")
		}

		if responseError.ProblemDetail == nil {
			t.Fatal("expected problem detail, got nil")
		}
	})

	t.Run("boolean parsing with true values", func(t *testing.T) {
		testCases := []string{"true", "1", "t", "T", "TRUE", "True"}
		for _, value := range testCases {
			t.Run(value, func(t *testing.T) {
				parser := New[boolQuery]()
				request := httptest.NewRequest(http.MethodGet, "/test?enabled="+value, nil)
				result, responseError := parser.Parse(request)

				if responseError != nil {
					t.Fatalf("unexpected response error: %v", responseError)
				}

				if !result.Enabled {
					t.Errorf("got enabled %v, expected true for value %q", result.Enabled, value)
				}
			})
		}
	})

	t.Run("boolean parsing with false values", func(t *testing.T) {
		testCases := []string{"false", "0", "f", "F", "FALSE", "False"}
		for _, value := range testCases {
			t.Run(value, func(t *testing.T) {
				parser := New[boolQuery]()
				request := httptest.NewRequest(http.MethodGet, "/test?enabled="+value, nil)
				result, responseError := parser.Parse(request)

				if responseError != nil {
					t.Fatalf("unexpected response error: %v", responseError)
				}

				if result.Enabled {
					t.Errorf("got enabled %v, expected false for value %q", result.Enabled, value)
				}
			})
		}
	})

	t.Run("boolean with empty value is true", func(t *testing.T) {
		parser := New[boolQuery]()
		request := httptest.NewRequest(http.MethodGet, "/test?enabled", nil)
		result, responseError := parser.Parse(request)

		if responseError != nil {
			t.Fatalf("unexpected response error: %v", responseError)
		}

		if !result.Enabled {
			t.Errorf("got enabled %v, expected true", result.Enabled)
		}
	})

	t.Run("invalid boolean returns error", func(t *testing.T) {
		parser := New[boolQuery]()
		request := httptest.NewRequest(http.MethodGet, "/test?enabled=notabool", nil)
		_, responseError := parser.Parse(request)

		if responseError == nil {
			t.Fatal("expected response error, got nil")
		}

		if responseError.ProblemDetail == nil {
			t.Fatal("expected problem detail, got nil")
		}
	})

	t.Run("slice parsing", func(t *testing.T) {
		parser := New[sliceQuery]()
		request := httptest.NewRequest(http.MethodGet, "/test?tags=a&tags=b&tags=c&ids=1&ids=2&ids=3", nil)
		result, responseError := parser.Parse(request)

		if responseError != nil {
			t.Fatalf("unexpected response error: %v", responseError)
		}

		expectedTags := []string{"a", "b", "c"}
		if len(result.Tags) != len(expectedTags) {
			t.Errorf("got %d tags, expected %d", len(result.Tags), len(expectedTags))
		} else {
			for i, tag := range result.Tags {
				if tag != expectedTags[i] {
					t.Errorf("got tag[%d] %q, expected %q", i, tag, expectedTags[i])
				}
			}
		}

		expectedIDs := []int{1, 2, 3}
		if len(result.IDs) != len(expectedIDs) {
			t.Errorf("got %d ids, expected %d", len(result.IDs), len(expectedIDs))
		} else {
			for i, id := range result.IDs {
				if id != expectedIDs[i] {
					t.Errorf("got id[%d] %d, expected %d", i, id, expectedIDs[i])
				}
			}
		}
	})

	t.Run("array parsing", func(t *testing.T) {
		parser := New[arrayQuery]()
		request := httptest.NewRequest(http.MethodGet, "/test?coords=1&coords=2&coords=3", nil)
		result, responseError := parser.Parse(request)

		if responseError != nil {
			t.Fatalf("unexpected response error: %v", responseError)
		}

		expectedCoords := [3]int{1, 2, 3}
		if result.Coords != expectedCoords {
			t.Errorf("got coords %v, expected %v", result.Coords, expectedCoords)
		}
	})

	t.Run("array wrong count returns error", func(t *testing.T) {
		parser := New[arrayQuery]()
		request := httptest.NewRequest(http.MethodGet, "/test?coords=1&coords=2", nil)
		_, responseError := parser.Parse(request)

		if responseError == nil {
			t.Fatal("expected response error, got nil")
		}

		if responseError.ProblemDetail == nil {
			t.Fatal("expected problem detail, got nil")
		}
	})

	t.Run("byte slice parsing", func(t *testing.T) {
		parser := New[byteSliceQuery]()
		request := httptest.NewRequest(http.MethodGet, "/test?data=hello", nil)
		result, responseError := parser.Parse(request)

		if responseError != nil {
			t.Fatalf("unexpected response error: %v", responseError)
		}

		expectedData := []byte("hello")
		if string(result.Data) != string(expectedData) {
			t.Errorf("got data %q, expected %q", result.Data, expectedData)
		}
	})

	t.Run("byte slice multiple values returns error", func(t *testing.T) {
		parser := New[byteSliceQuery]()
		request := httptest.NewRequest(http.MethodGet, "/test?data=hello&data=world", nil)
		_, responseError := parser.Parse(request)

		if responseError == nil {
			t.Fatal("expected response error, got nil")
		}

		if responseError.ProblemDetail == nil {
			t.Fatal("expected problem detail, got nil")
		}
	})

	t.Run("skip field with json tag", func(t *testing.T) {
		parser := New[skipFieldQuery]()
		request := httptest.NewRequest(http.MethodGet, "/test?name=test", nil)
		result, responseError := parser.Parse(request)

		if responseError != nil {
			t.Fatalf("unexpected response error: %v", responseError)
		}

		if result.Name != "test" {
			t.Errorf("got name %q, expected %q", result.Name, "test")
		}
	})

	t.Run("unexported fields are ignored", func(t *testing.T) {
		parser := New[unexportedQuery]()
		request := httptest.NewRequest(http.MethodGet, "/test?Name=test", nil)
		result, responseError := parser.Parse(request)

		if responseError != nil {
			t.Fatalf("unexpected response error: %v", responseError)
		}

		if result.Name != "test" {
			t.Errorf("got name %q, expected %q", result.Name, "test")
		}
	})

	t.Run("pointer field returns server error", func(t *testing.T) {
		parser := New[pointerQuery]()
		request := httptest.NewRequest(http.MethodGet, "/test?name=test&value=hello", nil)
		_, responseError := parser.Parse(request)

		if responseError == nil {
			t.Fatal("expected response error, got nil")
		}

		if responseError.ServerError == nil {
			t.Fatal("expected server error, got nil")
		}
	})

	t.Run("unknown parameter without allow additional returns error", func(t *testing.T) {
		parser := New[basicQuery]()
		request := httptest.NewRequest(http.MethodGet, "/test?name=test&value=hello&unknown=foo", nil)
		_, responseError := parser.Parse(request)

		if responseError == nil {
			t.Fatal("expected response error, got nil")
		}

		if responseError.ProblemDetail == nil {
			t.Fatal("expected problem detail, got nil")
		}

		extension := responseError.ProblemDetail.Extension
		if extension == nil {
			t.Fatal("expected problem detail extension, got nil")
		}

		errorsValue, ok := extension["errors"]
		if !ok {
			t.Fatal("expected 'errors' key in extension")
		}

		errorsList, ok := errorsValue.([]string)
		if !ok {
			t.Fatal("expected errors to be []string")
		}

		found := false
		for _, err := range errorsList {
			if err == "unknown parameter: unknown" {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("expected error about unknown parameter, got %v", errorsList)
		}
	})

	t.Run("unknown parameter with allow additional succeeds", func(t *testing.T) {
		parser := New[basicQuery](query_extractor_config.WithAllowAdditionalParameters(true))
		request := httptest.NewRequest(http.MethodGet, "/test?name=test&value=hello&unknown=foo", nil)
		result, responseError := parser.Parse(request)

		if responseError != nil {
			t.Fatalf("unexpected response error: %v", responseError)
		}

		if result.Name != "test" {
			t.Errorf("got name %q, expected %q", result.Name, "test")
		}

		if result.Value != "hello" {
			t.Errorf("got value %q, expected %q", result.Value, "hello")
		}
	})

	t.Run("multiple values for scalar field returns error", func(t *testing.T) {
		parser := New[basicQuery]()
		request := httptest.NewRequest(http.MethodGet, "/test?name=test&name=test2&value=hello", nil)
		_, responseError := parser.Parse(request)

		if responseError == nil {
			t.Fatal("expected response error, got nil")
		}

		if responseError.ProblemDetail == nil {
			t.Fatal("expected problem detail, got nil")
		}
	})

	t.Run("parsing to pointer struct type", func(t *testing.T) {
		parser := New[*basicQuery]()
		request := httptest.NewRequest(http.MethodGet, "/test?name=test&value=hello", nil)
		result, responseError := parser.Parse(request)

		if responseError != nil {
			t.Fatalf("unexpected response error: %v", responseError)
		}

		if result == nil {
			t.Fatal("expected non-nil result")
		}

		if result.Name != "test" {
			t.Errorf("got name %q, expected %q", result.Name, "test")
		}

		if result.Value != "hello" {
			t.Errorf("got value %q, expected %q", result.Value, "hello")
		}
	})

	t.Run("URL encoded values", func(t *testing.T) {
		parser := New[basicQuery]()
		request := httptest.NewRequest(http.MethodGet, "/test?name=hello%20world&value=foo%2Bbar", nil)
		result, responseError := parser.Parse(request)

		if responseError != nil {
			t.Fatalf("unexpected response error: %v", responseError)
		}

		if result.Name != "hello world" {
			t.Errorf("got name %q, expected %q", result.Name, "hello world")
		}

		if result.Value != "foo+bar" {
			t.Errorf("got value %q, expected %q", result.Value, "foo+bar")
		}
	})
}
