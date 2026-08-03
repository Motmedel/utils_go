package database

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDefaultTimeout(t *testing.T) {
	t.Parallel()

	if DefaultTimeout != 5*time.Second {
		t.Fatalf("expected DefaultTimeout of 5s, got %v", DefaultTimeout)
	}
}

func TestMakeTimeoutCtxDeadline(t *testing.T) {
	t.Parallel()

	before := time.Now()
	ctx, cancel := MakeTimeoutCtx(t.Context())
	defer cancel()

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("expected the context to carry a deadline")
	}

	// The deadline must land within (now, now+DefaultTimeout], allowing a
	// second of slack below for scheduling jitter.
	minDeadline := before.Add(DefaultTimeout - time.Second)
	maxDeadline := time.Now().Add(DefaultTimeout)
	if deadline.Before(minDeadline) || deadline.After(maxDeadline) {
		t.Fatalf("deadline %v not within expected window [%v, %v]", deadline, minDeadline, maxDeadline)
	}
}

func TestMakeTimeoutCtxCancel(t *testing.T) {
	t.Parallel()

	ctx, cancel := MakeTimeoutCtx(t.Context())

	if err := ctx.Err(); err != nil {
		t.Fatalf("expected no error before cancel, got %v", err)
	}

	cancel()

	if err := ctx.Err(); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled after cancel, got %v", err)
	}
}

func TestMakeTimeoutCtxParentCancel(t *testing.T) {
	t.Parallel()

	parent, cancelParent := context.WithCancel(t.Context())
	ctx, cancel := MakeTimeoutCtx(parent)
	defer cancel()

	cancelParent()

	<-ctx.Done()
	if err := ctx.Err(); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected parent cancellation to propagate, got %v", err)
	}
}
