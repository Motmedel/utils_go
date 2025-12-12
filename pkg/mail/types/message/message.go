package message

import (
	"crypto/rand"
	"fmt"
	"net/mail"
	"strings"
	"time"

	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelMailErrors "github.com/Motmedel/utils_go/pkg/mail/errors"
)

type Option func(*Message)

type Body struct {
	Content     []byte
	ContentType string
}

type Message struct {
	From    *mail.Address
	To      []*mail.Address
	Cc      []*mail.Address
	Bcc     []*mail.Address
	Subject string
	Body    *Body
	ReplyTo []*mail.Address
	Domain  string
}

func (message *Message) String() (string, error) {
	var builder strings.Builder
	timeNow := time.Now()

	fromMailAddress := message.From

	builder.WriteString(fmt.Sprintf("From: %s\r\n", fromMailAddress.String()))

	var toStrings []string
	for _, to := range message.To {
		if to != nil {
			toStrings = append(toStrings, to.String())
		}
	}
	builder.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(toStrings, ", ")))

	if replyTo := message.ReplyTo; len(replyTo) > 0 {
		var replyToStrings []string
		for _, replyToAddress := range replyTo {
			replyToStrings = append(replyToStrings, replyToAddress.String())
		}
		builder.WriteString(fmt.Sprintf("Reply-To: %s\r\n", strings.Join(replyToStrings, ", ")))
	}
	if cc := message.Cc; len(cc) > 0 {
		var ccStrings []string
		for _, ccAddress := range cc {
			ccStrings = append(ccStrings, ccAddress.String())
		}
		builder.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(ccStrings, ", ")))
	}
	if bcc := message.Bcc; len(bcc) > 0 {
		var bccStrings []string
		for _, bccAddress := range bcc {
			bccStrings = append(bccStrings, bccAddress.String())
		}
		builder.WriteString(fmt.Sprintf("Bcc: %s\r\n", strings.Join(bccStrings, ", ")))
	}
	builder.WriteString(fmt.Sprintf("Subject: %s\r\n", message.Subject))
	builder.WriteString(fmt.Sprintf("Date: %s\r\n", timeNow.Format(time.RFC1123Z)))

	domain := message.Domain
	if domain == "" {
		fromAddress := fromMailAddress.Address
		if at := strings.LastIndex(fromAddress, "@"); at != -1 && at+1 < len(fromAddress) {
			domain = fromMailAddress.Address[at+1:]
		} else {
			return "", motmedelErrors.NewWithTrace(motmedelMailErrors.ErrBadFromAddress)
		}
	}
	if domain == "" {
		return "", motmedelErrors.NewWithTrace(motmedelMailErrors.ErrEmptyDomain)
	}

	messageIdRandomBuffer := make([]byte, 16)
	if _, err := rand.Read(messageIdRandomBuffer); err != nil {
		return "", motmedelErrors.NewWithTrace(fmt.Errorf("rand read: %w", err))
	}
	builder.WriteString(
		fmt.Sprintf(
			"Message-ID: %s\r\n",
			fmt.Sprintf("<%d.%x@%s>", timeNow.UnixNano(), messageIdRandomBuffer, domain),
		),
	)

	body := message.Body

	builder.WriteString("MIME-Version: 1.0\r\n")
	if body != nil && body.ContentType != "" {
		builder.WriteString(fmt.Sprintf("Content-Type: %s\r\n", body.ContentType))
	}
	builder.WriteString("\r\n")
	if body != nil && len(body.Content) > 0 {
		builder.Write(body.Content)
	}

	return builder.String(), nil
}

func New(from *mail.Address, to []*mail.Address, subject string, body *Body, options ...Option) (*Message, error) {
	if from == nil {
		return nil, motmedelErrors.NewWithTrace(fmt.Errorf("%w (from)", motmedelMailErrors.ErrNilAddress))
	}

	if len(to) == 0 {
		return nil, motmedelErrors.NewWithTrace(motmedelMailErrors.ErrEmptyTo)
	}

	if subject == "" {
		return nil, motmedelErrors.NewWithTrace(motmedelMailErrors.ErrEmptySubject)
	}

	if body != nil && body.ContentType == "" {
		return nil, motmedelErrors.NewWithTrace(motmedelMailErrors.ErrEmptyContentType)
	}

	config := &Message{
		From:    from,
		To:      to,
		Subject: subject,
		Body:    body,
	}

	for _, option := range options {
		if option != nil {
			option(config)
		}
	}

	return config, nil
}

func WithCc(cc []*mail.Address) Option {
	return func(config *Message) {
		config.Cc = cc
	}
}

func WithBcc(bcc []*mail.Address) Option {
	return func(config *Message) {
		config.Bcc = bcc
	}
}

func WithReplyTo(replyTo []*mail.Address) Option {
	return func(config *Message) {
		config.ReplyTo = replyTo
	}
}

func WithDomain(domain string) Option {
	return func(config *Message) {
		config.Domain = domain
	}
}
