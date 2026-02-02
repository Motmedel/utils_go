package forwarded

import (
	_ "embed"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Motmedel/parsing_utils/pkg/parsing_utils"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	goabnf "github.com/pandatix/go-abnf"
)

//go:embed grammar.txt
var grammar []byte

var Grammar *goabnf.Grammar

var (
	ErrInvalidQuotedValue = errors.New("invalid quoted value")
)

// ForwardedElement represents a single forwarded element containing multiple parameters.
// Standard parameters defined in RFC 7239 are:
//   - For: identifies the node making the request to the proxy
//   - By: identifies the interface where the request came in to the proxy
//   - Host: the original value of the Host request header
//   - Proto: indicates the protocol used to make the request (http or https)
type ForwardedElement struct {
	For   string
	By    string
	Host  string
	Proto string
	// Extensions contains any non-standard parameters
	Extensions map[string]string
}

// Forwarded represents the parsed Forwarded HTTP header as defined in RFC 7239.
// The header can contain multiple elements, each potentially originating from
// different proxies in the request chain.
type Forwarded struct {
	Elements []*ForwardedElement
}

func Parse(data []byte) (*Forwarded, error) {
	paths, err := parsing_utils.GetParsedDataPaths(Grammar, data)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("get parsed data paths: %w", err), data)
	}
	if len(paths) == 0 {
		return nil, motmedelErrors.NewWithTrace(motmedelErrors.ErrSyntaxError, data)
	}

	var forwarded Forwarded

	elementPaths := parsing_utils.SearchPath(
		paths[0],
		[]string{"forwarded-element"}, 2, false,
	)

	for _, elementPath := range elementPaths {
		element := &ForwardedElement{}

		pairPaths := parsing_utils.SearchPath(
			elementPath,
			[]string{"forwarded-pair"}, 2, false,
		)

		for _, pairPath := range pairPaths {
			pairValue := string(parsing_utils.ExtractPathValue(data, pairPath))

			name, value, found := strings.Cut(pairValue, "=")
			if !found {
				continue
			}

			name = strings.ToLower(strings.TrimSpace(name))
			value = strings.TrimSpace(value)

			quotedStringPath := parsing_utils.SearchPathSingleName(
				pairPath,
				"quoted-string",
				1,
				false,
			)
			if quotedStringPath != nil {
				quotedString := string(parsing_utils.ExtractPathValue(data, quotedStringPath))
				unquotedValue, err := strconv.Unquote(quotedString)
				if err != nil {
					return nil, motmedelErrors.NewWithTrace(
						fmt.Errorf(
							"%w: %w: strconv unquote: %w",
							motmedelErrors.ErrSemanticError,
							ErrInvalidQuotedValue,
							err,
						),
						quotedString,
					)
				}
				value = unquotedValue
			}

			switch name {
			case "for":
				element.For = value
			case "by":
				element.By = value
			case "host":
				element.Host = value
			case "proto":
				element.Proto = value
			default:
				if element.Extensions == nil {
					element.Extensions = make(map[string]string)
				}
				element.Extensions[name] = value
			}
		}

		forwarded.Elements = append(forwarded.Elements, element)
	}

	return &forwarded, nil
}

func init() {
	var err error
	Grammar, err = goabnf.ParseABNF(grammar)
	if err != nil {
		panic(fmt.Sprintf("goabnf parse abnf (forwarded grammar): %v", err))
	}
}
