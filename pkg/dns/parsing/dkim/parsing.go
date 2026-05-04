package dkim

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"iter"
	"regexp"
	"strings"

	"github.com/Motmedel/parsing_utils/pkg/parsing_utils"
	dnsTypes "github.com/Motmedel/utils_go/pkg/dns/types"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	goabnf "github.com/pandatix/go-abnf"
)

var (
	ErrNilTagNamePath         = errors.New("nil tag name path")
	ErrVNotFirstTag           = errors.New("v not first tag")
	ErrMalformedTag           = errors.New("malformed tag")
	ErrMultipleTagPaths       = errors.New("multiple tag paths")
	ErrDuplicateTags          = errors.New("duplicate tags")
	ErrMissingPublicKeyData   = errors.New("missing public key data")
	ErrMalformedPublicKeyData = errors.New("malformed public key data")
	ErrUnexpectedTag          = errors.New("unexpected tag")
	ErrNilHeaderName          = errors.New("nil header name")
	ErrNilHeaderValue         = errors.New("nil header value")
	ErrEmptyTagType           = errors.New("empty tag type")
	ErrUnexpectedTagType      = errors.New("unexpected tag type")
	ErrEmptyPathInput         = errors.New("empty path input")
	ErrNilItem                = errors.New("nil item")
	ErrNilTagMap              = errors.New("nil tag map")

	ErrMissingRequiredTag = errors.New("missing required tag")
)

var (
	reUnfold = regexp.MustCompile(`\r?\n[ \t]+`)
	reTabs   = regexp.MustCompile(`\t+`)
	reSpaces = regexp.MustCompile(` +`)
)

func extractTagPath(tagName string, tagValue []byte, tagType string) (*goabnf.Path, error) {
	if tagName == "" {
		return nil, nil
	}

	if tagType == "" {
		return nil, motmedelErrors.NewWithTrace(ErrEmptyTagType)
	}

	var ruleName string

	if tagType == "key" {
		switch tagName {
		case "v", "h", "k", "n", "p", "s", "t":
			ruleName = fmt.Sprintf("key-%s-tag-root", tagName)
		default:
			return nil, nil
		}
	} else if tagType == "signature" {
		switch tagName {
		case "v", "a", "b", "bh", "c", "d", "h", "i", "l", "q", "s", "t", "x", "z":
			ruleName = fmt.Sprintf("sig-%s-tag-root", tagName)
		default:
			return nil, nil
		}
	} else {
		return nil, motmedelErrors.NewWithTrace(fmt.Errorf("%w: %s", ErrUnexpectedTagType, tagType), tagType)
	}

	tagPaths, err := goabnf.Parse(tagValue, DkimGrammar, ruleName)
	if err != nil {
		return nil, motmedelErrors.NewWithTrace(fmt.Errorf("go abnf parse: %w", err), tagValue, DkimGrammar)
	}
	if len(tagPaths) == 0 {
		return nil, motmedelErrors.NewWithTrace(
			fmt.Errorf("%w: %w: %q", motmedelErrors.ErrSyntaxError, ErrMalformedTag, tagName),
			tagValue, DkimGrammar,
		)
	}
	if len(tagPaths) > 1 {
		return nil, motmedelErrors.NewWithTrace(ErrMultipleTagPaths, tagValue, DkimGrammar)
	}

	return tagPaths[0], nil
}

func extractBase64String(path *goabnf.Path, value []byte) (string, error) {
	if path == nil {
		return "", nil
	}

	if len(value) == 0 {
		return "", motmedelErrors.NewWithTrace(ErrEmptyPathInput)
	}

	var segments []string
	for _, p := range parsing_utils.SearchPath(path, []string{"ALPHADIGITPS"}, 2, false) {
		segments = append(segments, string(parsing_utils.ExtractPathValue(value, p)))
	}

	return strings.Join(segments, ""), nil
}

func normalizeEmailHeader(headerValue []byte) []byte {
	if len(headerValue) == 0 {
		return nil
	}

	unfolded := reUnfold.ReplaceAll(headerValue, []byte(" "))
	withoutTabs := reTabs.ReplaceAll(unfolded, []byte(""))
	collapsed := reSpaces.ReplaceAll(withoutTabs, []byte(" "))

	return bytes.TrimSpace(collapsed)
}

type tagSpecItem struct {
	Name  string
	Value []byte
}

