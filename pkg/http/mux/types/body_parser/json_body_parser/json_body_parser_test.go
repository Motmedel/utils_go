package json_body_parser

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type simpleStruct struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

type nestedStruct struct {
	ID   string       `json:"id"`
	Data simpleStruct `json:"data"`
}

type structWithSlice struct {
	Items []string `json:"items"`
}

type structWithMap struct {
	Properties map[string]int `json:"properties"`
}

type structWithPointer struct {
	Name  string  `json:"name"`
	Value *string `json:"value"`
}

func TestNew(t *testing.T) {
	t.Run("creates parser", func(t *testing.T) {
		parser := New[simpleStruct]()
		if parser == nil {
			t.Fatal("expected non-nil parser")
		}
	})
}

func TestParser_Parse(t *testing.T) {
	t.Run("valid JSON parses successfully", func(t *testing.T) {
		parser := New[simpleStruct]()
		request := httptest.NewRequest(http.MethodPost, "/test", nil)
		body := []byte(`{"name":"test","value":42}`)

		result, responseError := parser.Parse(request, body)

		if responseError != nil {
			t.Fatalf("unexpected response error: %v", responseError)
		}

		if result.Name != "test" {
			t.Errorf("got name %q, expected %q", result.Name, "test")
		}

		if result.Value != 42 {
			t.Errorf("got value %d, expected %d", result.Value, 42)
		}
	})

	t.Run("parses to pointer type", func(t *testing.T) {
		parser := New[*simpleStruct]()
		request := httptest.NewRequest(http.MethodPost, "/test", nil)
		body := []byte(`{"name":"test","value":42}`)

		result, responseError := parser.Parse(request, body)

		if responseError != nil {
			t.Fatalf("unexpected response error: %v", responseError)
		}

		if result == nil {
			t.Fatal("expected non-nil result")
		}

		if result.Name != "test" {
			t.Errorf("got name %q, expected %q", result.Name, "test")
		}

		if result.Value != 42 {
			t.Errorf("got value %d, expected %d", result.Value, 42)
		}
	})

	t.Run("parses nested struct", func(t *testing.T) {
		parser := New[nestedStruct]()
		request := httptest.NewRequest(http.MethodPost, "/test", nil)
		body := []byte(`{"id":"abc123","data":{"name":"nested","value":100}}`)

		result, responseError := parser.Parse(request, body)

		if responseError != nil {
			t.Fatalf("unexpected response error: %v", responseError)
		}

		if result.ID != "abc123" {
			t.Errorf("got id %q, expected %q", result.ID, "abc123")
		}

		if result.Data.Name != "nested" {
			t.Errorf("got data.name %q, expected %q", result.Data.Name, "nested")
		}

		if result.Data.Value != 100 {
			t.Errorf("got data.value %d, expected %d", result.Data.Value, 100)
		}
	})

	t.Run("parses struct with slice", func(t *testing.T) {
		parser := New[structWithSlice]()
		request := httptest.NewRequest(http.MethodPost, "/test", nil)
		body := []byte(`{"items":["a","b","c"]}`)

		result, responseError := parser.Parse(request, body)

		if responseError != nil {
			t.Fatalf("unexpected response error: %v", responseError)
		}

		expectedItems := []string{"a", "b", "c"}
		if len(result.Items) != len(expectedItems) {
			t.Errorf("got %d items, expected %d", len(result.Items), len(expectedItems))
		} else {
			for i, item := range result.Items {
				if item != expectedItems[i] {
					t.Errorf("got item[%d] %q, expected %q", i, item, expectedItems[i])
				}
			}
		}
	})

	t.Run("parses struct with map", func(t *testing.T) {
		parser := New[structWithMap]()
		request := httptest.NewRequest(http.MethodPost, "/test", nil)
		body := []byte(`{"properties":{"a":1,"b":2,"c":3}}`)

		result, responseError := parser.Parse(request, body)

		if responseError != nil {
			t.Fatalf("unexpected response error: %v", responseError)
		}

		expectedProperties := map[string]int{"a": 1, "b": 2, "c": 3}
		if len(result.Properties) != len(expectedProperties) {
			t.Errorf("got %d properties, expected %d", len(result.Properties), len(expectedProperties))
		} else {
			for key, value := range expectedProperties {
				if result.Properties[key] != value {
					t.Errorf("got properties[%q] %d, expected %d", key, result.Properties[key], value)
				}
			}
		}
	})

	t.Run("parses struct with null pointer field", func(t *testing.T) {
		parser := New[structWithPointer]()
		request := httptest.NewRequest(http.MethodPost, "/test", nil)
		body := []byte(`{"name":"test","value":null}`)

		result, responseError := parser.Parse(request, body)

		if responseError != nil {
			t.Fatalf("unexpected response error: %v", responseError)
		}

		if result.Name != "test" {
			t.Errorf("got name %q, expected %q", result.Name, "test")
		}

		if result.Value != nil {
			t.Errorf("got value %v, expected nil", result.Value)
		}
	})

	t.Run("parses struct with non-null pointer field", func(t *testing.T) {
		parser := New[structWithPointer]()
		request := httptest.NewRequest(http.MethodPost, "/test", nil)
		body := []byte(`{"name":"test","value":"hello"}`)

		result, responseError := parser.Parse(request, body)

		if responseError != nil {
			t.Fatalf("unexpected response error: %v", responseError)
		}

		if result.Name != "test" {
			t.Errorf("got name %q, expected %q", result.Name, "test")
		}

		if result.Value == nil {
			t.Fatal("expected non-nil value")
		}

		if *result.Value != "hello" {
			t.Errorf("got value %q, expected %q", *result.Value, "hello")
		}
	})

	t.Run("invalid JSON returns server error", func(t *testing.T) {
		parser := New[simpleStruct]()
		request := httptest.NewRequest(http.MethodPost, "/test", nil)
		body := []byte(`{"name":"test","value":42`)

		_, responseError := parser.Parse(request, body)

		if responseError == nil {
			t.Fatal("expected response error, got nil")
		}

		if responseError.ServerError == nil {
			t.Fatal("expected server error, got nil")
		}
	})

	t.Run("type mismatch returns client error with 422", func(t *testing.T) {
		parser := New[simpleStruct]()
		request := httptest.NewRequest(http.MethodPost, "/test", nil)
		body := []byte(`{"name":"test","value":"not an int"}`)

		_, responseError := parser.Parse(request, body)

		if responseError == nil {
			t.Fatal("expected response error, got nil")
		}

		if responseError.ClientError == nil {
			t.Fatal("expected client error, got nil")
		}

		if responseError.ProblemDetail == nil {
			t.Fatal("expected problem detail, got nil")
		}

		expectedStatusCode := http.StatusUnprocessableEntity
		if responseError.ProblemDetail.Status != expectedStatusCode {
			t.Errorf("got status %d, expected %d", responseError.ProblemDetail.Status, expectedStatusCode)
		}

		expectedDetail := "Invalid body. The value is not appropriate for the JSON type."
		if responseError.ProblemDetail.Detail != expectedDetail {
			t.Errorf("got detail %q, expected %q", responseError.ProblemDetail.Detail, expectedDetail)
		}
	})

	t.Run("parses to map type", func(t *testing.T) {
		parser := New[map[string]any]()
		request := httptest.NewRequest(http.MethodPost, "/test", nil)
		body := []byte(`{"name":"test","value":42}`)

		result, responseError := parser.Parse(request, body)

		if responseError != nil {
			t.Fatalf("unexpected response error: %v", responseError)
		}

		if result["name"] != "test" {
			t.Errorf("got name %v, expected %q", result["name"], "test")
		}

		value, ok := result["value"].(float64)
		if !ok {
			t.Fatalf("expected value to be float64, got %T", result["value"])
		}

		if value != 42 {
			t.Errorf("got value %f, expected %f", value, 42.0)
		}
	})

	t.Run("parses to slice type", func(t *testing.T) {
		parser := New[[]string]()
		request := httptest.NewRequest(http.MethodPost, "/test", nil)
		body := []byte(`["a","b","c"]`)

		result, responseError := parser.Parse(request, body)

		if responseError != nil {
			t.Fatalf("unexpected response error: %v", responseError)
		}

		expectedItems := []string{"a", "b", "c"}
		if len(result) != len(expectedItems) {
			t.Errorf("got %d items, expected %d", len(result), len(expectedItems))
		} else {
			for i, item := range result {
				if item != expectedItems[i] {
					t.Errorf("got item[%d] %q, expected %q", i, item, expectedItems[i])
				}
			}
		}
	})

	t.Run("empty body returns error", func(t *testing.T) {
		parser := New[simpleStruct]()
		request := httptest.NewRequest(http.MethodPost, "/test", nil)
		body := []byte("")

		_, responseError := parser.Parse(request, body)

		if responseError == nil {
			t.Fatal("expected response error, got nil")
		}
	})

	t.Run("null body parses to zero value struct", func(t *testing.T) {
		parser := New[simpleStruct]()
		request := httptest.NewRequest(http.MethodPost, "/test", nil)
		body := []byte("null")

		result, responseError := parser.Parse(request, body)

		if responseError != nil {
			t.Fatalf("unexpected response error: %v", responseError)
		}

		if result.Name != "" {
			t.Errorf("got name %q, expected empty string", result.Name)
		}

		if result.Value != 0 {
			t.Errorf("got value %d, expected 0", result.Value)
		}
	})

	t.Run("parses numbers", func(t *testing.T) {
		parser := New[int]()
		request := httptest.NewRequest(http.MethodPost, "/test", nil)
		body := []byte("42")

		result, responseError := parser.Parse(request, body)

		if responseError != nil {
			t.Fatalf("unexpected response error: %v", responseError)
		}

		if result != 42 {
			t.Errorf("got %d, expected 42", result)
		}
	})

	t.Run("parses strings", func(t *testing.T) {
		parser := New[string]()
		request := httptest.NewRequest(http.MethodPost, "/test", nil)
		body := []byte(`"hello world"`)

		result, responseError := parser.Parse(request, body)

		if responseError != nil {
			t.Fatalf("unexpected response error: %v", responseError)
		}

		if result != "hello world" {
			t.Errorf("got %q, expected %q", result, "hello world")
		}
	})

	t.Run("parses booleans", func(t *testing.T) {
		parser := New[bool]()
		request := httptest.NewRequest(http.MethodPost, "/test", nil)
		body := []byte("true")

		result, responseError := parser.Parse(request, body)

		if responseError != nil {
			t.Fatalf("unexpected response error: %v", responseError)
		}

		if !result {
			t.Errorf("got %v, expected true", result)
		}
	})

	t.Run("parses floats", func(t *testing.T) {
		parser := New[float64]()
		request := httptest.NewRequest(http.MethodPost, "/test", nil)
		body := []byte("3.14159")

		result, responseError := parser.Parse(request, body)

		if responseError != nil {
			t.Fatalf("unexpected response error: %v", responseError)
		}

		if result != 3.14159 {
			t.Errorf("got %f, expected %f", result, 3.14159)
		}
	})

	t.Run("extra fields are ignored by default", func(t *testing.T) {
		parser := New[simpleStruct]()
		request := httptest.NewRequest(http.MethodPost, "/test", nil)
		body := []byte(`{"name":"test","value":42,"extra":"field"}`)

		result, responseError := parser.Parse(request, body)

		if responseError != nil {
			t.Fatalf("unexpected response error: %v", responseError)
		}

		if result.Name != "test" {
			t.Errorf("got name %q, expected %q", result.Name, "test")
		}

		if result.Value != 42 {
			t.Errorf("got value %d, expected %d", result.Value, 42)
		}
	})

	t.Run("missing fields get zero values", func(t *testing.T) {
		parser := New[simpleStruct]()
		request := httptest.NewRequest(http.MethodPost, "/test", nil)
		body := []byte(`{"name":"test"}`)

		result, responseError := parser.Parse(request, body)

		if responseError != nil {
			t.Fatalf("unexpected response error: %v", responseError)
		}

		if result.Name != "test" {
			t.Errorf("got name %q, expected %q", result.Name, "test")
		}

		if result.Value != 0 {
			t.Errorf("got value %d, expected 0", result.Value)
		}
	})

	t.Run("unicode content parses correctly", func(t *testing.T) {
		parser := New[simpleStruct]()
		request := httptest.NewRequest(http.MethodPost, "/test", nil)
		body := []byte(`{"name":"こんにちは","value":42}`)

		result, responseError := parser.Parse(request, body)

		if responseError != nil {
			t.Fatalf("unexpected response error: %v", responseError)
		}

		if result.Name != "こんにちは" {
			t.Errorf("got name %q, expected %q", result.Name, "こんにちは")
		}
	})

	t.Run("escaped characters parse correctly", func(t *testing.T) {
		parser := New[simpleStruct]()
		request := httptest.NewRequest(http.MethodPost, "/test", nil)
		body := []byte(`{"name":"hello\nworld\t\"quoted\"","value":42}`)

		result, responseError := parser.Parse(request, body)

		if responseError != nil {
			t.Fatalf("unexpected response error: %v", responseError)
		}

		expectedName := "hello\nworld\t\"quoted\""
		if result.Name != expectedName {
			t.Errorf("got name %q, expected %q", result.Name, expectedName)
		}
	})
}
