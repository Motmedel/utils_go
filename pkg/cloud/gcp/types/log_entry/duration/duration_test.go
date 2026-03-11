package duration

import (
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	d := 2*time.Second + 500*time.Millisecond
	result := New(&d)
	if result.Seconds != 2 {
		t.Errorf("expected 2 seconds, got %d", result.Seconds)
	}
	if result.Nanos != 500000000 {
		t.Errorf("expected 500000000 nanos, got %d", result.Nanos)
	}
}

func TestNew_Zero(t *testing.T) {
	d := time.Duration(0)
	result := New(&d)
	if result.Seconds != 0 {
		t.Errorf("expected 0 seconds, got %d", result.Seconds)
	}
	if result.Nanos != 0 {
		t.Errorf("expected 0 nanos, got %d", result.Nanos)
	}
}

func TestNew_SubSecond(t *testing.T) {
	d := 250 * time.Millisecond
	result := New(&d)
	if result.Seconds != 0 {
		t.Errorf("expected 0 seconds, got %d", result.Seconds)
	}
	if result.Nanos != 250000000 {
		t.Errorf("expected 250000000 nanos, got %d", result.Nanos)
	}
}

func TestNew_ExactSeconds(t *testing.T) {
	d := 5 * time.Second
	result := New(&d)
	if result.Seconds != 5 {
		t.Errorf("expected 5 seconds, got %d", result.Seconds)
	}
	if result.Nanos != 0 {
		t.Errorf("expected 0 nanos, got %d", result.Nanos)
	}
}
