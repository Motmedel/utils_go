package types

import (
	"crypto"
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	motmedelCrypto "github.com/Motmedel/utils_go/pkg/crypto"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"github.com/Motmedel/utils_go/pkg/errors/types/empty_error"
)

const (
	DkimPrefix  = "v=DKIM1"
	DmarcPrefix = "v=DMARC1"
	SpfPrefix   = "v=spf1 "
)

const SpfMaximumLookupLimit = 10

const (
	SpfNeutralQualifier  = "?"
	SpfSoftfailQualifier = "~"
	SpfFailQualifier     = "-"
)

var ErrBadEd25519Length = errors.New("bad ed25519 length")

type DkimRecord struct {
	Version                  int      `json:"version,omitzero"`
	AcceptableHashAlgorithms []string `json:"acceptable_hash_algorithms,omitzero"`
	KeyType                  string   `json:"key_type,omitzero"`
	Notes                    string   `json:"notes,omitzero"`
	PublicKeyData            string   `json:"public_key_data,omitzero"`
	ServiceType              string   `json:"service_type,omitzero"`
	Flags                    []string `json:"flags,omitzero"`

	Raw        string      `json:"raw,omitzero"`
	Domain     string      `json:"domain,omitzero"`
	Selector   string      `json:"selector,omitzero"`
	Extensions [][2]string `json:"extensions,omitzero"`
}

func (r *DkimRecord) GetVersion() int {
	if r.Version == 0 {
		return 1
	}

	return r.Version
}

func (r *DkimRecord) GetKeyType() string {
	if r.KeyType == "" {
		return "rsa"
	}

	return r.KeyType
}

func (r *DkimRecord) GetServiceType() string {
	if r.ServiceType == "" {
		return "*"
	}

	return r.ServiceType
}

func (r *DkimRecord) GetPublicKey() (crypto.PublicKey, error) {
	publicKeyData := r.PublicKeyData
	if len(publicKeyData) == 0 {
		return nil, nil
	}

	keyType := r.GetKeyType()
	key, err := ParseDkimKey(publicKeyData, keyType)
	if err != nil {
		return nil, motmedelErrors.New(
			fmt.Errorf("parse key: %w", err),
			publicKeyData, keyType,
		)
	}

	return key, nil
}

type DkimHeader struct {
	Version                 int         `json:"version,omitzero"`
	Algorithm               string      `json:"algorithm,omitzero"`
	Signature               string      `json:"signature,omitzero"`
	Hash                    string      `json:"hash,omitzero"`
	MessageCanonicalization string      `json:"message_canonicalization,omitzero"`
	SigningDomainIdentifier string      `json:"signing_domain_identifier,omitzero"`
	SignedHeaderFields      []string    `json:"signed_header_fields,omitzero"`
	AgentOrUserIdentifier   string      `json:"agent_or_user_identifier"`
	BodyLengthCount         string      `json:"body_length_count,omitzero"`
	QueryMethods            []string    `json:"query_methods,omitzero"`
	Selector                string      `json:"selector,omitzero"`
	SignatureTimestamp      string      `json:"signature_timestamp,omitzero"`
	SignatureExpiration     string      `json:"signature_expiration,omitzero"`
	CopiedHeaderFields      [][2]string `json:"copied_header_fields,omitzero"`

	Raw        string      `json:"raw,omitzero"`
	Domain     string      `json:"domain,omitzero"`
	Extensions [][2]string `json:"extensions,omitzero"`
}

func GetDkimKeyData(data string) ([]byte, error) {
	if data == "" {
		return nil, nil
	}

	keyData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, motmedelErrors.NewWithTrace(
			fmt.Errorf("base64 std encoding decode string: %w", err),
			data,
		)
	}

	return keyData, nil
}

func ParseDkimKey(data string, keyType string) (crypto.PublicKey, error) {
	if data == "" {
		return nil, nil
	}

	if keyType == "" {
		return nil, motmedelErrors.NewWithTrace(empty_error.New("key type"))
	}

	keyData, err := GetDkimKeyData(data)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("get key data: %w", err), data)
	}

	switch strings.ToLower(keyType) {
	case "rsa":
		key, err := motmedelCrypto.PublicKeyFromDer[crypto.PublicKey](keyData)
		if err != nil {
			return nil, fmt.Errorf("public key from der: %w", err)
		}

		return key, nil
	case "ed25519":
		if len(keyData) != ed25519.PublicKeySize {
			return nil, motmedelErrors.NewWithTrace(ErrBadEd25519Length, keyData)
		}

		return ed25519.PublicKey(keyData), nil
	default:
		return keyData, nil
	}
}

type DmarcRecord struct {
	Domain string `json:"domain,omitzero"`
	Raw    string `json:"raw,omitzero"`
	P      string `json:"p,omitzero"`
	Sp     string `json:"sp,omitzero"`
	Rua    string `json:"rua,omitzero"`
	Ruf    string `json:"ruf,omitzero"`
	Adkim  string `json:"adkim,omitzero"`
	Aspf   string `json:"aspf,omitzero"`
	Ri     string `json:"ri,omitzero"`
	Fo     string `json:"fo,omitzero"`
	Rf     string `json:"rf,omitzero"`
	Pct    string `json:"pct,omitzero"`
}

type SpfMechanism struct {
	Label string `json:"label,omitzero"`
	Value string `json:"value,omitzero"`
}

type SpfDirective struct {
	Index     int           `json:"index"`
	Qualifier string        `json:"qualifier,omitzero"`
	Mechanism *SpfMechanism `json:"mechanism,omitzero"`
}

type SpfModifier struct {
	Index int    `json:"index"`
	Label string `json:"label,omitzero"`
	Value string `json:"value,omitzero"`
}

type SpfTermPtr interface {
	*SpfModifier | *SpfDirective
}

func getTypedSpfTerms[T SpfTermPtr](record *SpfRecord) []T {
	var typedTerms []T

	for _, term := range record.Terms {
		switch typedTerm := term.(type) {
		case T:
			typedTerms = append(typedTerms, typedTerm)
		}
	}

	return typedTerms
}

type SpfRecord struct {
	Domain string `json:"domain,omitzero"`
	Raw    string `json:"raw,omitzero"`
	Terms  []any  `json:"-" jsonschema:"-"`
}

func (r *SpfRecord) Modifiers() []*SpfModifier {
	return getTypedSpfTerms[*SpfModifier](r)
}

func (r *SpfRecord) Directives() []*SpfDirective {
	return getTypedSpfTerms[*SpfDirective](r)
}
