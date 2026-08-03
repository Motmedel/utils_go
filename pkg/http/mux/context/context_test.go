package context

import (
	stdcontext "context"
	"testing"
)

func TestContentNegotiationContextKey_NotNil(t *testing.T) {
	t.Parallel()

	if ContentNegotiationContextKey == nil {
		t.Fatal("ContentNegotiationContextKey is nil")
	}
}

func TestContentNegotiationContextKey_RoundTrip(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		value any
	}{
		{name: "string", value: "text/html"},
		{name: "int", value: 42},
		{name: "nil", value: nil},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			ctx := stdcontext.WithValue(stdcontext.Background(), ContentNegotiationContextKey, testCase.value)

			got := ctx.Value(ContentNegotiationContextKey)
			if got != testCase.value {
				t.Errorf("ctx.Value = %v, want %v", got, testCase.value)
			}
		})
	}
}

func TestContentNegotiationContextKey_MissingReturnsNil(t *testing.T) {
	t.Parallel()

	if got := stdcontext.Background().Value(ContentNegotiationContextKey); got != nil {
		t.Errorf("value for absent key = %v, want nil", got)
	}
}

func TestContentNegotiationContextKey_StableIdentity(t *testing.T) {
	t.Parallel()

	// The exported key is a stable pointer; storing under it and reading it back
	// with the same package-level variable must resolve to the stored value even
	// when an unrelated key of a different type coexists.
	type otherKeyType struct{}
	otherKey := &otherKeyType{}

	ctx := stdcontext.WithValue(stdcontext.Background(), ContentNegotiationContextKey, "negotiated")
	ctx = stdcontext.WithValue(ctx, otherKey, "other")

	if got := ctx.Value(ContentNegotiationContextKey); got != "negotiated" {
		t.Errorf("ContentNegotiationContextKey value = %v, want %q", got, "negotiated")
	}
	if got := ctx.Value(otherKey); got != "other" {
		t.Errorf("otherKey value = %v, want %q", got, "other")
	}
}
