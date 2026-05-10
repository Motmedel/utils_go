package iam_credentials

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Motmedel/utils_go/pkg/cloud/gcp/iam_credentials/iam_credentials_config"
	"github.com/Motmedel/utils_go/pkg/cloud/gcp/iam_credentials/types/sign_blob_request"
	"github.com/Motmedel/utils_go/pkg/cloud/gcp/iam_credentials/types/sign_blob_response"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"github.com/Motmedel/utils_go/pkg/errors/types/empty_error"
	"github.com/Motmedel/utils_go/pkg/errors/types/nil_error"
	"github.com/Motmedel/utils_go/pkg/http/types/fetch_config"
	motmedelHttpUtils "github.com/Motmedel/utils_go/pkg/http/utils"
)

const Domain = "iamcredentials.googleapis.com"

var defaultBaseUrl = &url.URL{
	Scheme: "https",
	Host:   Domain,
}

type Client struct {
	baseUrl *url.URL
	config  *iam_credentials_config.Config
}

func NewClient(options ...iam_credentials_config.Option) *Client {
	config := iam_credentials_config.New(options...)
	baseUrl := config.BaseUrl
	if baseUrl == nil {
		baseUrl = defaultBaseUrl
	}
	u := *baseUrl
	u.Path = "/v1/"

	return &Client{baseUrl: &u, config: config}
}

// SignBlob signs the given payload bytes on behalf of the specified service account using the
// IAM Credentials API. The runtime identity must have the iam.serviceAccountTokenCreator role
// on the target service account (which may be itself).
func (c *Client) SignBlob(ctx context.Context, serviceAccountEmail string, payload []byte, options ...fetch_config.Option) (*sign_blob_response.Response, error) {
	if serviceAccountEmail == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("service account email"))
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	// gRPC transcoding requires the slashes in the resource name to remain literal —
	// only the email itself is per-request user input that needs path-segment escaping.
	u := *c.baseUrl
	u.RawPath = u.Path + "projects/-/serviceAccounts/" + url.PathEscape(serviceAccountEmail) + ":signBlob"
	u.Path += "projects/-/serviceAccounts/" + serviceAccountEmail + ":signBlob"
	urlString := u.String()

	body := &sign_blob_request.Request{Payload: base64.StdEncoding.EncodeToString(payload)}

	options = append(append(c.config.FetchOptions, options...), fetch_config.WithMethod(http.MethodPost))
	_, response, err := motmedelHttpUtils.FetchJsonWithBody[*sign_blob_response.Response](ctx, urlString, body, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json with body: %w", err), urlString)
	}
	if response == nil {
		return nil, motmedelErrors.NewWithTrace(nil_error.New("sign blob response"))
	}

	return response, nil
}
