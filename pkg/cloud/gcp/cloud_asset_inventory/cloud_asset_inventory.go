package cloud_asset_inventory

import (
	"context"
	"fmt"
	"net/url"

	"github.com/Motmedel/utils_go/pkg/cloud/gcp/cloud_asset_inventory/cloud_asset_inventory_config"
	"github.com/Motmedel/utils_go/pkg/cloud/gcp/cloud_asset_inventory/types/asset_list"
	"github.com/Motmedel/utils_go/pkg/cloud/gcp/cloud_asset_inventory/types/resource_search_result_list"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"github.com/Motmedel/utils_go/pkg/errors/types/empty_error"
	"github.com/Motmedel/utils_go/pkg/http/types/fetch_config"
	motmedelHttpUtils "github.com/Motmedel/utils_go/pkg/http/utils"
)

const Domain = "cloudasset.googleapis.com"

var defaultBaseUrl = &url.URL{
	Scheme: "https",
	Host:   Domain,
}

type Client struct {
	baseUrl *url.URL
	config  *cloud_asset_inventory_config.Config
}

func NewClient(options ...cloud_asset_inventory_config.Option) *Client {
	config := cloud_asset_inventory_config.New(options...)
	baseUrl := config.BaseUrl
	if baseUrl == nil {
		baseUrl = defaultBaseUrl
	}
	u := *baseUrl
	u.Path = "/v1/"

	return &Client{baseUrl: &u, config: config}
}

// ListAssets lists assets under the specified parent (e.g. "organizations/123456", "projects/my-project", or "folders/123456").
// Use the query parameter to specify assetTypes, contentType, pageSize, pageToken, and other query parameters.
func (c *Client) ListAssets(ctx context.Context, parent string, query url.Values, options ...fetch_config.Option) (*asset_list.AssetList, error) {
	if parent == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("parent"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	u := *c.baseUrl
	u.Path += parent + "/assets"
	if query != nil {
		u.RawQuery = query.Encode()
	}
	urlString := u.String()

	options = append(c.config.FetchOptions, options...)
	_, list, err := motmedelHttpUtils.FetchJson[*asset_list.AssetList](ctx, urlString, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json: %w", err), urlString)
	}

	return list, nil
}

// SearchAllResources searches all resources within the specified scope (e.g. "organizations/123456", "projects/my-project", or "folders/123456").
// Use the query parameter to specify query, assetTypes, pageSize, pageToken, orderBy, readMask, and other query parameters.
func (c *Client) SearchAllResources(ctx context.Context, scope string, query url.Values, options ...fetch_config.Option) (*resource_search_result_list.ResourceSearchResultList, error) {
	if scope == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("scope"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	u := *c.baseUrl
	u.Path += scope + ":searchAllResources"
	if query != nil {
		u.RawQuery = query.Encode()
	}
	urlString := u.String()

	options = append(c.config.FetchOptions, options...)
	_, list, err := motmedelHttpUtils.FetchJson[*resource_search_result_list.ResourceSearchResultList](ctx, urlString, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json: %w", err), urlString)
	}

	return list, nil
}
