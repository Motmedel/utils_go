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

	"github.com/Motmedel/utils_go/pkg/cloud/gws/gmail/get_message_config"
	"github.com/Motmedel/utils_go/pkg/cloud/gws/gmail/gmail_config"
	"github.com/Motmedel/utils_go/pkg/cloud/gws/gmail/list_history_config"
	"github.com/Motmedel/utils_go/pkg/cloud/gws/gmail/types/filter"
	"github.com/Motmedel/utils_go/pkg/cloud/gws/gmail/types/history"
	"github.com/Motmedel/utils_go/pkg/cloud/gws/gmail/types/message"
	"github.com/Motmedel/utils_go/pkg/cloud/gws/gmail/types/send_as"
	"github.com/Motmedel/utils_go/pkg/cloud/gws/gmail/types/watch_request"
	"github.com/Motmedel/utils_go/pkg/cloud/gws/gmail/types/watch_response"
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

func (c *Client) messagesUrl(userId string, messageId string) string {
	u := *c.baseUrl
	u.Path += url.PathEscape(userId) + "/messages"
	if messageId != "" {
		u.Path += "/" + url.PathEscape(messageId)
	}
	return u.String()
}

func (c *Client) sendUrl(userId string) string {
	u := *c.baseUrl
	u.Path += url.PathEscape(userId) + "/messages/send"
	return u.String()
}

func (c *Client) watchUrl(userId string) string {
	u := *c.baseUrl
	u.Path += url.PathEscape(userId) + "/watch"
	return u.String()
}

