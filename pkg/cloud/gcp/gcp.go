package gcp

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"github.com/Motmedel/utils_go/pkg/errors/types/empty_error"
	"github.com/Motmedel/utils_go/pkg/http/types/fetch_config"
	motmedelHttpUtils "github.com/Motmedel/utils_go/pkg/http/utils"
	"github.com/Motmedel/utils_go/pkg/oauth2/types/token"
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

	ErrNoDefaultCredentials = errors.New("could not find default credentials")
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

type credentialsFile struct {
	Type string `json:"type"`

	// authorized_user fields
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RefreshToken string `json:"refresh_token"`

	// service_account fields
	ClientEmail  string `json:"client_email"`
	PrivateKeyID string `json:"private_key_id"`
	PrivateKey   string `json:"private_key"`
	TokenURI     string `json:"token_uri"`
	ProjectID    string `json:"project_id"`
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	TokenType   string `json:"token_type"`
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

type authorizedUserTokenSource struct {
	ctx          context.Context
	clientID     string
	clientSecret string
	refreshToken string
	tokenUrl     string
	options      []fetch_config.Option
}

func (s *authorizedUserTokenSource) Token() (*token.Token, error) {
	if err := s.ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	v := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {s.clientID},
		"client_secret": {s.clientSecret},
		"refresh_token": {s.refreshToken},
	}

	options := append(
		[]fetch_config.Option{
			fetch_config.WithMethod(http.MethodPost),
			fetch_config.WithHeaders(map[string]string{
				"Content-Type": "application/x-www-form-urlencoded",
			}),
			fetch_config.WithBody([]byte(v.Encode())),
		},
		s.options...,
	)

	_, responseBody, err := motmedelHttpUtils.Fetch(s.ctx, s.tokenUrl, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch: %w", err), s.tokenUrl)
	}

	return parseTokenResponse(responseBody)
}

type serviceAccountTokenSource struct {
	ctx          context.Context
	clientEmail  string
	privateKeyID string
	privateKey   *rsa.PrivateKey
	tokenURI     string
	scopes       []string
	options      []fetch_config.Option
}

func (s *serviceAccountTokenSource) Token() (*token.Token, error) {
	if err := s.ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	now := time.Now()

	headerJSON, err := json.Marshal(map[string]string{
		"alg": "RS256",
		"typ": "JWT",
		"kid": s.privateKeyID,
	})
	if err != nil {
		return nil, motmedelErrors.NewWithTrace(fmt.Errorf("json marshal (jwt header): %w", err))
	}

	claimsJSON, err := json.Marshal(map[string]any{
		"iss":   s.clientEmail,
		"scope": strings.Join(s.scopes, " "),
		"aud":   s.tokenURI,
		"iat":   now.Unix(),
		"exp":   now.Add(time.Hour).Unix(),
	})
	if err != nil {
		return nil, motmedelErrors.NewWithTrace(fmt.Errorf("json marshal (jwt claims): %w", err))
	}

	signingInput := base64.RawURLEncoding.EncodeToString(headerJSON) +
		"." +
		base64.RawURLEncoding.EncodeToString(claimsJSON)

	h := sha256.Sum256([]byte(signingInput))
	signature, err := rsa.SignPKCS1v15(rand.Reader, s.privateKey, crypto.SHA256, h[:])
	if err != nil {
		return nil, motmedelErrors.NewWithTrace(fmt.Errorf("rsa sign pkcs1v15: %w", err))
	}

	assertion := signingInput + "." + base64.RawURLEncoding.EncodeToString(signature)

	v := url.Values{
		"grant_type": {"urn:ietf:params:oauth:grant-type:jwt-bearer"},
		"assertion":  {assertion},
	}

	options := append(
		[]fetch_config.Option{
			fetch_config.WithMethod(http.MethodPost),
			fetch_config.WithHeaders(map[string]string{
				"Content-Type": "application/x-www-form-urlencoded",
			}),
			fetch_config.WithBody([]byte(v.Encode())),
		},
		s.options...,
	)

	_, responseBody, err := motmedelHttpUtils.Fetch(s.ctx, s.tokenURI, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch: %w", err), s.tokenURI)
	}

	return parseTokenResponse(responseBody)
}

type metadataTokenSource struct {
	ctx             context.Context
	metadataBaseUrl *url.URL
	scopes          []string
	options         []fetch_config.Option
}

func (s *metadataTokenSource) Token() (*token.Token, error) {
	if err := s.ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	tokenUrl := *s.metadataBaseUrl
	tokenUrl.Path += "/instance/service-accounts/default/token"
	if len(s.scopes) > 0 {
		tokenUrl.RawQuery = url.Values{"scopes": {strings.Join(s.scopes, ",")}}.Encode()
	}

	urlString := tokenUrl.String()

	options := append(
		[]fetch_config.Option{
			fetch_config.WithHeaders(map[string]string{"Metadata-Flavor": "Google"}),
		},
		s.options...,
	)

	_, responseBody, err := motmedelHttpUtils.Fetch(s.ctx, urlString, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch: %w", err), urlString)
	}

	return parseTokenResponse(responseBody)
}

