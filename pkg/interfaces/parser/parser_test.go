package parser

import (
	"context"
	"errors"
	"strconv"
	"testing"
)

var errSentinel = errors.New("parser failure")

// mockParser is a minimal implementation of the Parser interface.
type mockParser struct{}

func (mockParser) Parse(input string) (int, error) {
	return strconv.Atoi(input)
}

// mockParserCtx is a minimal implementation of the ParserCtx interface.
type mockParserCtx struct{}

func (mockParserCtx) Parse(_ context.Context, input string) (int, error) {
	return strconv.Atoi(input)
}

// Compile-time assertions that the mocks and the function adapters satisfy the interfaces.
var (
	_ Parser[int, string]    = mockParser{}
	_ Parser[int, string]    = Function[int, string](nil)
	_ ParserCtx[int, string] = mockParserCtx{}
	_ ParserCtx[int, string] = CtxFunction[int, string](nil)
)

func TestParserThroughInterface(t *testing.T) {
	t.Parallel()

	var parser Parser[int, string] = mockParser{}

	got, err := parser.Parse("42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 42 {
		t.Fatalf("expected 42, got %d", got)
	}

	if _, err = parser.Parse("not-a-number"); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNew(t *testing.T) {
	t.Parallel()

	parser := New(func(input string) (string, error) {
		return input + "!", errSentinel
	})

	got, err := parser.Parse("hi")
	if got != "hi!" {
		t.Fatalf("expected %q, got %q", "hi!", got)
	}
	if !errors.Is(err, errSentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestNewCtx(t *testing.T) {
	t.Parallel()

	type ctxKey struct{}

	parser := NewCtx(func(ctx context.Context, input string) (string, error) {
		v, _ := ctx.Value(ctxKey{}).(string)
		return input + v, nil
	})

	ctx := context.WithValue(context.Background(), ctxKey{}, "-suffix")
	got, err := parser.Parse(ctx, "value")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "value-suffix" {
		t.Fatalf("expected %q, got %q", "value-suffix", got)
	}
}

func TestCtxFunctionErrorPropagation(t *testing.T) {
	t.Parallel()

	fn := CtxFunction[int, string](func(_ context.Context, _ string) (int, error) {
		return 0, errSentinel
	})

	if _, err := fn.Parse(context.Background(), "x"); !errors.Is(err, errSentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}
