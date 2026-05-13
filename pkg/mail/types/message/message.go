package message

import (
	"crypto/rand"
	"fmt"
	"net/mail"
	"strings"
	"time"

	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"github.com/Motmedel/utils_go/pkg/errors/types/empty_error"
	"github.com/Motmedel/utils_go/pkg/errors/types/nil_error"
	motmedelMailErrors "github.com/Motmedel/utils_go/pkg/mail/errors"
	"github.com/Motmedel/utils_go/pkg/mail/types/message/message_config"
	"github.com/Motmedel/utils_go/pkg/mail/types/message/message_header"
)

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
	Headers []*message_header.Header
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
		return "", motmedelErrors.NewWithTrace(empty_error.New("domain"))
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

	for _, header := range message.Headers {
		if header == nil {
			continue
		}
		builder.WriteString(fmt.Sprintf("%s: %s\r\n", header.Name, header.Value))
	}

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

func New(from *mail.Address, to []*mail.Address, subject string, body *Body, options ...message_config.Option) (*Message, error) {
	if from == nil {
		return nil, motmedelErrors.NewWithTrace(nil_error.NewWithInstance("address", "from"))
	}

	if len(to) == 0 {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("to"))
	}

	if subject == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("subject"))
	}

	if body != nil && body.ContentType == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("content type"))
	}

	config := message_config.New(options...)

	return &Message{
		From:    from,
		To:      to,
		Cc:      config.Cc,
		Bcc:     config.Bcc,
		Subject: subject,
		Body:    body,
		ReplyTo: config.ReplyTo,
		Domain:  config.Domain,
		Headers: config.Headers,
	}, nil
}
