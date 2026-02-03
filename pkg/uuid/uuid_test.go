package uuid

import (
	"regexp"
	"testing"
	"time"
)

var uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func TestNew(t *testing.T) {
	uuid := New()
	s := uuid.String()

	if !uuidRegex.MatchString(s) {
		t.Errorf("New() produced invalid UUID format: %s", s)
	}

	if uuid.Version() != 4 {
		t.Errorf("New() produced UUID with version %d, expected 4", uuid.Version())
	}

	if uuid.Variant() != 1 {
		t.Errorf("New() produced UUID with variant %d, expected 1 (RFC 4122)", uuid.Variant())
	}
}

func TestNewRandom(t *testing.T) {
	uuid, err := NewRandom()
	if err != nil {
		t.Fatalf("NewRandom() returned error: %v", err)
	}

	s := uuid.String()
	if !uuidRegex.MatchString(s) {
		t.Errorf("NewRandom() produced invalid UUID format: %s", s)
	}

	if uuid.Version() != 4 {
		t.Errorf("NewRandom() produced UUID with version %d, expected 4", uuid.Version())
	}
}

func TestNewV7(t *testing.T) {
	before := time.Now()
	uuid, err := NewV7()
	after := time.Now()

	if err != nil {
		t.Fatalf("NewV7() returned error: %v", err)
	}

	s := uuid.String()
	if !uuidRegex.MatchString(s) {
		t.Errorf("NewV7() produced invalid UUID format: %s", s)
	}

	if uuid.Version() != 7 {
		t.Errorf("NewV7() produced UUID with version %d, expected 7", uuid.Version())
	}

	if uuid.Variant() != 1 {
		t.Errorf("NewV7() produced UUID with variant %d, expected 1 (RFC 4122)", uuid.Variant())
	}

	// Check that timestamp is within expected range.
	ts := uuid.Timestamp()
	if ts.Before(before.Truncate(time.Millisecond)) || ts.After(after.Add(time.Millisecond)) {
		t.Errorf("NewV7() timestamp %v not within expected range [%v, %v]", ts, before, after)
	}
}

func TestNewV7FromTime(t *testing.T) {
	testTime := time.Date(2024, 6, 15, 12, 30, 45, 0, time.UTC)
	uuid, err := NewV7FromTime(testTime)
	if err != nil {
		t.Fatalf("NewV7FromTime() returned error: %v", err)
	}

	if uuid.Version() != 7 {
		t.Errorf("NewV7FromTime() produced UUID with version %d, expected 7", uuid.Version())
	}

	ts := uuid.Timestamp()
	if !ts.Equal(testTime.Truncate(time.Millisecond)) {
		t.Errorf("NewV7FromTime() timestamp %v doesn't match input %v", ts, testTime)
	}
}

func TestNewString(t *testing.T) {
	s := NewString()
	if !uuidRegex.MatchString(s) {
		t.Errorf("NewString() produced invalid UUID format: %s", s)
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid lowercase", "550e8400-e29b-41d4-a716-446655440000", false},
		{"valid uppercase", "550E8400-E29B-41D4-A716-446655440000", false},
		{"too short", "550e8400-e29b-41d4-a716", true},
		{"too long", "550e8400-e29b-41d4-a716-4466554400001", true},
		{"wrong format", "550e8400e29b41d4a716446655440000", true},
		{"invalid chars", "550e8400-e29b-41d4-a716-44665544000g", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestParseRoundTrip(t *testing.T) {
	original := New()
	s := original.String()

	parsed, err := Parse(s)
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if !original.Equal(parsed) {
		t.Errorf("Round trip failed: original %v != parsed %v", original, parsed)
	}
}

func TestMustParse(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("MustParse() did not panic on invalid input")
		}
	}()

	MustParse("invalid")
}

func TestFromBytes(t *testing.T) {
	original := New()
	bytes := original.Bytes()

	restored, err := FromBytes(bytes)
	if err != nil {
		t.Fatalf("FromBytes() returned error: %v", err)
	}

	if !original.Equal(restored) {
		t.Errorf("FromBytes() result doesn't match original: %v != %v", restored, original)
	}
}

func TestFromBytesInvalidLength(t *testing.T) {
	_, err := FromBytes([]byte{1, 2, 3})
	if err != ErrInvalidUUIDLength {
		t.Errorf("FromBytes() with invalid length should return ErrInvalidUUIDLength, got %v", err)
	}
}

func TestURN(t *testing.T) {
	uuid := MustParse("550e8400-e29b-41d4-a716-446655440000")
	urn := uuid.URN()
	expected := "urn:uuid:550e8400-e29b-41d4-a716-446655440000"

	if urn != expected {
		t.Errorf("URN() = %q, want %q", urn, expected)
	}
}

func TestIsNil(t *testing.T) {
	var nilUUID UUID
	if !nilUUID.IsNil() {
		t.Error("Zero UUID should be nil")
	}

	uuid := New()
	if uuid.IsNil() {
		t.Error("Generated UUID should not be nil")
	}
}

func TestMarshalText(t *testing.T) {
	uuid := MustParse("550e8400-e29b-41d4-a716-446655440000")
	text, err := uuid.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText() returned error: %v", err)
	}

	if string(text) != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("MarshalText() = %q, want %q", text, "550e8400-e29b-41d4-a716-446655440000")
	}
}

