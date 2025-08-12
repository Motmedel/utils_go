package jwt

import (
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/Motmedel/utils_go/pkg/crypto/ecdsa"
	motmedelCryptoErrors "github.com/Motmedel/utils_go/pkg/crypto/errors"
	motmedelCryptoInterfaces "github.com/Motmedel/utils_go/pkg/crypto/interfaces"
	"github.com/Motmedel/utils_go/pkg/crypto/rsa"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"github.com/Motmedel/utils_go/pkg/interfaces/validator"
	jwtErrors "github.com/Motmedel/utils_go/pkg/jwt/errors"
	"github.com/Motmedel/utils_go/pkg/jwt/parsing/types/raw_token"
	jwtKey "github.com/Motmedel/utils_go/pkg/jwt/types/key"
	jwtToken "github.com/Motmedel/utils_go/pkg/jwt/types/token"
	"github.com/Motmedel/utils_go/pkg/jwt/validation/types/base_validator"
	"github.com/Motmedel/utils_go/pkg/jwt/validation/types/header_validator"
	"github.com/Motmedel/utils_go/pkg/jwt/validation/types/jwk_validator"
	"github.com/Motmedel/utils_go/pkg/jwt/validation/types/registered_claims_validator"
	"github.com/Motmedel/utils_go/pkg/jwt/validation/types/setting"
	motmedelMaps "github.com/Motmedel/utils_go/pkg/maps"
	"github.com/Motmedel/utils_go/pkg/utils"
	"io"
)

func ParseAndCheckWithValidator(
	tokenString string,
	signatureVerifier motmedelCryptoInterfaces.NamedVerifier,
	tokenValidator validator.Validator[*jwtToken.Token],
) (*jwtToken.Token, error) {
	if tokenString == "" {
		return nil, nil
	}

	rawToken, err := raw_token.New(tokenString)
	if err != nil {
		return nil, fmt.Errorf("%w: raw token new: %w", motmedelErrors.ErrParseError, err)
	}
	if rawToken == nil {
		return nil, motmedelErrors.NewWithTrace(jwtErrors.ErrNilRawToken)
	}

	token, err := jwtToken.FromRawToken(rawToken)
	if err != nil {
		return nil, motmedelErrors.New(
			fmt.Errorf("%w: token from raw token: %w", motmedelErrors.ErrParseError, err),
			rawToken,
		)
	}

	if !utils.IsNil(signatureVerifier) {
		if token == nil {
			return nil, motmedelErrors.NewWithTrace(jwtErrors.ErrNilToken)
		}

		tokenHeader := token.Header
		if tokenHeader == nil {
			return token, motmedelErrors.NewWithTrace(jwtErrors.ErrNilTokenHeader)
		}

		alg, err := motmedelMaps.MapGetConvert[string](tokenHeader, "alg")
		if err != nil {
			var wrappedErr error = motmedelErrors.New(fmt.Errorf("map get convert: %w", err), tokenHeader)
			if motmedelErrors.IsAny(err, motmedelErrors.ErrConversionNotOk, motmedelErrors.ErrNotInMap) {
				wrappedErr = fmt.Errorf("%w: %w", motmedelErrors.ErrValidationError, wrappedErr)
			}
			return token, wrappedErr
		}

		verifierMethodName := signatureVerifier.GetName()
		if alg != verifierMethodName {
			return nil, motmedelErrors.NewWithTrace(
				fmt.Errorf("%w: %w", motmedelErrors.ErrVerificationError, jwtErrors.ErrAlgorithmMismatch),
				jwtErrors.ErrAlgorithmMismatch, alg, verifierMethodName,
			)
		}

		if err := rawToken.Verify(signatureVerifier); err != nil {
			return nil, motmedelErrors.New(fmt.Errorf("raw token verify: %w", err), rawToken)
		}
	}

	if !utils.IsNil(tokenValidator) {
		if err := tokenValidator.Validate(token); err != nil {
			return nil, motmedelErrors.New(
				fmt.Errorf("%w: token validator validate: %w", motmedelErrors.ErrValidationError, err),
				token,
			)
		}
	}

	return token, nil
}

