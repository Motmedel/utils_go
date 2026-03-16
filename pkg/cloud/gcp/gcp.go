package gcp

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"

	"github.com/Motmedel/utils_go/pkg/cloud/gcp/types/credentials_file"
	"github.com/Motmedel/utils_go/pkg/cloud/gcp/types/token_source/authorized_user_token_source"
	"github.com/Motmedel/utils_go/pkg/cloud/gcp/types/token_source/metadata_token_source"
	"github.com/Motmedel/utils_go/pkg/cloud/gcp/types/token_source/service_account_token_source"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"github.com/Motmedel/utils_go/pkg/errors/types/empty_error"
	"github.com/Motmedel/utils_go/pkg/http/types/fetch_config"
	motmedelHttpUtils "github.com/Motmedel/utils_go/pkg/http/utils"
	"github.com/Motmedel/utils_go/pkg/oauth2/types/token_source"
)

const (
	credentialTypeAuthorizedUser = "authorized_user"
	credentialTypeServiceAccount = "service_account"

	DefaultTokenUrl = "https://oauth2.googleapis.com/token"
)

var (
	defaultMetadataBaseUrl = &url.URL{
		Scheme: "http",
		Host:   "metadata.google.internal",
		Path:   "/computeMetadata/v1",
	}
)

type Client struct {
	metadataBaseUrl *url.URL
	tokenUrl        string
}

func NewClient() *Client {
	return NewClientWithUrls(defaultMetadataBaseUrl, DefaultTokenUrl)
}

func NewClientWithUrls(metadataBaseUrl *url.URL, tokenUrl string) *Client {
	return &Client{
		metadataBaseUrl: metadataBaseUrl,
		tokenUrl:        tokenUrl,
	}
}

func (c *Client) GetIdToken(ctx context.Context, audience string) (string, error) {
	if audience == "" {
		return "", motmedelErrors.NewWithTrace(empty_error.New("audience"))
	}

	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("context err: %w", err)
	}

	identityUrl := *c.metadataBaseUrl
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

func (c *Client) GetProjectId(ctx context.Context) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("context err: %w", err)
	}

	requestUrl := *c.metadataBaseUrl
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

func wellKnownCredentialsPath() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("APPDATA"), "gcloud", "application_default_credentials.json")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	// TODO: Use XDG?

	return filepath.Join(home, ".config", "gcloud", "application_default_credentials.json")
}

func (c *Client) credentialsFileTokenSource(ctx context.Context, data []byte, scopes []string, options ...fetch_config.Option) (token_source.TokenSource, error) {
	var credentialsFile credentials_file.CredentialsFile
	if err := json.Unmarshal(data, &credentialsFile); err != nil {
		return nil, motmedelErrors.NewWithTrace(
			fmt.Errorf("json unmarshal (credentials file): %w", err),
		)
	}

	switch credentialsFile.Type {
	case credentialTypeAuthorizedUser:
		tokenSource, err := authorized_user_token_source.NewFromCredentialsFile(ctx, c.tokenUrl, &credentialsFile, options...)
		if err != nil {
			return nil, motmedelErrors.NewWithTrace(fmt.Errorf("authorized user token source new: %w", err), credentialsFile)
		}

		return token_source.NewReusable(nil, tokenSource), nil
	case credentialTypeServiceAccount:
		tokenURI := credentialsFile.TokenURI
		if tokenURI == "" {
			tokenURI = c.tokenUrl
		}

		tokenSource, err := service_account_token_source.NewFromCredentialsFile(ctx, tokenURI, &credentialsFile, scopes, options...)
		if err != nil {
			return nil, motmedelErrors.NewWithTrace(fmt.Errorf("service account token source new: %w", err), credentialsFile)
		}

		return token_source.NewReusable(nil, tokenSource), nil
	default:
		return nil, motmedelErrors.NewWithTrace(
			fmt.Errorf("unsupported credential type: %s", credentialsFile.Type),
		)
	}
}

// FindDefaultCredentials discovers Google credentials using the Application Default Credentials chain:
//  1. GOOGLE_APPLICATION_CREDENTIALS env var — reads the JSON file it points to.
//  2. Well-known file — user credentials from gcloud auth application-default login.
//  3. Metadata server — if running on GCP (Compute Engine, Cloud Run, etc.).
func (c *Client) FindDefaultCredentials(ctx context.Context, scopes []string, fetchOptions ...fetch_config.Option) (token_source.TokenSource, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	// 1. GOOGLE_APPLICATION_CREDENTIALS env var.
	if envPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); envPath != "" {
		data, err := os.ReadFile(envPath)
		if err != nil {
			return nil, motmedelErrors.NewWithTrace(fmt.Errorf("os read file: %w", err), envPath)
		}
		return c.credentialsFileTokenSource(ctx, data, scopes, fetchOptions...)
	}

	// 2. Well-known file.
	if wellKnownPath := wellKnownCredentialsPath(); wellKnownPath != "" {
		if data, err := os.ReadFile(wellKnownPath); err == nil {
			return c.credentialsFileTokenSource(ctx, data, scopes, fetchOptions...)
		}
	}

	// 3. Metadata server.
	if metadataBaseUrl := c.metadataBaseUrl; metadataBaseUrl != nil {
		metadataTokenSource, err := metadata_token_source.New(ctx, c.metadataBaseUrl, scopes, fetchOptions...)
		if err != nil {
			return nil, motmedelErrors.NewWithTrace(fmt.Errorf("metadata token source new: %w", err))
		}
		return token_source.NewReusable(nil, metadataTokenSource), nil
	}

	return nil, nil
}
