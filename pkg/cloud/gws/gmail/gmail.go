package gmail

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"github.com/Motmedel/utils_go/pkg/errors/types/empty_error"
	"github.com/Motmedel/utils_go/pkg/http/types/fetch_config"
	motmedelHttpUtils "github.com/Motmedel/utils_go/pkg/http/utils"

	"github.com/Motmedel/utils_go/pkg/cloud/gws/gmail/types/message"
)

const Domain = "gmail.googleapis.com"

var defaultBaseUrl = &url.URL{
	Scheme: "https",
	Host:   Domain,
}

type Client struct {
	baseUrl *url.URL
}

func NewClient() *Client {
	return NewClientWithBaseUrl(defaultBaseUrl)
}

func NewClientWithBaseUrl(baseUrl *url.URL) *Client {
	u := *baseUrl
	u.Path = "/gmail/v1/users/"
	return &Client{baseUrl: &u}
}

func (c *Client) sendUrl(userId string) string {
	u := *c.baseUrl
	u.Path += url.PathEscape(userId) + "/messages/send"
	return u.String()
}

// Send sends the specified message to the recipients in the To, Cc, and Bcc headers.
// The message should have its Raw field set to a base64url-encoded RFC 2822 email.
func (c *Client) Send(ctx context.Context, userId string, msg *message.Message, options ...fetch_config.Option) (*message.Message, error) {
	if userId == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("user id"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	if msg == nil {
		return nil, nil
	}

	urlString := c.sendUrl(userId)
	options = append(options, fetch_config.WithMethod(http.MethodPost))
	_, sentMessage, err := motmedelHttpUtils.FetchJsonWithBody[*message.Message](ctx, urlString, msg, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json with body: %w", err), urlString)
	}

	return sentMessage, nil
}
