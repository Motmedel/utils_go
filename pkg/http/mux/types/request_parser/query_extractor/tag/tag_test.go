package tag

import (
	"testing"
)

func TestNew_EmptyString(t *testing.T) {
	if New("") != nil {
		t.Fatal("expected nil for empty string")
	}
}

func TestNew_Skip(t *testing.T) {
	tag := New("-")
	if tag == nil {
		t.Fatal("expected non-nil tag")
	}
	if !tag.Skip {
		t.Fatal("expected Skip to be true")
	}
}

func TestNew_NameOnly(t *testing.T) {
	tag := New("my_param")
	if tag == nil {
		t.Fatal("expected non-nil tag")
	}
	if tag.Name != "my_param" {
		t.Fatalf("expected name 'my_param', got %q", tag.Name)
	}
	if tag.OmitEmpty || tag.OmitZero || tag.Skip {
		t.Fatal("expected no flags set")
	}
	if tag.Format != "" {
		t.Fatalf("expected empty format, got %q", tag.Format)
	}
}

func TestNew_OmitEmpty(t *testing.T) {
	tag := New("name,omitempty")
	if tag == nil {
		t.Fatal("expected non-nil tag")
	}
	if tag.Name != "name" {
		t.Fatalf("expected name 'name', got %q", tag.Name)
	}
	if !tag.OmitEmpty {
		t.Fatal("expected OmitEmpty to be true")
	}
}

func TestNew_OmitZero(t *testing.T) {
	tag := New("name,omitzero")
	if tag == nil {
		t.Fatal("expected non-nil tag")
	}
	if !tag.OmitZero {
		t.Fatal("expected OmitZero to be true")
	}
}

func TestNew_FormatEmail(t *testing.T) {
	tag := New("email_field,format=email")
	if tag == nil {
		t.Fatal("expected non-nil tag")
	}
	if tag.Name != "email_field" {
		t.Fatalf("expected name 'email_field', got %q", tag.Name)
	}
	if tag.Format != "email" {
		t.Fatalf("expected format 'email', got %q", tag.Format)
	}
}

func TestNew_FormatUuid(t *testing.T) {
	tag := New("id,format=uuid")
	if tag == nil {
		t.Fatal("expected non-nil tag")
	}
	if tag.Format != "uuid" {
		t.Fatalf("expected format 'uuid', got %q", tag.Format)
	}
}

func TestNew_AllOptions(t *testing.T) {
	tag := New("field,omitempty,omitzero,format=email")
	if tag == nil {
		t.Fatal("expected non-nil tag")
	}
	if tag.Name != "field" {
		t.Fatalf("expected name 'field', got %q", tag.Name)
	}
	if !tag.OmitEmpty {
		t.Fatal("expected OmitEmpty")
	}
	if !tag.OmitZero {
		t.Fatal("expected OmitZero")
	}
	if tag.Format != "email" {
		t.Fatalf("expected format 'email', got %q", tag.Format)
	}
}
