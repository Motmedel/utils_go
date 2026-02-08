package forwarded

import (
	_ "embed"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Motmedel/parsing_utils/pkg/parsing_utils"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"github.com/Motmedel/utils_go/pkg/http/types"
	goabnf "github.com/pandatix/go-abnf"
)

//go:embed grammar.txt
var grammar []byte

var Grammar *goabnf.Grammar

var (
	ErrInvalidQuotedValue = errors.New("invalid quoted value")
)

func Parse(data []byte) (*types.Forwarded, error) {
	paths, err := parsing_utils.GetParsedDataPaths(Grammar, data)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("get parsed data paths: %w", err), data)
	}
	if len(paths) == 0 {
		return nil, motmedelErrors.NewWithTrace(motmedelErrors.ErrSyntaxError, data)
	}

	var forwarded types.Forwarded

	elementPaths := parsing_utils.SearchPath(
		paths[0],
		[]string{"forwarded-element"}, 2, false,
	)

	for _, elementPath := range elementPaths {
		element := &types.ForwardedElement{}

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
