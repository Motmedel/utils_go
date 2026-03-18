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

	"github.com/Motmedel/utils_go/pkg/cloud/gws/gmail/gmail_config"
	"github.com/Motmedel/utils_go/pkg/cloud/gws/gmail/types/message"
	"github.com/Motmedel/utils_go/pkg/cloud/gws/gmail/types/send_as"
)

const Domain = "gmail.googleapis.com"

var defaultBaseUrl = &url.URL{
	Scheme: "https",
	Host:   Domain,
}

type Client struct {
	baseUrl *url.URL
	config  *gmail_config.Config
}

func NewClient(options ...gmail_config.Option) *Client {
	return NewClientWithBaseUrl(defaultBaseUrl, options...)
}

func NewClientWithBaseUrl(baseUrl *url.URL, options ...gmail_config.Option) *Client {
	u := *baseUrl
	u.Path = "/gmail/v1/users/"
	return &Client{baseUrl: &u, config: gmail_config.New(options...)}
}

func (c *Client) sendUrl(userId string) string {
	u := *c.baseUrl
	u.Path += url.PathEscape(userId) + "/messages/send"
	return u.String()
}

func (c *Client) sendAsUrl(userId string, sendAsEmail string) string {
	u := *c.baseUrl
	u.Path += url.PathEscape(userId) + "/settings/sendAs"
	if sendAsEmail != "" {
		u.Path += "/" + url.PathEscape(sendAsEmail)
	}
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
	options = append(append(c.config.FetchOptions, options...), fetch_config.WithMethod(http.MethodPost))
	_, sentMessage, err := motmedelHttpUtils.FetchJsonWithBody[*message.Message](ctx, urlString, msg, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json with body: %w", err), urlString)
	}

	return sentMessage, nil
}

// CreateSendAs creates a custom "from" send-as alias for the given user.
func (c *Client) CreateSendAs(ctx context.Context, userId string, s *send_as.SendAs, options ...fetch_config.Option) (*send_as.SendAs, error) {
	if userId == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("user id"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	if s == nil {
		return nil, nil
	}

	urlString := c.sendAsUrl(userId, "")
	options = append(append(c.config.FetchOptions, options...), fetch_config.WithMethod(http.MethodPost))
	_, created, err := motmedelHttpUtils.FetchJsonWithBody[*send_as.SendAs](ctx, urlString, s, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json with body: %w", err), urlString)
	}

	return created, nil
}

// GetSendAs retrieves a send-as alias identified by sendAsEmail for the given user.
func (c *Client) GetSendAs(ctx context.Context, userId string, sendAsEmail string, options ...fetch_config.Option) (*send_as.SendAs, error) {
	if userId == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("user id"))
	}
	if sendAsEmail == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("send-as email"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	urlString := c.sendAsUrl(userId, sendAsEmail)
	options = append(c.config.FetchOptions, options...)
	_, s, err := motmedelHttpUtils.FetchJson[*send_as.SendAs](ctx, urlString, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json: %w", err), urlString)
	}

	return s, nil
}

// UpdateSendAs updates a send-as alias identified by sendAsEmail for the given user.
func (c *Client) UpdateSendAs(ctx context.Context, userId string, sendAsEmail string, s *send_as.SendAs, options ...fetch_config.Option) (*send_as.SendAs, error) {
	if userId == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("user id"))
	}
	if sendAsEmail == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("send-as email"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	if s == nil {
		return nil, nil
	}

	urlString := c.sendAsUrl(userId, sendAsEmail)
	options = append(append(c.config.FetchOptions, options...), fetch_config.WithMethod(http.MethodPut))
	_, updated, err := motmedelHttpUtils.FetchJsonWithBody[*send_as.SendAs](ctx, urlString, s, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json with body: %w", err), urlString)
	}

	return updated, nil
}

// DeleteSendAs deletes a send-as alias identified by sendAsEmail for the given user.
func (c *Client) DeleteSendAs(ctx context.Context, userId string, sendAsEmail string, options ...fetch_config.Option) error {
	if userId == "" {
		return motmedelErrors.NewWithTrace(empty_error.New("user id"))
	}
	if sendAsEmail == "" {
		return motmedelErrors.NewWithTrace(empty_error.New("send-as email"))
	}

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context err: %w", err)
	}

	urlString := c.sendAsUrl(userId, sendAsEmail)
	options = append(append(c.config.FetchOptions, options...), fetch_config.WithMethod(http.MethodDelete))
	_, _, err := motmedelHttpUtils.Fetch(ctx, urlString, options...)
	if err != nil {
		return motmedelErrors.New(fmt.Errorf("fetch: %w", err), urlString)
	}

	return nil
}
