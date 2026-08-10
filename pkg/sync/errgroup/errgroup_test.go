package errgroup

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

var (
	errFirst    = errors.New("first error")
	errSecond   = errors.New("second error")
	errExpected = errors.New("expected error")
)

func TestGroupAllSucceed(t *testing.T) {
	t.Parallel()

	group, groupCtx := WithContext(context.Background())

	var counter atomic.Int64
	for range 8 {
		group.Go(func() error {
			counter.Add(1)
			return nil
		})
	}

	if err := group.Wait(); err != nil {
		t.Fatalf("wait: %v", err)
	}
	if counter.Load() != 8 {
		t.Errorf("counter = %d, want 8", counter.Load())
	}

	select {
	case <-groupCtx.Done():
	default:
		t.Error("expected the context to be canceled after Wait")
	}
	if cause := context.Cause(groupCtx); !errors.Is(cause, context.Canceled) {
		t.Errorf("cause = %v, want %v", cause, context.Canceled)
	}
}

func TestGroupFirstErrorWins(t *testing.T) {
	t.Parallel()

	group, groupCtx := WithContext(context.Background())

	group.Go(func() error {
		return errFirst
	})

	// The second function starts only once the first error has canceled
	// the context, so its error cannot be the first.
	group.Go(func() error {
		<-groupCtx.Done()
		return fmt.Errorf("wrapped: %w", errSecond)
	})

	if err := group.Wait(); !errors.Is(err, errFirst) {
		t.Errorf("wait error = %v, want %v", err, errFirst)
	}
	if cause := context.Cause(groupCtx); !errors.Is(cause, errFirst) {
		t.Errorf("cause = %v, want %v", cause, errFirst)
	}
}

func TestGroupErrorCancelsContextBeforeWait(t *testing.T) {
	t.Parallel()

	group, groupCtx := WithContext(context.Background())

	group.Go(func() error {
		return errExpected
	})

	select {
	case <-groupCtx.Done():
	case <-time.After(5 * time.Second):
		t.Fatal("expected the context to be canceled before Wait")
	}

	if err := group.Wait(); !errors.Is(err, errExpected) {
		t.Errorf("wait error = %v, want %v", err, errExpected)
	}
}

func TestGroupParentCancellationPropagates(t *testing.T) {
	t.Parallel()

	parentCtx, parentCancel := context.WithCancel(context.Background())
	group, groupCtx := WithContext(parentCtx)

	group.Go(func() error {
		<-groupCtx.Done()
		return nil
	})

	parentCancel()

	if err := group.Wait(); err != nil {
		t.Errorf("wait: %v", err)
	}
}

func TestZeroValueGroup(t *testing.T) {
	t.Parallel()

	var group Group

	group.Go(func() error {
		return errExpected
	})
	group.Go(func() error {
		return nil
	})

	if err := group.Wait(); !errors.Is(err, errExpected) {
		t.Errorf("wait error = %v, want %v", err, errExpected)
	}
}
