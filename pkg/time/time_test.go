package time

import (
	"testing"
	"time"
)

func ptr(t time.Time) *time.Time { return &t }

func TestMin_Empty(t *testing.T) {
	t.Parallel()
	if got := Min(); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestMin_AllNil(t *testing.T) {
	t.Parallel()
	if got := Min(nil, nil); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestMin_PicksEarliest(t *testing.T) {
	t.Parallel()
	a := ptr(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	b := ptr(time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC))
	c := ptr(time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC))

	got := Min(a, b, c)
	if got == nil || !got.Equal(*c) {
		t.Fatalf("expected %v, got %v", c, got)
	}
}

func TestMin_IgnoresNil(t *testing.T) {
	t.Parallel()
	a := ptr(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	got := Min(nil, a, nil)
	if got == nil || !got.Equal(*a) {
		t.Fatalf("expected %v, got %v", a, got)
	}
}

func TestMax_Empty(t *testing.T) {
	t.Parallel()
	if got := Max(); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestMax_AllNil(t *testing.T) {
	t.Parallel()
	if got := Max(nil, nil); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestMax_PicksLatest(t *testing.T) {
	t.Parallel()
	a := ptr(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	b := ptr(time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC))
	c := ptr(time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC))

	got := Max(a, b, c)
	if got == nil || !got.Equal(*a) {
		t.Fatalf("expected %v, got %v", a, got)
	}
}

func TestMax_IgnoresNil(t *testing.T) {
	t.Parallel()
	a := ptr(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	got := Max(nil, a, nil)
	if got == nil || !got.Equal(*a) {
		t.Fatalf("expected %v, got %v", a, got)
	}
}
