package oauth2

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"github.com/Motmedel/utils_go/pkg/errors/types/empty_error"
	"github.com/Motmedel/utils_go/pkg/http/types/fetch_config"
	motmedelHttpUtils "github.com/Motmedel/utils_go/pkg/http/utils"
	oauth2Errors "github.com/Motmedel/utils_go/pkg/oauth2/errors"
	"github.com/Motmedel/utils_go/pkg/oauth2/types/token"
)

type AuthStyle int

const (
	AuthStyleAutoDetect AuthStyle = iota
	AuthStyleInParams
	AuthStyleInHeader
)

type Endpoint struct {
	AuthURL   string
	TokenURL  string
	AuthStyle AuthStyle
}

type AuthCodeOption struct {
	key   string
	value string
}

func SetAuthURLParam(key, value string) AuthCodeOption {
	return AuthCodeOption{key: key, value: value}
}

type Config struct {
	ClientID     string
	ClientSecret string
	Endpoint     Endpoint
	RedirectURL  string
	Scopes       []string

	FetchOptions []fetch_config.Option
}

func (c *Config) AuthCodeURL(state string, opts ...AuthCodeOption) string {
	var buf strings.Builder
	buf.WriteString(c.Endpoint.AuthURL)

	v := url.Values{
		"response_type": {"code"},
		"client_id":     {c.ClientID},
	}
	if c.RedirectURL != "" {
		v.Set("redirect_uri", c.RedirectURL)
	}
	if len(c.Scopes) > 0 {
		v.Set("scope", strings.Join(c.Scopes, " "))
	}
	if state != "" {
		v.Set("state", state)
	}
	for _, opt := range opts {
		v.Set(opt.key, opt.value)
	}

	if strings.Contains(c.Endpoint.AuthURL, "?") {
		buf.WriteByte('&')
	} else {
		buf.WriteByte('?')
	}
	buf.WriteString(v.Encode())

	return buf.String()
}

func (c *Config) Exchange(ctx context.Context, code string, opts ...AuthCodeOption) (*token.Token, error) {
	v := url.Values{
		"grant_type": {"authorization_code"},
		"code":       {code},
	}
	if c.RedirectURL != "" {
		v.Set("redirect_uri", c.RedirectURL)
	}
	for _, opt := range opts {
		v.Set(opt.key, opt.value)
	}

	return c.retrieveToken(ctx, v)
}

func (c *Config) PasswordCredentialsToken(ctx context.Context, username, password string) (*token.Token, error) {
	v := url.Values{
		"grant_type": {"password"},
		"username":   {username},
		"password":   {password},
	}
	if len(c.Scopes) > 0 {
		v.Set("scope", strings.Join(c.Scopes, " "))
	}

	return c.retrieveToken(ctx, v)
}

func (c *Config) ClientCredentialsToken(ctx context.Context) (*token.Token, error) {
	v := url.Values{
		"grant_type": {"client_credentials"},
	}
	if len(c.Scopes) > 0 {
		v.Set("scope", strings.Join(c.Scopes, " "))
	}

	return c.retrieveToken(ctx, v)
}

func (c *Config) TokenSource(ctx context.Context, t *token.Token) TokenSource {
	tkr := &tokenRefresher{
		ctx:  ctx,
		conf: c,
	}
	if t != nil {
		tkr.refreshToken = t.RefreshToken
	}

	return &reuseTokenSource{
		t:   t,
		new: tkr,
	}
}

func (c *Config) retrieveToken(ctx context.Context, v url.Values) (*token.Token, error) {
	if c.Endpoint.TokenURL == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("token url"))
	}

	authStyle := c.Endpoint.AuthStyle

	if authStyle == AuthStyleAutoDetect {
		// Try header first, fall back to params.
		tok, err := c.doRetrieveToken(ctx, v, AuthStyleInHeader)
		if err == nil {
			return tok, nil
		}
		tok, err = c.doRetrieveToken(ctx, v, AuthStyleInParams)
		if err == nil {
			return tok, nil
		}
		return nil, fmt.Errorf("do retrieve token: %w", err)
	}

	return c.doRetrieveToken(ctx, v, authStyle)
}

