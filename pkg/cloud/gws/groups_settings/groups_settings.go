package groups_settings

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"github.com/Motmedel/utils_go/pkg/errors/types/empty_error"
	"github.com/Motmedel/utils_go/pkg/http/types/fetch_config"
	motmedelHttpUtils "github.com/Motmedel/utils_go/pkg/http/utils"

	"github.com/Motmedel/utils_go/pkg/cloud/gws/groups_settings/groups_settings_config"
	"github.com/Motmedel/utils_go/pkg/cloud/gws/groups_settings/types/group"
)

const Domain = "groups.googleapis.com"

var defaultBaseUrl = &url.URL{
	Scheme: "https",
	Host:   Domain,
}

type Client struct {
	baseUrl *url.URL
	config  *groups_settings_config.Config
}

func NewClient(options ...groups_settings_config.Option) *Client {
	return NewClientWithBaseUrl(defaultBaseUrl, options...)
}

func NewClientWithBaseUrl(baseUrl *url.URL, options ...groups_settings_config.Option) *Client {
	u := *baseUrl
	u.Path = "/groups/v1/groups/"
	return &Client{baseUrl: &u, config: groups_settings_config.New(options...)}
}

func (c *Client) groupUrl(groupEmail string) string {
	u := *c.baseUrl
	u.Path += url.PathEscape(groupEmail)
	return u.String()
}

// Get retrieves a group's settings identified by the group email address.
func (c *Client) Get(ctx context.Context, groupEmail string, options ...fetch_config.Option) (*group.Group, error) {
	if groupEmail == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("group email"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	urlString := c.groupUrl(groupEmail)
	options = append(c.config.FetchOptions, options...)
	_, groupSettings, err := motmedelHttpUtils.FetchJson[*group.Group](ctx, urlString, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json: %w", err), urlString)
	}

	return groupSettings, nil
}

// Update updates an existing group's settings identified by the group email address.
func (c *Client) Update(ctx context.Context, groupEmail string, groupSettings *group.Group, options ...fetch_config.Option) (*group.Group, error) {
	if groupEmail == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("group email"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	urlString := c.groupUrl(groupEmail)
	options = append(append(c.config.FetchOptions, options...), fetch_config.WithMethod(http.MethodPut))
	_, updatedGroupSettings, err := motmedelHttpUtils.FetchJsonWithBody[*group.Group](ctx, urlString, groupSettings, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json with body: %w", err), urlString)
	}

	return updatedGroupSettings, nil
}

// Patch updates an existing group's settings using patch semantics.
func (c *Client) Patch(ctx context.Context, groupEmail string, groupSettings *group.Group, options ...fetch_config.Option) (*group.Group, error) {
	if groupEmail == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("group email"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	urlString := c.groupUrl(groupEmail)
	options = append(append(c.config.FetchOptions, options...), fetch_config.WithMethod(http.MethodPatch))
	_, patchedGroupSettings, err := motmedelHttpUtils.FetchJsonWithBody[*group.Group](ctx, urlString, groupSettings, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json with body: %w", err), urlString)
	}

	return patchedGroupSettings, nil
}