func TestUnmarshalText(t *testing.T) {
	var uuid UUID
	err := uuid.UnmarshalText([]byte("550e8400-e29b-41d4-a716-446655440000"))
	if err != nil {
		t.Fatalf("UnmarshalText() returned error: %v", err)
	}

	expected := MustParse("550e8400-e29b-41d4-a716-446655440000")
	if !uuid.Equal(expected) {
		t.Errorf("UnmarshalText() result = %v, want %v", uuid, expected)
	}
}

func TestMarshalBinary(t *testing.T) {
	uuid := New()
	data, err := uuid.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary() returned error: %v", err)
	}

	if len(data) != 16 {
		t.Errorf("MarshalBinary() returned %d bytes, want 16", len(data))
	}
}

func TestUnmarshalBinary(t *testing.T) {
	original := New()
	data, _ := original.MarshalBinary()

	var uuid UUID
	err := uuid.UnmarshalBinary(data)
	if err != nil {
		t.Fatalf("UnmarshalBinary() returned error: %v", err)
	}

	if !uuid.Equal(original) {
		t.Errorf("UnmarshalBinary() result = %v, want %v", uuid, original)
	}
}

func TestUnmarshalBinaryInvalidLength(t *testing.T) {
	var uuid UUID
	err := uuid.UnmarshalBinary([]byte{1, 2, 3})
	if err != ErrInvalidUUIDLength {
		t.Errorf("UnmarshalBinary() with invalid length should return ErrInvalidUUIDLength, got %v", err)
	}
}

func TestUint64Pair(t *testing.T) {
	uuid := New()
	high, low := uuid.ToUint64Pair()

	restored := FromUint64Pair(high, low)
	if !uuid.Equal(restored) {
		t.Errorf("Uint64Pair round trip failed: %v != %v", uuid, restored)
	}
}

func TestUniqueness(t *testing.T) {
	seen := make(map[UUID]bool)
	const count = 1000

	for i := 0; i < count; i++ {
		uuid := New()
		if seen[uuid] {
			t.Fatalf("Duplicate UUID generated: %v", uuid)
		}
		seen[uuid] = true
	}
}

func TestV7Ordering(t *testing.T) {
	// Generate multiple v7 UUIDs and verify they are in order.
	var uuids []UUID
	for i := 0; i < 100; i++ {
		uuid, err := NewV7()
		if err != nil {
			t.Fatalf("NewV7() returned error: %v", err)
		}
		uuids = append(uuids, uuid)
		time.Sleep(time.Millisecond)
	}

	for i := 1; i < len(uuids); i++ {
		prev := uuids[i-1].Timestamp()
		curr := uuids[i].Timestamp()
		if curr.Before(prev) {
			t.Errorf("UUID v7 ordering violated: %v (ts=%v) should come after %v (ts=%v)",
				uuids[i], curr, uuids[i-1], prev)
		}
	}
}

func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = New()
	}
}

func BenchmarkNewV7(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = NewV7()
	}
}

func BenchmarkNewString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewString()
	}
}

func BenchmarkParse(b *testing.B) {
	s := "550e8400-e29b-41d4-a716-446655440000"
	for i := 0; i < b.N; i++ {
		_, _ = Parse(s)
	}
}