func (c *Client) historyUrl(userId string) string {
	u := *c.baseUrl
	u.Path += url.PathEscape(userId) + "/history"
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

func (c *Client) filtersUrl(userId string, filterId string) string {
	u := *c.baseUrl
	u.Path += url.PathEscape(userId) + "/settings/filters"
	if filterId != "" {
		u.Path += "/" + url.PathEscape(filterId)
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

// Watch sets up or renews a push notification watch on the given user's mailbox.
// Notifications are delivered to the Cloud Pub/Sub topic specified in the request's TopicName.
func (c *Client) Watch(ctx context.Context, userId string, request *watch_request.WatchRequest, options ...fetch_config.Option) (*watch_response.WatchResponse, error) {
	if userId == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("user id"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	if request == nil {
		return nil, nil
	}

	urlString := c.watchUrl(userId)
	options = append(append(c.config.FetchOptions, options...), fetch_config.WithMethod(http.MethodPost))
	_, response, err := motmedelHttpUtils.FetchJsonWithBody[*watch_response.WatchResponse](ctx, urlString, request, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json with body: %w", err), urlString)
	}

	return response, nil
}

type listHistoryResponse struct {
	History       []*history.Record `json:"history"`
	NextPageToken string            `json:"nextPageToken"`
	HistoryId     string            `json:"historyId"`
}

// ListHistory retrieves all history records for the given user after the specified startHistoryId.
func (c *Client) ListHistory(ctx context.Context, userId string, startHistoryId string, options ...list_history_config.Option) ([]*history.Record, error) {
	if userId == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("user id"))
	}
	if startHistoryId == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("start history id"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	listHistoryConfig := list_history_config.New(options...)

	var allRecords []*history.Record
	pageToken := ""

	for {
		urlObj, err := url.Parse(c.historyUrl(userId))
		if err != nil {
			return nil, motmedelErrors.NewWithTrace(fmt.Errorf("url parse: %w", err))
		}

		query := url.Values{}
		query.Set("startHistoryId", startHistoryId)
		for _, historyType := range listHistoryConfig.HistoryTypes {
			query.Add("historyTypes", string(historyType))
		}
		if listHistoryConfig.LabelId != "" {
			query.Set("labelId", listHistoryConfig.LabelId)
		}
		if pageToken != "" {
			query.Set("pageToken", pageToken)
		}
		urlObj.RawQuery = query.Encode()
		urlString := urlObj.String()

		fetchOptions := append(c.config.FetchOptions, listHistoryConfig.FetchOptions...)
		_, resp, err := motmedelHttpUtils.FetchJson[*listHistoryResponse](ctx, urlString, fetchOptions...)
		if err != nil {
			return nil, motmedelErrors.New(fmt.Errorf("fetch json: %w", err), urlString)
		}

		if resp != nil {
			allRecords = append(allRecords, resp.History...)

			if resp.NextPageToken == "" {
				break
			}
			pageToken = resp.NextPageToken
		} else {
			break
		}
	}

	return allRecords, nil
}

type listMessagesResponse struct {
	Messages           []*message.Message `json:"messages"`
	NextPageToken      string             `json:"nextPageToken"`
	ResultSizeEstimate int                `json:"resultSizeEstimate"`
}

// ListMessages retrieves all messages for the given user matching the optional query.
// The query string uses the same format as the Gmail search box (e.g. "in:inbox", "from:user@example.com").
// Only message IDs and thread IDs are populated; use GetMessage to retrieve the full message.
func (c *Client) ListMessages(ctx context.Context, userId string, q string, options ...fetch_config.Option) ([]*message.Message, error) {
	if userId == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("user id"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	var allMessages []*message.Message
	pageToken := ""

	for {
		urlObj, err := url.Parse(c.messagesUrl(userId, ""))
		if err != nil {
			return nil, motmedelErrors.NewWithTrace(fmt.Errorf("url parse: %w", err))
		}

		query := url.Values{}
		if q != "" {
			query.Set("q", q)
		}
		if pageToken != "" {
			query.Set("pageToken", pageToken)
		}
		urlObj.RawQuery = query.Encode()
		urlString := urlObj.String()

		paginatedOptions := append(c.config.FetchOptions, options...)
		_, resp, err := motmedelHttpUtils.FetchJson[*listMessagesResponse](ctx, urlString, paginatedOptions...)
		if err != nil {
			return nil, motmedelErrors.New(fmt.Errorf("fetch json: %w", err), urlString)
		}

		if resp != nil {
			allMessages = append(allMessages, resp.Messages...)

			if resp.NextPageToken == "" {
				break
			}
			pageToken = resp.NextPageToken
		} else {
			break
		}
	}

	return allMessages, nil
}

// GetMessage retrieves a message identified by messageId for the given user.
func (c *Client) GetMessage(ctx context.Context, userId string, messageId string, options ...get_message_config.Option) (*message.Message, error) {
	if userId == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("user id"))
	}
	if messageId == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("message id"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	getMessageConfig := get_message_config.New(options...)

	urlObj, err := url.Parse(c.messagesUrl(userId, messageId))
	if err != nil {
		return nil, motmedelErrors.NewWithTrace(fmt.Errorf("url parse: %w", err))
	}

	query := urlObj.Query()
	if getMessageConfig.Format != "" {
		query.Set("format", string(getMessageConfig.Format))
	}
	for _, header := range getMessageConfig.MetadataHeaders {
		query.Add("metadataHeaders", header)
	}
	urlObj.RawQuery = query.Encode()
	urlString := urlObj.String()

	fetchOptions := append(c.config.FetchOptions, getMessageConfig.FetchOptions...)
	_, msg, err := motmedelHttpUtils.FetchJson[*message.Message](ctx, urlString, fetchOptions...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json: %w", err), urlString)
	}

	return msg, nil
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

type listFiltersResponse struct {
	Filter []*filter.Filter `json:"filter"`
}

// CreateFilter creates a filter for the given user.
func (c *Client) CreateFilter(ctx context.Context, userId string, f *filter.Filter, options ...fetch_config.Option) (*filter.Filter, error) {
	if userId == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("user id"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	if f == nil {
		return nil, nil
	}

	urlString := c.filtersUrl(userId, "")
	options = append(append(c.config.FetchOptions, options...), fetch_config.WithMethod(http.MethodPost))
	_, created, err := motmedelHttpUtils.FetchJsonWithBody[*filter.Filter](ctx, urlString, f, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json with body: %w", err), urlString)
	}

	return created, nil
}

// GetFilter retrieves a filter identified by filterId for the given user.
func (c *Client) GetFilter(ctx context.Context, userId string, filterId string, options ...fetch_config.Option) (*filter.Filter, error) {
	if userId == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("user id"))
	}
	if filterId == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("filter id"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	urlString := c.filtersUrl(userId, filterId)
	options = append(c.config.FetchOptions, options...)
	_, f, err := motmedelHttpUtils.FetchJson[*filter.Filter](ctx, urlString, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json: %w", err), urlString)
	}

	return f, nil
}

// ListFilters retrieves all filters for the given user.
func (c *Client) ListFilters(ctx context.Context, userId string, options ...fetch_config.Option) ([]*filter.Filter, error) {
	if userId == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("user id"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	urlString := c.filtersUrl(userId, "")
	options = append(c.config.FetchOptions, options...)
	_, resp, err := motmedelHttpUtils.FetchJson[*listFiltersResponse](ctx, urlString, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json: %w", err), urlString)
	}

	if resp == nil {
		return nil, nil
	}

	return resp.Filter, nil
}

// DeleteFilter deletes a filter identified by filterId for the given user.
func (c *Client) DeleteFilter(ctx context.Context, userId string, filterId string, options ...fetch_config.Option) error {
	if userId == "" {
		return motmedelErrors.NewWithTrace(empty_error.New("user id"))
	}
	if filterId == "" {
		return motmedelErrors.NewWithTrace(empty_error.New("filter id"))
	}

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context err: %w", err)
	}

	urlString := c.filtersUrl(userId, filterId)
	options = append(append(c.config.FetchOptions, options...), fetch_config.WithMethod(http.MethodDelete))
	_, _, err := motmedelHttpUtils.Fetch(ctx, urlString, options...)
	if err != nil {
		return motmedelErrors.New(fmt.Errorf("fetch: %w", err), urlString)
	}

	return nil
}
