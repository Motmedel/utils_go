package cache_group

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
)

var errSentinel = errors.New("cache failure")

func TestDoReturnsResult(t *testing.T) {
	t.Parallel()

	var group Group[int]
	got, err := group.Do("key", func() (int, error) {
		return 42, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 42 {
		t.Fatalf("expected 42, got %d", got)
	}
}

func TestDoCachesResult(t *testing.T) {
	t.Parallel()

	var group Group[int]
	var calls int32

	fn := func() (int, error) {
		atomic.AddInt32(&calls, 1)
		return 7, nil
	}

	for range 3 {
		got, err := group.Do("key", fn)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 7 {
			t.Fatalf("expected 7, got %d", got)
		}
	}

	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("expected fn to be called once, got %d", got)
	}
}

func TestDoDifferentKeys(t *testing.T) {
	t.Parallel()

	var group Group[string]
	var calls int32

	first, err := group.Do("a", func() (string, error) {
		atomic.AddInt32(&calls, 1)
		return "first", nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	second, err := group.Do("b", func() (string, error) {
		atomic.AddInt32(&calls, 1)
		return "second", nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if first != "first" || second != "second" {
		t.Fatalf("unexpected results: %q, %q", first, second)
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Fatalf("expected 2 calls, got %d", got)
	}
}

func TestDoCachesError(t *testing.T) {
	t.Parallel()

	var group Group[int]
	var calls int32

	fn := func() (int, error) {
		atomic.AddInt32(&calls, 1)
		return 0, errSentinel
	}

	for range 2 {
		_, err := group.Do("key", fn)
		if !errors.Is(err, errSentinel) {
			t.Fatalf("expected sentinel error, got %v", err)
		}
	}

	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("expected fn to be called once (error cached), got %d", got)
	}
}

func TestDoConcurrent(t *testing.T) {
	t.Parallel()

	var group Group[int]
	var calls int32

	const goroutines = 100

	var start sync.WaitGroup
	start.Add(1)

	var done sync.WaitGroup
	done.Add(goroutines)

	results := make([]int, goroutines)

	for i := range goroutines {
		go func() {
			defer done.Done()
			start.Wait()
			value, err := group.Do("key", func() (int, error) {
				atomic.AddInt32(&calls, 1)
				return 99, nil
			})
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			results[i] = value
		}()
	}

	start.Done()
	done.Wait()

	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("expected fn to be called exactly once, got %d", got)
	}
	for i, value := range results {
		if value != 99 {
			t.Fatalf("goroutine %d got %d, expected 99", i, value)
		}
	}
}