func getTagSpecItems(path *goabnf.Path, tagMap map[string]struct{}, data []byte) iter.Seq2[*tagSpecItem, error] {
	return func(yield func(*tagSpecItem, error) bool) {
		if tagMap == nil {
			yield(nil, motmedelErrors.NewWithTrace(ErrNilTagMap))
			return
		}

		if len(data) == 0 {
			yield(nil, motmedelErrors.NewWithTrace(ErrEmptyPathInput))
			return
		}

		for _, tagSpecPath := range parsing_utils.SearchPath(path, []string{"tag-spec"}, 2, false) {
			tagNamePath := parsing_utils.SearchPathSingleName(tagSpecPath, "tag-name", 1, false)
			if tagNamePath == nil {
				yield(nil, motmedelErrors.NewWithTrace(ErrNilTagNamePath))
				return
			}
			tagName := string(parsing_utils.ExtractPathValue(data, tagNamePath))
			if _, ok := tagMap[tagName]; ok {
				yield(
					nil,
					motmedelErrors.NewWithTrace(
						fmt.Errorf("%w: %w: %s", motmedelErrors.ErrSemanticError, ErrDuplicateTags, tagName),
					),
				)
				return
			}
			tagMap[tagName] = struct{}{}

			var tagValue []byte
			tagValuePath := parsing_utils.SearchPathSingleName(tagSpecPath, "tag-value", 1, false)
			if tagValuePath != nil {
				tagValue = bytes.TrimSpace(parsing_utils.ExtractPathValue(data, tagValuePath))
			}

			if !yield(&tagSpecItem{Name: tagName, Value: tagValue}, nil) {
				return
			}
		}
	}
}

func ParseRecord(data []byte) (*dnsTypes.DkimRecord, error) {
	paths, err := parsing_utils.GetParsedDataPaths(DkimGrammar, data)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("get parsed data paths: %w", err), data)
	}
	if len(paths) == 0 {
		return nil, motmedelErrors.NewWithTrace(motmedelErrors.ErrSyntaxError, data)
	}

	var record dnsTypes.DkimRecord
	record.Raw = string(data)

	tagMap := make(map[string]struct{})

	i := 0
	for item, err := range getTagSpecItems(paths[0], tagMap, data) {
		if err != nil {
			return nil, fmt.Errorf("get tag spec item: %w", err)
		}
		if item == nil {
			return nil, motmedelErrors.NewWithTrace(ErrNilItem)
		}

		i += 1
		tagName := item.Name
		tagValue := item.Value

		tagPath, err := extractTagPath(tagName, tagValue, "key")
		if err != nil {
			return nil, motmedelErrors.New(fmt.Errorf("extract tag path: %w", err), tagName, tagValue)
		}
		if tagPath == nil {
			record.Extensions = append(record.Extensions, [2]string{tagName, string(tagValue)})
			continue
		}
		if len(tagValue) == 0 {
			continue
		}

		switch tagName {
		case "v":
			if i != 1 {
				return nil, motmedelErrors.NewWithTrace(
					fmt.Errorf("%w: %w", motmedelErrors.ErrSemanticError, ErrVNotFirstTag),
				)
			}
			record.Version = 1
		case "h":
			var algorithms []string
			for _, path := range parsing_utils.SearchPath(tagPath, []string{"key-h-tag-alg"}, 1, false) {
				algorithms = append(algorithms, string(parsing_utils.ExtractPathValue(tagValue, path)))
			}
			record.AcceptableHashAlgorithms = algorithms
		case "k":
			record.KeyType = string(
				parsing_utils.ExtractPathValue(
					tagValue,
					parsing_utils.SearchPathSingleName(tagPath, "key-k-tag-type", 1, false),
				),
			)
		case "n":
			record.Notes = string(
				parsing_utils.ExtractPathValue(
					tagValue,
					parsing_utils.SearchPathSingleName(tagPath, "qp-section", 1, false),
				),
			)
		case "p":
			base64String, err := extractBase64String(tagPath, tagValue)
			if err != nil {
				return nil, motmedelErrors.New(fmt.Errorf("extract base64 string: %w", err), tagName, tagValue)
			}
			record.PublicKeyData = base64String
		case "s":
			record.ServiceType = string(
				parsing_utils.ExtractPathValue(
					tagValue,
					parsing_utils.SearchPathSingleName(tagPath, "key-s-tag-type", 1, false),
				),
			)
		case "t":
			var flags []string
			for _, path := range parsing_utils.SearchPath(tagPath, []string{"key-t-tag-flag"}, 1, false) {
				flags = append(flags, string(parsing_utils.ExtractPathValue(tagValue, path)))
			}
			record.Flags = flags
		}
	}

	if _, ok := tagMap["p"]; !ok {
		return nil, motmedelErrors.NewWithTrace(
			fmt.Errorf("%w: %w", motmedelErrors.ErrSemanticError, ErrMissingPublicKeyData),
		)
	}

	if _, err := record.GetPublicKey(); err != nil {
		return nil, motmedelErrors.New(
			fmt.Errorf("%w: %w: get public key: %w", motmedelErrors.ErrSemanticError, ErrMalformedPublicKeyData, err),
			record,
		)
	}

	return &record, nil
}

