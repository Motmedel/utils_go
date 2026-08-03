package authenticator

import (
	"context"
	"errors"
	"testing"
)

var errSentinel = errors.New("authenticator failure")

// mockAuthenticator is a minimal implementation of the Authenticator interface.
type mockAuthenticator struct {
	result string
	err    error
}

func (m *mockAuthenticator) Authenticate(_ context.Context, _ int) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.result, nil
}

// Compile-time assertions that both the mock and the Function adapter satisfy the interface.
var (
	_ Authenticator[string, int] = (*mockAuthenticator)(nil)
	_ Authenticator[string, int] = Function[string, int](nil)
)

func TestMockThroughInterface(t *testing.T) {
	t.Parallel()

	var authenticator Authenticator[string, int] = &mockAuthenticator{result: "ok"}

	got, err := authenticator.Authenticate(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "ok" {
		t.Fatalf("expected %q, got %q", "ok", got)
	}
}

func TestMockErrorPropagation(t *testing.T) {
	t.Parallel()

	var authenticator Authenticator[string, int] = &mockAuthenticator{err: errSentinel}

	_, err := authenticator.Authenticate(context.Background(), 1)
	if !errors.Is(err, errSentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestNew(t *testing.T) {
	t.Parallel()

	type ctxKey struct{}
	want := "authenticated"

	authenticator := New(func(ctx context.Context, input string) (string, error) {
		if input != "input" {
			t.Errorf("expected input %q, got %q", "input", input)
		}
		if v, _ := ctx.Value(ctxKey{}).(string); v != "carried" {
			t.Errorf("expected context value %q, got %q", "carried", v)
		}
		return want, nil
	})

	ctx := context.WithValue(context.Background(), ctxKey{}, "carried")
	got, err := authenticator.Authenticate(ctx, "input")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestFunctionAuthenticate(t *testing.T) {
	t.Parallel()

	fn := Function[int, int](func(_ context.Context, input int) (int, error) {
		return input * 2, errSentinel
	})

	got, err := fn.Authenticate(context.Background(), 21)
	if got != 42 {
		t.Fatalf("expected 42, got %d", got)
	}
	if !errors.Is(err, errSentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}