func parseTokenResponse(data []byte) (*token.Token, error) {
	var tr tokenResponse
	if err := json.Unmarshal(data, &tr); err != nil {
		return nil, motmedelErrors.NewWithTrace(
			fmt.Errorf("json unmarshal (token response): %w", err),
			data,
		)
	}

	if tr.AccessToken == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("access token"))
	}

	tok := &token.Token{
		AccessToken: tr.AccessToken,
		TokenType:   tr.TokenType,
	}
	if tr.ExpiresIn > 0 {
		tok.Expiry = time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second)
	}

	return tok, nil
}

func wellKnownCredentialsPath() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("APPDATA"), "gcloud", "application_default_credentials.json")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	return filepath.Join(home, ".config", "gcloud", "application_default_credentials.json")
}

func parsePrivateKey(pemData string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, motmedelErrors.NewWithTrace(errors.New("pem decode: no PEM block found"))
	}

	switch block.Type {
	case "RSA PRIVATE KEY":
		rsaKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, motmedelErrors.NewWithTrace(fmt.Errorf("x509 parse pkcs1 private key: %w", err))
		}
		return rsaKey, nil
	case "PRIVATE KEY":
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, motmedelErrors.NewWithTrace(fmt.Errorf("x509 parse pkcs8 private key: %w", err))
		}
		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, motmedelErrors.NewWithTrace(fmt.Errorf("unexpected key type: %T", key))
		}
		return rsaKey, nil
	default:
		return nil, motmedelErrors.NewWithTrace(fmt.Errorf("unsupported PEM block type: %s", block.Type))
	}
}

func (c *Client) credentialsFileTokenSource(ctx context.Context, data []byte, scopes []string, options ...fetch_config.Option) (token_source.TokenSource, error) {
	var cf credentialsFile
	if err := json.Unmarshal(data, &cf); err != nil {
		return nil, motmedelErrors.NewWithTrace(
			fmt.Errorf("json unmarshal (credentials file): %w", err),
		)
	}

	switch cf.Type {
	case credentialTypeAuthorizedUser:
		ts := &authorizedUserTokenSource{
			ctx:          ctx,
			clientID:     cf.ClientID,
			clientSecret: cf.ClientSecret,
			refreshToken: cf.RefreshToken,
			tokenUrl:     c.tokenUrl,
			options:      options,
		}
		return token_source.NewReusable(nil, ts), nil

	case credentialTypeServiceAccount:
		tokenURI := cf.TokenURI
		if tokenURI == "" {
			tokenURI = c.tokenUrl
		}

		rsaKey, err := parsePrivateKey(cf.PrivateKey)
		if err != nil {
			return nil, motmedelErrors.NewWithTrace(fmt.Errorf("parse private key: %w", err))
		}

		ts := &serviceAccountTokenSource{
			ctx:          ctx,
			clientEmail:  cf.ClientEmail,
			privateKeyID: cf.PrivateKeyID,
			privateKey:   rsaKey,
			tokenURI:     tokenURI,
			scopes:       scopes,
			options:      options,
		}
		return token_source.NewReusable(nil, ts), nil

	default:
		return nil, motmedelErrors.NewWithTrace(
			fmt.Errorf("unsupported credential type: %s", cf.Type),
		)
	}
}

// FindDefaultCredentials discovers Google credentials using the Application Default Credentials chain:
//  1. GOOGLE_APPLICATION_CREDENTIALS env var — reads the JSON file it points to.
//  2. Well-known file — user credentials from gcloud auth application-default login.
//  3. Metadata server — if running on GCP (Compute Engine, Cloud Run, etc.).
//  4. Error — if none of the above succeed.
func (c *Client) FindDefaultCredentials(ctx context.Context, scopes []string, fetchOptions ...fetch_config.Option) (token_source.TokenSource, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	// 1. GOOGLE_APPLICATION_CREDENTIALS env var.
	if envPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); envPath != "" {
		data, err := os.ReadFile(envPath)
		if err != nil {
			return nil, motmedelErrors.NewWithTrace(
				fmt.Errorf("read file (%s): %w", envPath, err),
			)
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
	metadataSource := &metadataTokenSource{
		ctx:             ctx,
		metadataBaseUrl: c.metadataBaseUrl,
		scopes:          scopes,
		options:         fetchOptions,
	}
	if tok, err := metadataSource.Token(); err == nil {
		return token_source.NewReusable(tok, metadataSource), nil
	}

	// 4. Error.
	return nil, motmedelErrors.NewWithTrace(ErrNoDefaultCredentials)
}