func ParseHeader(data []byte) (*dnsTypes.DkimHeader, error) {
	normalizedData := normalizeEmailHeader(data)
	paths, err := parsing_utils.GetParsedDataPaths(DkimGrammar, normalizedData)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("get parsed data paths: %w", err), normalizedData)
	}
	if len(paths) == 0 {
		return nil, motmedelErrors.NewWithTrace(motmedelErrors.ErrSyntaxError, normalizedData)
	}

	var header dnsTypes.DkimHeader
	header.Raw = string(data)

	tagMap := make(map[string]struct{})

	for item, err := range getTagSpecItems(paths[0], tagMap, normalizedData) {
		if err != nil {
			return nil, fmt.Errorf("get tag spec item: %w", err)
		}
		if item == nil {
			return nil, motmedelErrors.NewWithTrace(ErrNilItem)
		}

		tagName := item.Name
		tagValue := item.Value

		tagPath, err := extractTagPath(tagName, tagValue, "signature")
		if err != nil {
			return nil, motmedelErrors.New(fmt.Errorf("extract tag path: %w", err), tagName, tagValue)
		}
		if tagPath == nil {
			header.Extensions = append(header.Extensions, [2]string{tagName, string(tagValue)})
			continue
		}
		if len(tagValue) == 0 {
			continue
		}

		switch tagName {
		case "v":
			header.Version = 1
		case "a":
			header.Algorithm = string(tagValue)
		case "b":
			base64String, err := extractBase64String(tagPath, tagValue)
			if err != nil {
				return nil, motmedelErrors.New(fmt.Errorf("extract base64 string: %w", err), tagName, tagValue)
			}
			header.Signature = base64String
		case "bh":
			base64String, err := extractBase64String(tagPath, tagValue)
			if err != nil {
				return nil, motmedelErrors.New(fmt.Errorf("extract base64 string: %w", err), tagName, tagValue)
			}
			header.Hash = base64String
		case "c":
			header.MessageCanonicalization = string(tagValue)
		case "d":
			header.SigningDomainIdentifier = string(tagValue)
		case "h":
			var fields []string
			for _, path := range parsing_utils.SearchPath(tagPath, []string{"hdr-name"}, 1, false) {
				fields = append(fields, string(parsing_utils.ExtractPathValue(tagValue, path)))
			}
			header.SignedHeaderFields = fields
		case "i":
			header.AgentOrUserIdentifier = string(tagValue)
		case "l":
			header.BodyLengthCount = string(tagValue)
		case "q":
			var methods []string
			for _, path := range parsing_utils.SearchPath(tagPath, []string{"sig-q-tag-method"}, 1, false) {
				methods = append(methods, string(parsing_utils.ExtractPathValue(tagValue, path)))
			}
			header.QueryMethods = methods
		case "s":
			header.Selector = string(tagValue)
		case "t":
			header.SignatureTimestamp = string(tagValue)
		case "x":
			header.SignatureExpiration = string(tagValue)
		case "z":
			var fields [][2]string
			for _, path := range parsing_utils.SearchPath(tagPath, []string{"sig-z-tag-copy"}, 2, false) {
				namePath := parsing_utils.SearchPathSingleName(path, "hdr-name", 1, false)
				if namePath == nil {
					return nil, motmedelErrors.NewWithTrace(ErrNilHeaderName)
				}

				valuePath := parsing_utils.SearchPathSingleName(path, "qp-hdr-value", 1, false)
				if valuePath == nil {
					return nil, motmedelErrors.NewWithTrace(ErrNilHeaderValue)
				}

				name := string(parsing_utils.ExtractPathValue(tagValue, namePath))
				value := string(parsing_utils.ExtractPathValue(tagValue, valuePath))

				fields = append(fields, [2]string{name, value})
			}

			header.CopiedHeaderFields = fields
		}
	}

	for _, tag := range []string{"v", "a", "b", "bh", "d", "h", "s"} {
		if _, ok := tagMap[tag]; !ok {
			return nil, motmedelErrors.NewWithTrace(
				fmt.Errorf("%w: %w: %s", motmedelErrors.ErrSemanticError, ErrMissingRequiredTag, tag),
			)
		}
	}

	for _, tag := range []string{"b", "bh"} {
		var data string
		switch tag {
		case "b":
			data = header.Signature
		case "bh":
			data = header.Hash
		default:
			return nil, motmedelErrors.NewWithTrace(fmt.Errorf("%w: %s", ErrUnexpectedTag, tag))
		}

		if _, err = base64.StdEncoding.DecodeString(data); err != nil {
			return nil, motmedelErrors.NewWithTrace(
				fmt.Errorf(
					"%w: %s: base64 std encoding decode string: %w",
					motmedelErrors.ErrSemanticError, tag, err,
				),
				data,
			)
		}
	}

	return &header, nil
}
