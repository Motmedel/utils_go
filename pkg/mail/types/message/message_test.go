package message

import (
	"errors"
	"net/mail"
	"reflect"
	"strings"
	"testing"

	"github.com/Motmedel/utils_go/pkg/errors/types/empty_error"
	"github.com/Motmedel/utils_go/pkg/errors/types/nil_error"
	motmedelMailErrors "github.com/Motmedel/utils_go/pkg/mail/errors"
	"github.com/Motmedel/utils_go/pkg/mail/types/message/message_config"
	"github.com/Motmedel/utils_go/pkg/mail/types/message/message_header"
)

func validFrom() *mail.Address { return &mail.Address{Address: "from@example.com"} }
func validTo() []*mail.Address  { return []*mail.Address{{Address: "to@example.com"}} }

func TestNew_NilFrom(t *testing.T) {
	t.Parallel()

	msg, err := New(nil, validTo(), "subj", nil)
	if msg != nil {
		t.Fatalf("msg = %v, want nil", msg)
	}
	ne, ok := errors.AsType[*nil_error.Error](err)
	if !ok {
		t.Fatalf("err type = %T (%v), want *nil_error.Error", err, err)
	}
	if ne.Field != "address" || ne.Instance != "from" {
		t.Errorf("ne = %+v, want Field=address Instance=from", ne)
	}
}

func TestNew_EmptyTo(t *testing.T) {
	t.Parallel()

	_, err := New(validFrom(), nil, "subj", nil)
	ee, ok := errors.AsType[*empty_error.Error](err)
	if !ok {
		t.Fatalf("err type = %T (%v), want *empty_error.Error", err, err)
	}
	if ee.Field != "to" {
		t.Errorf("Field = %q, want %q", ee.Field, "to")
	}
}

func TestNew_EmptySubject(t *testing.T) {
	t.Parallel()

	_, err := New(validFrom(), validTo(), "", nil)
	ee, ok := errors.AsType[*empty_error.Error](err)
	if !ok {
		t.Fatalf("err type = %T (%v), want *empty_error.Error", err, err)
	}
	if ee.Field != "subject" {
		t.Errorf("Field = %q, want %q", ee.Field, "subject")
	}
}

func TestNew_EmptyContentType(t *testing.T) {
	t.Parallel()

	_, err := New(validFrom(), validTo(), "subj", &Body{Content: []byte("hi")})
	ee, ok := errors.AsType[*empty_error.Error](err)
	if !ok {
		t.Fatalf("err type = %T (%v), want *empty_error.Error", err, err)
	}
	if ee.Field != "content type" {
		t.Errorf("Field = %q, want %q", ee.Field, "content type")
	}
}

func TestNew_AppliesOptions(t *testing.T) {
	t.Parallel()

	cc := []*mail.Address{{Address: "cc@example.com"}}
	bcc := []*mail.Address{{Address: "bcc@example.com"}}
	replyTo := []*mail.Address{{Address: "reply@example.com"}}
	headers := []*message_header.Header{
		{Name: "X-Foo", Value: "bar"},
		{Name: "X-Baz", Value: "qux"},
	}

	msg, err := New(
		validFrom(),
		validTo(),
		"subj",
		&Body{Content: []byte("hi"), ContentType: "text/plain"},
		message_config.WithCc(cc),
		message_config.WithBcc(bcc),
		message_config.WithReplyTo(replyTo),
		message_config.WithDomain("custom.example"),
		message_config.WithHeaders(headers...),
		nil,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(msg.Cc, cc) {
		t.Errorf("Cc = %v, want %v", msg.Cc, cc)
	}
	if !reflect.DeepEqual(msg.Bcc, bcc) {
		t.Errorf("Bcc = %v, want %v", msg.Bcc, bcc)
	}
	if !reflect.DeepEqual(msg.ReplyTo, replyTo) {
		t.Errorf("ReplyTo = %v, want %v", msg.ReplyTo, replyTo)
	}
	if msg.Domain != "custom.example" {
		t.Errorf("Domain = %q, want %q", msg.Domain, "custom.example")
	}
	if !reflect.DeepEqual(msg.Headers, headers) {
		t.Errorf("Headers = %v, want %v", msg.Headers, headers)
	}
}

func TestMessage_String_IncludesHeaders(t *testing.T) {
	t.Parallel()

	msg, err := New(
		validFrom(),
		validTo(),
		"subj",
		&Body{Content: []byte("hi"), ContentType: "text/plain"},
		message_config.WithHeaders(
			&message_header.Header{Name: "X-Foo", Value: "bar"},
			nil,
			&message_header.Header{Name: "List-Unsubscribe", Value: "<mailto:u@example.com>"},
		),
	)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	out, err := msg.String()
	if err != nil {
		t.Fatalf("String: %v", err)
	}
	if !strings.Contains(out, "\r\nX-Foo: bar\r\n") {
		t.Errorf("output missing X-Foo header:\n%s", out)
	}
	if !strings.Contains(out, "\r\nList-Unsubscribe: <mailto:u@example.com>\r\n") {
		t.Errorf("output missing List-Unsubscribe header:\n%s", out)
	}
	if strings.Contains(out, ": \r\n\r\n") {
		t.Errorf("nil header should be skipped:\n%s", out)
	}
}

func TestMessage_String_DerivesDomainFromFrom(t *testing.T) {
	t.Parallel()

	msg, err := New(
		&mail.Address{Address: "alice@derived.example"},
		validTo(),
		"subj",
		&Body{Content: []byte("hi"), ContentType: "text/plain"},
	)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	out, err := msg.String()
	if err != nil {
		t.Fatalf("String: %v", err)
	}
	if !strings.Contains(out, "@derived.example>\r\n") {
		t.Errorf("Message-ID did not use From domain:\n%s", out)
	}
}

func TestMessage_String_BadFromAddress(t *testing.T) {
	t.Parallel()

	msg := &Message{
		From:    &mail.Address{Address: "no-at-sign"},
		To:      validTo(),
		Subject: "subj",
	}
	if _, err := msg.String(); !errors.Is(err, motmedelMailErrors.ErrBadFromAddress) {
		t.Fatalf("err = %v, want ErrBadFromAddress", err)
	}
}