func ParseAndCheck(tokenString string, signatureVerifier motmedelCryptoInterfaces.NamedVerifier) (*jwtToken.Token, error) {
	return ParseAndCheckWithValidator(
		tokenString,
		signatureVerifier,
		&base_validator.BaseValidator{
			PayloadValidator: &registered_claims_validator.RegisteredClaimsValidator{},
		},
	)
}

func ParseAndCheckJwkWithValidator(tokenString string, tokenValidator validator.Validator[*jwtToken.Token]) (*jwtToken.Token, []byte, error) {
	if tokenString == "" {
		return nil, nil, nil
	}

	rawToken, err := raw_token.New(tokenString)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: raw token new: %w", motmedelErrors.ErrParseError, err)
	}
	if rawToken == nil {
		return nil, nil, motmedelErrors.NewWithTrace(jwtErrors.ErrNilRawToken)
	}

	token, err := jwtToken.FromRawToken(rawToken)
	if err != nil {
		return nil, nil, motmedelErrors.New(
			fmt.Errorf("%w: token from raw token: %w", motmedelErrors.ErrParseError, err),
			rawToken,
		)
	}

	if !utils.IsNil(tokenValidator) {
		if err := tokenValidator.Validate(token); err != nil {
			return token, nil, motmedelErrors.New(
				fmt.Errorf("%w: token validator validate: %w", motmedelErrors.ErrValidationError, err),
				token,
			)
		}
	}

	if token == nil {
		return nil, nil, motmedelErrors.NewWithTrace(jwtErrors.ErrNilToken)
	}

	tokenHeader := token.Header
	if tokenHeader == nil {
		return token, nil, motmedelErrors.NewWithTrace(jwtErrors.ErrNilTokenHeader)
	}

	alg, err := motmedelMaps.MapGetConvert[string](tokenHeader, "alg")
	if err != nil {
		var wrappedErr error = motmedelErrors.New(fmt.Errorf("map get convert: %w", err), tokenHeader)
		if motmedelErrors.IsAny(err, motmedelErrors.ErrConversionNotOk, motmedelErrors.ErrNotInMap) {
			wrappedErr = fmt.Errorf("%w: %w", motmedelErrors.ErrValidationError, wrappedErr)
		}
		return token, nil, wrappedErr
	}

	tokenPayload := token.Payload
	if tokenPayload == nil {
		return token, nil, motmedelErrors.NewWithTrace(jwtErrors.ErrNilTokenPayload)
	}

	keyMap, err := motmedelMaps.MapGetConvert[map[string]any](tokenPayload, "key")
	if err != nil {
		var wrappedErr error = motmedelErrors.New(fmt.Errorf("map get convert: %w", err), tokenPayload)
		if motmedelErrors.IsAny(err, motmedelErrors.ErrConversionNotOk, motmedelErrors.ErrNotInMap) {
			wrappedErr = fmt.Errorf("%w: %w", motmedelErrors.ErrValidationError, wrappedErr)
		}
		return token, nil, wrappedErr
	}

	kty, err := motmedelMaps.MapGetConvert[string](keyMap, "kty")
	if err != nil {
		var wrappedErr error = motmedelErrors.New(fmt.Errorf("map get convert: %w", err), keyMap)
		if motmedelErrors.IsAny(err, motmedelErrors.ErrConversionNotOk, motmedelErrors.ErrNotInMap) {
			wrappedErr = fmt.Errorf("%w: %w", motmedelErrors.ErrValidationError, wrappedErr)
		}
		return token, nil, wrappedErr
	}

	var method motmedelCryptoInterfaces.Method

	var key any

	switch kty {
	case "RSA":
		rsaKey, err := jwtKey.RsaKeyFromMap(keyMap)
		if err != nil {
			var wrappedErr error = motmedelErrors.New(fmt.Errorf("rsa key from map: %w", err), keyMap)
			if motmedelErrors.IsAny(err, motmedelErrors.ErrConversionNotOk, motmedelErrors.ErrNotInMap) {
				wrappedErr = fmt.Errorf("%w: %w", motmedelErrors.ErrValidationError, wrappedErr)
			}
			return token, nil, wrappedErr
		}

		publicKey, err := rsaKey.PublicKey()
		if err != nil {
			var wrappedErr error = motmedelErrors.New(fmt.Errorf("ec key public key: %w", err), rsaKey)

			var corruptInputError base64.CorruptInputError
			if errors.Is(err, io.ErrUnexpectedEOF) || errors.As(err, &corruptInputError) {
				wrappedErr = fmt.Errorf("%w: %w", motmedelErrors.ErrValidationError, wrappedErr)
			}

			return token, nil, wrappedErr
		}

		method, err = rsa.New(alg, nil, publicKey)
		if err != nil {
			var wrappedErr error = motmedelErrors.New(fmt.Errorf("ecdsa new: %w", err), alg)
			if errors.Is(err, motmedelCryptoErrors.ErrUnsupportedAlgorithm) {
				wrappedErr = fmt.Errorf("%w: %w", motmedelErrors.ErrValidationError, wrappedErr)
			}
			return token, nil, wrappedErr
		}

		key = rsaKey
	case "EC":
		ecKey, err := jwtKey.EcKeyFromMap(keyMap)
		if err != nil {
			var wrappedErr error = motmedelErrors.New(fmt.Errorf("ec key from map: %w", err), keyMap)
			if motmedelErrors.IsAny(err, motmedelErrors.ErrConversionNotOk, motmedelErrors.ErrNotInMap) {
				wrappedErr = fmt.Errorf("%w: %w", motmedelErrors.ErrValidationError, wrappedErr)
			}
			return token, nil, wrappedErr
		}

		publicKey, err := ecKey.PublicKey()
		if err != nil {
			var wrappedErr error = motmedelErrors.New(fmt.Errorf("ec key public key: %w", err), ecKey)

			var corruptInputError base64.CorruptInputError
			if motmedelErrors.IsAny(err, jwtErrors.ErrUnsupportedCrv, io.ErrUnexpectedEOF) || errors.As(err, &corruptInputError) {
				wrappedErr = fmt.Errorf("%w: %w", motmedelErrors.ErrValidationError, wrappedErr)
			}

			return token, nil, wrappedErr
		}

		method, err = ecdsa.FromPublicKey(publicKey)
		if err != nil {
			var wrappedErr error = motmedelErrors.New(fmt.Errorf("ecdsa from public key: %w", err), publicKey)
			if motmedelErrors.IsAny(err, motmedelCryptoErrors.ErrCurveMismatch, motmedelCryptoErrors.ErrUnsupportedAlgorithm) {
				wrappedErr = fmt.Errorf("%w: %w", motmedelErrors.ErrValidationError, wrappedErr)
			}
			return token, nil, wrappedErr
		}

		key = ecKey
	default:
		return token, nil, motmedelErrors.NewWithTrace(jwtErrors.ErrUnsupportedKty, kty)
	}

	derEncodedKey, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return token, nil, motmedelErrors.NewWithTrace(
			fmt.Errorf("%w: x509 marshal pkix public key: %w", motmedelErrors.ErrValidationError, err),
			key,
		)
	}

	if err := rawToken.Verify(method); err != nil {
		return token, derEncodedKey, motmedelErrors.New(fmt.Errorf("raw token verify: %w", err), rawToken, method)
	}

	return token, derEncodedKey, nil
}

func ParseAndCheckJwk(tokenString string) (*jwtToken.Token, any, error) {
	return ParseAndCheckJwkWithValidator(
		tokenString,
		&jwk_validator.JwkValidator{
			BaseValidator: base_validator.BaseValidator{
				HeaderValidator: &header_validator.HeaderValidator{
					Settings: map[string]setting.Setting{
						"alg": setting.SettingRequired,
					},
				},
				PayloadValidator: &registered_claims_validator.RegisteredClaimsValidator{
					Settings: map[string]setting.Setting{
						"key": setting.SettingRequired,
					},
				},
			},
		},
	)
}
