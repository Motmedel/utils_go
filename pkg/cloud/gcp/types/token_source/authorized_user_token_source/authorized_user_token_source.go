package authorized_user_token_source

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Motmedel/utils_go/pkg/cloud/gcp/types/credentials_file"
	"github.com/Motmedel/utils_go/pkg/cloud/gcp/types/token_response"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"github.com/Motmedel/utils_go/pkg/errors/types/empty_error"
	"github.com/Motmedel/utils_go/pkg/errors/types/nil_error"
	"github.com/Motmedel/utils_go/pkg/http/types/fetch_config"
	"github.com/Motmedel/utils_go/pkg/http/utils"
	"github.com/Motmedel/utils_go/pkg/oauth2/types/token"
)

type TokenSource struct {
	ctx          context.Context
	clientID     string
	clientSecret string
	refreshToken string
	tokenUrl     string
	options      []fetch_config.Option

	credentialsFile *credentials_file.CredentialsFile
}

func (ts *TokenSource) Token() (*token.Token, error) {
	if err := ts.ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	v := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {ts.clientID},
		"client_secret": {ts.clientSecret},
		"refresh_token": {ts.refreshToken},
	}

	options := append(
		[]fetch_config.Option{
			fetch_config.WithMethod(http.MethodPost),
			fetch_config.WithHeaders(map[string]string{
				"Content-Type": "application/x-www-form-urlencoded",
			}),
			fetch_config.WithBody([]byte(v.Encode())),
		},
		ts.options...,
	)

	_, tokenResponse, err := utils.FetchJson[*token_response.Response](ts.ctx, ts.tokenUrl, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json: %w", err), ts.tokenUrl)
	}
	if tokenResponse == nil {
		return nil, motmedelErrors.NewWithTrace(nil_error.New("token response"))
	}

	return tokenResponse.Token(), nil
}

func (ts *TokenSource) CredentialsFile() *credentials_file.CredentialsFile {
	return ts.credentialsFile
}

func NewFromCredentialsFile(
	ctx context.Context,
	tokenUrl string,
	credentialsFile *credentials_file.CredentialsFile,
	options ...fetch_config.Option,
) (*TokenSource, error) {
	if tokenUrl == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("token url"))
	}

	if credentialsFile == nil {
		return nil, nil
	}

	return &TokenSource{
		ctx:          ctx,
		clientID:     credentialsFile.ClientID,
		clientSecret: credentialsFile.ClientSecret,
		refreshToken: credentialsFile.RefreshToken,
		tokenUrl:     tokenUrl,
		options:      options,
	}, nil
}