func (c *Config) doRetrieveToken(ctx context.Context, v url.Values, authStyle AuthStyle) (*token.Token, error) {
	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}

	switch authStyle {
	case AuthStyleInHeader:
		headers["Authorization"] = "Basic " + motmedelHttpUtils.BasicAuth(c.ClientID, c.ClientSecret)
	case AuthStyleInParams:
		v.Set("client_id", c.ClientID)
		if c.ClientSecret != "" {
			v.Set("client_secret", c.ClientSecret)
		}
	}

	body := []byte(v.Encode())

	options := append(
		[]fetch_config.Option{
			fetch_config.WithMethod("POST"),
			fetch_config.WithHeaders(headers),
			fetch_config.WithBody(body),
			fetch_config.WithSkipErrorOnStatus(true),
		},
		c.FetchOptions...,
	)

	response, responseBody, err := motmedelHttpUtils.Fetch(ctx, c.Endpoint.TokenURL, options...)
	if err != nil {
		return nil, fmt.Errorf("fetch: %w", err)
	}

	if response != nil && (response.StatusCode < 200 || response.StatusCode >= 300) {
		retrieveErr := &oauth2Errors.RetrieveError{
			StatusCode: response.StatusCode,
			Body:       responseBody,
		}

		var errorResponse struct {
			Error            string `json:"error"`
			ErrorDescription string `json:"error_description"`
			ErrorURI         string `json:"error_uri"`
		}
		if err := json.Unmarshal(responseBody, &errorResponse); err == nil {
			retrieveErr.ErrorCode = errorResponse.Error
			retrieveErr.ErrorDescription = errorResponse.ErrorDescription
			retrieveErr.ErrorURI = errorResponse.ErrorURI
		}

		return nil, retrieveErr
	}

	var tokenResponse struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
	}

	contentType := ""
	if response != nil {
		contentType = response.Header.Get("Content-Type")
	}

	if strings.Contains(contentType, "application/x-www-form-urlencoded") || strings.Contains(contentType, "text/plain") {
		vals, err := url.ParseQuery(string(responseBody))
		if err != nil {
			return nil, motmedelErrors.NewWithTrace(fmt.Errorf("url parse query (response body): %w", err), responseBody)
		}

		tokenResponse.AccessToken = vals.Get("access_token")
		tokenResponse.TokenType = vals.Get("token_type")
		tokenResponse.RefreshToken = vals.Get("refresh_token")
	} else {
		if err := json.Unmarshal(responseBody, &tokenResponse); err != nil {
			return nil, motmedelErrors.NewWithTrace(fmt.Errorf("json unmarshal (response body): %w", err), responseBody)
		}
	}

	tok := &token.Token{
		AccessToken:  tokenResponse.AccessToken,
		TokenType:    tokenResponse.TokenType,
		RefreshToken: tokenResponse.RefreshToken,
	}

	if tokenResponse.ExpiresIn > 0 {
		tok.Expiry = time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second)
	}

	var raw map[string]any
	if err := json.Unmarshal(responseBody, &raw); err == nil {
		tok.Raw = raw
	}

	if tok.AccessToken == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("access token"))
	}

	return tok, nil
}

// TokenSource is a source of tokens.
type TokenSource interface {
	Token() (*token.Token, error)
}

type tokenRefresher struct {
	ctx          context.Context
	conf         *Config
	refreshToken string
}

func (tf *tokenRefresher) Token() (*token.Token, error) {
	if tf.refreshToken == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("refresh token"))
	}

	v := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {tf.refreshToken},
	}
	if len(tf.conf.Scopes) > 0 {
		v.Set("scope", strings.Join(tf.conf.Scopes, " "))
	}

	tok, err := tf.conf.retrieveToken(tf.ctx, v)
	if err != nil {
		return nil, fmt.Errorf("retrieve token: %w", err)
	}

	if tok.RefreshToken == "" {
		tok.RefreshToken = tf.refreshToken
	} else {
		tf.refreshToken = tok.RefreshToken
	}

	return tok, nil
}

type reuseTokenSource struct {
	new TokenSource
	mu  sync.Mutex
	t   *token.Token
}

func (s *reuseTokenSource) Token() (*token.Token, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.t.Valid() {
		return s.t, nil
	}

	t, err := s.new.Token()
	if err != nil {
		return nil, err
	}

	s.t = t
	return t, nil
}

// StaticTokenSource returns a TokenSource that always returns the same token.
func StaticTokenSource(t *token.Token) TokenSource {
	return &staticTokenSource{t: t}
}

type staticTokenSource struct {
	t *token.Token
}

func (s *staticTokenSource) Token() (*token.Token, error) {
	return s.t, nil
}

// ReuseTokenSource returns a TokenSource that caches the token from src
// and refreshes it when expired.
func ReuseTokenSource(t *token.Token, src TokenSource) TokenSource {
	return &reuseTokenSource{
		t:   t,
		new: src,
	}
}
