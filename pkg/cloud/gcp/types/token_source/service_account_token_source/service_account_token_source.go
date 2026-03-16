package service_account_token_source

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
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Motmedel/utils_go/pkg/cloud/gcp/types/credentials_file"
	"github.com/Motmedel/utils_go/pkg/cloud/gcp/types/token_response"
	"github.com/Motmedel/utils_go/pkg/errors"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"github.com/Motmedel/utils_go/pkg/errors/types/nil_error"
	"github.com/Motmedel/utils_go/pkg/http/types/fetch_config"
	"github.com/Motmedel/utils_go/pkg/http/utils"
	"github.com/Motmedel/utils_go/pkg/oauth2/types/token"
)

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

type TokenSource struct {
	ctx          context.Context
	clientEmail  string
	privateKeyID string
	privateKey   *rsa.PrivateKey
	tokenURI     string
	scopes       []string
	options      []fetch_config.Option

	credentialsFile *credentials_file.CredentialsFile
}

func (s *TokenSource) Token() (*token.Token, error) {
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
		return nil, errors.NewWithTrace(fmt.Errorf("json marshal (jwt header): %w", err))
	}

	claimsJSON, err := json.Marshal(map[string]any{
		"iss":   s.clientEmail,
		"scope": strings.Join(s.scopes, " "),
		"aud":   s.tokenURI,
		"iat":   now.Unix(),
		"exp":   now.Add(time.Hour).Unix(),
	})
	if err != nil {
		return nil, errors.NewWithTrace(fmt.Errorf("json marshal (jwt claims): %w", err))
	}

	signingInput := base64.RawURLEncoding.EncodeToString(headerJSON) +
		"." +
		base64.RawURLEncoding.EncodeToString(claimsJSON)

	h := sha256.Sum256([]byte(signingInput))
	signature, err := rsa.SignPKCS1v15(rand.Reader, s.privateKey, crypto.SHA256, h[:])
	if err != nil {
		return nil, errors.NewWithTrace(fmt.Errorf("rsa sign pkcs1v15: %w", err))
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

	_, tokenResponse, err := utils.FetchJson[*token_response.Response](s.ctx, s.tokenURI, options...)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("fetch json: %w", err), s.tokenURI)
	}
	if tokenResponse == nil {
		return nil, motmedelErrors.NewWithTrace(nil_error.New("token response"))
	}

	return tokenResponse.Token(), nil
}

func (s *TokenSource) CredentialsFile() *credentials_file.CredentialsFile {
	return s.credentialsFile
}

func NewFromCredentialsFile(
	ctx context.Context,
	tokenUrl string,
	credentialsFile *credentials_file.CredentialsFile,
	scopes []string,
	options ...fetch_config.Option,
) (*TokenSource, error) {
	if tokenUrl == "" {
		return nil, motmedelErrors.NewWithTrace(errors.New("token url"))
	}

	if credentialsFile == nil {
		return nil, nil
	}

	rsaKey, err := parsePrivateKey(credentialsFile.PrivateKey)
	if err != nil {
		return nil, motmedelErrors.NewWithTrace(fmt.Errorf("parse private key: %w", err))
	}

	return &TokenSource{
		ctx:          ctx,
		clientEmail:  credentialsFile.ClientEmail,
		privateKeyID: credentialsFile.PrivateKeyID,
		privateKey:   rsaKey,
		tokenURI:     tokenUrl,
		scopes:       scopes,
		options:      options,
	}, nil
}
