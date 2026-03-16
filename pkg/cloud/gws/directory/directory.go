package directory

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"github.com/Motmedel/utils_go/pkg/errors/types/empty_error"
	"github.com/Motmedel/utils_go/pkg/http/types/fetch_config"
	motmedelHttpUtils "github.com/Motmedel/utils_go/pkg/http/utils"

	"github.com/Motmedel/utils_go/pkg/cloud/gws/directory/types/group"
	"github.com/Motmedel/utils_go/pkg/cloud/gws/directory/types/member"
	"github.com/Motmedel/utils_go/pkg/cloud/gws/directory/types/user"
)

const Domain = "admin.googleapis.com"

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
	u.Path = "/admin/directory/v1/"
	return &Client{baseUrl: &u}
}

// User operations

// CreateUser creates a new user account.
func (c *Client) CreateUser(ctx context.Context, u *user.User, options ...fetch_config.Option) (*user.User, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	if u == nil {
		return nil, nil
	}

	urlObj := *c.baseUrl
	urlObj.Path += "users"
	urlString := urlObj.String()

	options = append(options, fetch_config.WithMethod(http.MethodPost))
	_, createdUser, err := motmedelHttpUtils.FetchJsonWithBody[*user.User](ctx, urlString, u, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json with body: %w", err), urlString)
	}

	return createdUser, nil
}

