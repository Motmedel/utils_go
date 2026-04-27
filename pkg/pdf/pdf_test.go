package pdf

import (
	"testing"
)

func TestIsSigned(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{
			name: "signed document with both markers",
			data: []byte("some pdf content /Type /Sig more content /ByteRange [0 1 2 3]"),
			want: true,
		},
		{
			name: "only Type Sig marker",
			data: []byte("some pdf content /Type /Sig more content"),
			want: false,
		},
		{
			name: "only ByteRange marker",
			data: []byte("some pdf content /ByteRange [0 1 2 3]"),
			want: false,
		},
		{
			name: "no signature markers",
			data: []byte("some plain pdf content without signatures"),
			want: false,
		},
		{
			name: "empty data",
			data: []byte(""),
			want: false,
		},
		{
			name: "nil data",
			data: nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsSigned(tt.data); got != tt.want {
				t.Fatalf("IsSigned() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsEncrypted(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{
			name: "encrypted document",
			data: []byte("trailer\n<< /Size 42 /Encrypt 5 0 R /Root 1 0 R >>"),
			want: true,
		},
		{
			name: "unencrypted document",
			data: []byte("trailer\n<< /Size 42 /Root 1 0 R >>"),
			want: false,
		},
		{
			name: "empty data",
			data: []byte(""),
			want: false,
		},
		{
			name: "nil data",
			data: nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsEncrypted(tt.data); got != tt.want {
				t.Fatalf("IsEncrypted() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewPdfFileValidator(t *testing.T) {
	t.Parallel()
	v := NewPdfFileValidator()
	if v == nil {
		t.Fatalf("NewPdfFileValidator() returned nil")
	}
	if v.ExpectedFileExtension != ".pdf" {
		t.Fatalf("expected file extension .pdf, got %s", v.ExpectedFileExtension)
	}
	if v.ExpectedContentType != "application/pdf" {
		t.Fatalf("expected content type application/pdf, got %s", v.ExpectedContentType)
	}
}
