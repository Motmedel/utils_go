package gcp

import (
	"context"
	"fmt"
	"net/url"

	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"github.com/Motmedel/utils_go/pkg/errors/types/empty_error"
	motmedelHttpUtils "github.com/Motmedel/utils_go/pkg/http/utils"

	"github.com/Motmedel/utils_go/pkg/http/types/fetch_config"
)

var metadataBaseUrl = &url.URL{
	Scheme: "http",
	Host:   "metadata.google.internal",
	Path:   "/computeMetadata/v1",
}

func GetIdToken(ctx context.Context, audience string) (string, error) {
	if audience == "" {
		return "", motmedelErrors.NewWithTrace(empty_error.New("audience"))
	}

	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("context err: %w", err)
	}

	identityUrl := *metadataBaseUrl
	identityUrl.Path += "/instance/service-accounts/default/identity"
	identityUrl.RawQuery = url.Values{"audience": {audience}}.Encode()

	identityUrlString := identityUrl.String()
	_, responseBody, err := motmedelHttpUtils.Fetch(
		ctx,
		identityUrlString,
		fetch_config.WithHeaders(map[string]string{"Metadata-Flavor": "Google"}),
	)
	if err != nil {
		return "", motmedelErrors.New(fmt.Errorf("fetch: %w", err), identityUrlString)
	}

	return string(responseBody), nil
}

func GetProjectId(ctx context.Context) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("context err: %w", err)
	}

	requestUrl := *metadataBaseUrl
	requestUrl.Path += "/project/project-id"

	urlString := requestUrl.String()
	_, responseBody, err := motmedelHttpUtils.Fetch(
		ctx,
		urlString,
		fetch_config.WithHeaders(map[string]string{"Metadata-Flavor": "Google"}),
	)
	if err != nil {
		return "", motmedelErrors.New(fmt.Errorf("fetch: %w", err), urlString)
	}

	return string(responseBody), nil
}
