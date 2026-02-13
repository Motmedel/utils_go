package query_extractor

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/Motmedel/utils_go/pkg/http/mux/types/request_parser/query_extractor/query_extractor_config"
)

type basicQuery struct {
	Name string `query:"name"`
	Age  int    `query:"age"`
}

type jsonFallbackQuery struct {
	Name string `json:"name"`
	Age  int    `json:"age,omitempty"`
}

type queryTagOverridesJson struct {
	Name string `query:"q_name" json:"j_name"`
}

type optionalQuery struct {
	Name     string `query:"name"`
	Nickname string `query:"nickname,omitempty"`
}

type skipQuery struct {
	Name    string `query:"name"`
	Skipped string `query:"-"`
}

type emailQuery struct {
	Email string `query:"email,format=email"`
}

type uuidQuery struct {
	ID string `query:"id,format=uuid"`
}

func makeRequest(rawQuery string) *http.Request {
	return &http.Request{
		URL: &url.URL{RawQuery: rawQuery},
	}
}

func TestParse_QueryTag(t *testing.T) {
	parser := New[basicQuery]()
	result, respErr := parser.Parse(makeRequest("name=alice&age=30"))
	if respErr != nil {
		t.Fatalf("unexpected error: %v", respErr)
	}
	if result.Name != "alice" {
		t.Fatalf("expected name 'alice', got %q", result.Name)
	}
	if result.Age != 30 {
		t.Fatalf("expected age 30, got %d", result.Age)
	}
}

func TestParse_JsonFallback(t *testing.T) {
	parser := New[jsonFallbackQuery]()
	result, respErr := parser.Parse(makeRequest("name=bob&age=25"))
	if respErr != nil {
		t.Fatalf("unexpected error: %v", respErr)
	}
	if result.Name != "bob" {
		t.Fatalf("expected name 'bob', got %q", result.Name)
	}
	if result.Age != 25 {
		t.Fatalf("expected age 25, got %d", result.Age)
	}
}

func TestParse_JsonFallbackOptional(t *testing.T) {
	parser := New[jsonFallbackQuery](query_extractor_config.WithAllowAdditionalParameters(true))
	result, respErr := parser.Parse(makeRequest("name=bob"))
	if respErr != nil {
		t.Fatalf("unexpected error: %v", respErr)
	}
	if result.Name != "bob" {
		t.Fatalf("expected name 'bob', got %q", result.Name)
	}
	if result.Age != 0 {
		t.Fatalf("expected age 0, got %d", result.Age)
	}
}

func TestParse_QueryTagOverridesJson(t *testing.T) {
	parser := New[queryTagOverridesJson](query_extractor_config.WithAllowAdditionalParameters(true))
	result, respErr := parser.Parse(makeRequest("q_name=alice"))
	if respErr != nil {
		t.Fatalf("unexpected error: %v", respErr)
	}
	if result.Name != "alice" {
		t.Fatalf("expected name 'alice', got %q", result.Name)
	}
}

func TestParse_QueryTagSkip(t *testing.T) {
	parser := New[skipQuery](query_extractor_config.WithAllowAdditionalParameters(true))
	result, respErr := parser.Parse(makeRequest("name=alice"))
	if respErr != nil {
		t.Fatalf("unexpected error: %v", respErr)
	}
	if result.Name != "alice" {
		t.Fatalf("expected name 'alice', got %q", result.Name)
	}
	if result.Skipped != "" {
		t.Fatalf("expected empty Skipped, got %q", result.Skipped)
	}
}

func TestParse_OptionalQuery(t *testing.T) {
	parser := New[optionalQuery](query_extractor_config.WithAllowAdditionalParameters(true))
	result, respErr := parser.Parse(makeRequest("name=alice"))
	if respErr != nil {
		t.Fatalf("unexpected error: %v", respErr)
	}
	if result.Name != "alice" {
		t.Fatalf("expected name 'alice', got %q", result.Name)
	}
	if result.Nickname != "" {
		t.Fatalf("expected empty Nickname, got %q", result.Nickname)
	}
}

func TestParse_EmailFormatValid(t *testing.T) {
	parser := New[emailQuery](query_extractor_config.WithAllowAdditionalParameters(true))
	result, respErr := parser.Parse(makeRequest("email=user@example.com"))
	if respErr != nil {
		t.Fatalf("unexpected error: %v", respErr)
	}
	if result.Email != "user@example.com" {
		t.Fatalf("expected 'user@example.com', got %q", result.Email)
	}
}

func TestParse_EmailFormatInvalid(t *testing.T) {
	parser := New[emailQuery](query_extractor_config.WithAllowAdditionalParameters(true))
	_, respErr := parser.Parse(makeRequest("email=not-an-email"))
	if respErr == nil {
		t.Fatal("expected error for invalid email format")
	}
}

func TestParse_UuidFormatValid(t *testing.T) {
	parser := New[uuidQuery](query_extractor_config.WithAllowAdditionalParameters(true))
	result, respErr := parser.Parse(makeRequest("id=550e8400-e29b-41d4-a716-446655440000"))
	if respErr != nil {
		t.Fatalf("unexpected error: %v", respErr)
	}
	if result.ID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Fatalf("unexpected id: %q", result.ID)
	}
}

func TestParse_UuidFormatInvalid(t *testing.T) {
	parser := New[uuidQuery](query_extractor_config.WithAllowAdditionalParameters(true))
	_, respErr := parser.Parse(makeRequest("id=not-a-uuid"))
	if respErr == nil {
		t.Fatal("expected error for invalid uuid format")
	}
}
