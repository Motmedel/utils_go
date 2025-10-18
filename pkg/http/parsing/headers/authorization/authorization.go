package authorization

import (
	_ "embed"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Motmedel/parsing_utils/pkg/parsing_utils"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelHttpTypes "github.com/Motmedel/utils_go/pkg/http/types"
	goabnf "github.com/pandatix/go-abnf"
)

var (
	ErrNilAuthorization                   = errors.New("nil authorization")
	ErrInvalidQuotedParameterValue        = errors.New("invalid quoted parameter value")
	ErrDuplicateParameter                 = errors.New("duplicate parameter")
	ErrMutuallyExclusiveToken68Parameters = errors.New("mutually exclusive token68 or parameters")
)

var Grammar *goabnf.Grammar

func Parse(data []byte) (*motmedelHttpTypes.Authorization, error) {
	paths, err := parsing_utils.GetParsedDataPaths(Grammar, data)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("get parsed data paths: %w", err), data)
	}
	if len(paths) == 0 {
		return nil, motmedelErrors.NewWithTrace(motmedelErrors.ErrSyntaxError, data)
	}

	var authorization motmedelHttpTypes.Authorization

	interestingPaths := parsing_utils.SearchPath(
		paths[0],
		[]string{"auth-scheme", "token68", "auth-param"}, 2, false,
	)
	for _, interestingPath := range interestingPaths {
		value := string(parsing_utils.ExtractPathValue(data, interestingPath))

		switch interestingPath.MatchRule {
		case "auth-scheme":
			authorization.Scheme = value
		case "token68":
			// NOTE: Sanity check; should not be possible, is a violation of the grammar.
			if authorization.Params != nil {
				return nil, fmt.Errorf(
					"%w: %w",
					motmedelErrors.ErrSyntaxError, ErrMutuallyExclusiveToken68Parameters,
				)
			}
			authorization.Token68 = value
		case "auth-param":
			// NOTE: Sanity check; should not be possible, is a violation of the grammar.
			if authorization.Token68 != "" {
				return nil, fmt.Errorf(
					"%w: %w",
					motmedelErrors.ErrSyntaxError, ErrMutuallyExclusiveToken68Parameters,
				)
			}

			key, parameterValue, _ := strings.Cut(value, "=")
			key = strings.ToLower(strings.TrimSpace(key))
			parameterValue = strings.TrimSpace(parameterValue)

			quotedStringPath := parsing_utils.SearchPathSingleName(
				interestingPath,
				"quoted-string",
				1,
				false,
			)
			if quotedStringPath != nil {
				var err error
				quotedString := string(parsing_utils.ExtractPathValue(data, quotedStringPath))
				parameterValue, err = strconv.Unquote(quotedString)
				if err != nil {
					return nil, motmedelErrors.NewWithTrace(
						fmt.Errorf(
							"%w: %w: strvconv unquote: %w",
							motmedelErrors.ErrSemanticError,
							ErrInvalidQuotedParameterValue,
							err,
						),
						quotedString,
					)
				}
			}

			if authorization.Params == nil {
				authorization.Params = make(map[string]string)
			}

			if _, ok := authorization.Params[key]; ok {
				return nil, motmedelErrors.New(
					fmt.Errorf("%w: %w: %s", motmedelErrors.ErrSemanticError, ErrDuplicateParameter, key),
					key,
				)
			}

			authorization.Params[key] = parameterValue
		}
	}

	return &authorization, nil
}

//go:embed grammar.txt
var grammar []byte

func init() {
	var err error
	Grammar, err = goabnf.ParseABNF(grammar)
	if err != nil {
		panic(fmt.Sprintf("An error occurred when parsing the grammar: %v", err))
	}
}