// GetUser retrieves a user account identified by userKey (primary email address, alias email address, or unique user ID).
func (c *Client) GetUser(ctx context.Context, userKey string, options ...fetch_config.Option) (*user.User, error) {
	if userKey == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("user key"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	urlObj := *c.baseUrl
	urlObj.Path += "users/" + url.PathEscape(userKey)
	urlString := urlObj.String()

	_, u, err := motmedelHttpUtils.FetchJson[*user.User](ctx, urlString, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json: %w", err), urlString)
	}

	return u, nil
}

// UpdateUser updates a user account identified by userKey.
func (c *Client) UpdateUser(ctx context.Context, userKey string, u *user.User, options ...fetch_config.Option) (*user.User, error) {
	if userKey == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("user key"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	if u == nil {
		return nil, nil
	}

	urlObj := *c.baseUrl
	urlObj.Path += "users/" + url.PathEscape(userKey)
	urlString := urlObj.String()

	options = append(options, fetch_config.WithMethod(http.MethodPut))
	_, updatedUser, err := motmedelHttpUtils.FetchJsonWithBody[*user.User](ctx, urlString, u, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json with body: %w", err), urlString)
	}

	return updatedUser, nil
}

// DeleteUser deletes a user account identified by userKey.
func (c *Client) DeleteUser(ctx context.Context, userKey string, options ...fetch_config.Option) error {
	if userKey == "" {
		return motmedelErrors.NewWithTrace(empty_error.New("user key"))
	}

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context err: %w", err)
	}

	urlObj := *c.baseUrl
	urlObj.Path += "users/" + url.PathEscape(userKey)
	urlString := urlObj.String()

	options = append(options, fetch_config.WithMethod(http.MethodDelete))
	_, _, err := motmedelHttpUtils.Fetch(ctx, urlString, options...)
	if err != nil {
		return motmedelErrors.New(fmt.Errorf("fetch: %w", err), urlString)
	}

	return nil
}

// Group operations

type listGroupsResponse struct {
	Groups        []*group.Group `json:"groups"`
	NextPageToken string         `json:"nextPageToken"`
}

// ListGroups retrieves all groups for the given customer ID (use "my_customer" for the authenticated account).
func (c *Client) ListGroups(ctx context.Context, customer string, options ...fetch_config.Option) ([]*group.Group, error) {
	if customer == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("customer"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	var allGroups []*group.Group
	pageToken := ""

	for {
		urlObj := *c.baseUrl
		urlObj.Path += "groups"

		query := url.Values{}
		query.Set("customer", customer)
		if pageToken != "" {
			query.Set("pageToken", pageToken)
		}
		urlObj.RawQuery = query.Encode()
		urlString := urlObj.String()

		_, resp, err := motmedelHttpUtils.FetchJson[*listGroupsResponse](ctx, urlString, options...)
		if err != nil {
			return nil, motmedelErrors.New(fmt.Errorf("fetch json: %w", err), urlString)
		}

		if resp != nil {
			allGroups = append(allGroups, resp.Groups...)

			if resp.NextPageToken == "" {
				break
			}
			pageToken = resp.NextPageToken
		} else {
			break
		}
	}

	return allGroups, nil
}

// CreateGroup creates a new group.
func (c *Client) CreateGroup(ctx context.Context, g *group.Group, options ...fetch_config.Option) (*group.Group, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	if g == nil {
		return nil, nil
	}

	urlObj := *c.baseUrl
	urlObj.Path += "groups"
	urlString := urlObj.String()

	options = append(options, fetch_config.WithMethod(http.MethodPost))
	_, createdGroup, err := motmedelHttpUtils.FetchJsonWithBody[*group.Group](ctx, urlString, g, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json with body: %w", err), urlString)
	}

	return createdGroup, nil
}

// GetGroup retrieves a group identified by groupKey (group email address, group alias, or unique group ID).
func (c *Client) GetGroup(ctx context.Context, groupKey string, options ...fetch_config.Option) (*group.Group, error) {
	if groupKey == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("group key"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	urlObj := *c.baseUrl
	urlObj.Path += "groups/" + url.PathEscape(groupKey)
	urlString := urlObj.String()

	_, g, err := motmedelHttpUtils.FetchJson[*group.Group](ctx, urlString, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json: %w", err), urlString)
	}

	return g, nil
}

// UpdateGroup updates a group identified by groupKey.
func (c *Client) UpdateGroup(ctx context.Context, groupKey string, g *group.Group, options ...fetch_config.Option) (*group.Group, error) {
	if groupKey == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("group key"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	if g == nil {
		return nil, nil
	}

	urlObj := *c.baseUrl
	urlObj.Path += "groups/" + url.PathEscape(groupKey)
	urlString := urlObj.String()

	options = append(options, fetch_config.WithMethod(http.MethodPut))
	_, updatedGroup, err := motmedelHttpUtils.FetchJsonWithBody[*group.Group](ctx, urlString, g, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json with body: %w", err), urlString)
	}

	return updatedGroup, nil
}

// DeleteGroup deletes a group identified by groupKey.
func (c *Client) DeleteGroup(ctx context.Context, groupKey string, options ...fetch_config.Option) error {
	if groupKey == "" {
		return motmedelErrors.NewWithTrace(empty_error.New("group key"))
	}

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context err: %w", err)
	}

	urlObj := *c.baseUrl
	urlObj.Path += "groups/" + url.PathEscape(groupKey)
	urlString := urlObj.String()

	options = append(options, fetch_config.WithMethod(http.MethodDelete))
	_, _, err := motmedelHttpUtils.Fetch(ctx, urlString, options...)
	if err != nil {
		return motmedelErrors.New(fmt.Errorf("fetch: %w", err), urlString)
	}

	return nil
}

// Group member operations

// CreateMember adds a member to a group identified by groupKey.
func (c *Client) CreateMember(ctx context.Context, groupKey string, m *member.Member, options ...fetch_config.Option) (*member.Member, error) {
	if groupKey == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("group key"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	if m == nil {
		return nil, nil
	}

	urlObj := *c.baseUrl
	urlObj.Path += "groups/" + url.PathEscape(groupKey) + "/members"
	urlString := urlObj.String()

	options = append(options, fetch_config.WithMethod(http.MethodPost))
	_, createdMember, err := motmedelHttpUtils.FetchJsonWithBody[*member.Member](ctx, urlString, m, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json with body: %w", err), urlString)
	}

	return createdMember, nil
}

// GetMember retrieves a member of a group identified by groupKey and memberKey (member email address or unique member ID).
func (c *Client) GetMember(ctx context.Context, groupKey string, memberKey string, options ...fetch_config.Option) (*member.Member, error) {
	if groupKey == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("group key"))
	}
	if memberKey == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("member key"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	urlObj := *c.baseUrl
	urlObj.Path += "groups/" + url.PathEscape(groupKey) + "/members/" + url.PathEscape(memberKey)
	urlString := urlObj.String()

	_, m, err := motmedelHttpUtils.FetchJson[*member.Member](ctx, urlString, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json: %w", err), urlString)
	}

	return m, nil
}

// UpdateMember updates a member of a group identified by groupKey and memberKey.
func (c *Client) UpdateMember(ctx context.Context, groupKey string, memberKey string, m *member.Member, options ...fetch_config.Option) (*member.Member, error) {
	if groupKey == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("group key"))
	}
	if memberKey == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("member key"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	if m == nil {
		return nil, nil
	}

	urlObj := *c.baseUrl
	urlObj.Path += "groups/" + url.PathEscape(groupKey) + "/members/" + url.PathEscape(memberKey)
	urlString := urlObj.String()

	options = append(options, fetch_config.WithMethod(http.MethodPut))
	_, updatedMember, err := motmedelHttpUtils.FetchJsonWithBody[*member.Member](ctx, urlString, m, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json with body: %w", err), urlString)
	}

	return updatedMember, nil
}

// DeleteMember removes a member from a group identified by groupKey and memberKey.
func (c *Client) DeleteMember(ctx context.Context, groupKey string, memberKey string, options ...fetch_config.Option) error {
	if groupKey == "" {
		return motmedelErrors.NewWithTrace(empty_error.New("group key"))
	}
	if memberKey == "" {
		return motmedelErrors.NewWithTrace(empty_error.New("member key"))
	}

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context err: %w", err)
	}

	urlObj := *c.baseUrl
	urlObj.Path += "groups/" + url.PathEscape(groupKey) + "/members/" + url.PathEscape(memberKey)
	urlString := urlObj.String()

	options = append(options, fetch_config.WithMethod(http.MethodDelete))
	_, _, err := motmedelHttpUtils.Fetch(ctx, urlString, options...)
	if err != nil {
		return motmedelErrors.New(fmt.Errorf("fetch: %w", err), urlString)
	}

	return nil
}
