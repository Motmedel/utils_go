package content_type

import (
	_ "embed"
	"errors"
	"fmt"
	"github.com/Motmedel/parsing_utils/pkg/parsing_utils"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelHttpTypes "github.com/Motmedel/utils_go/pkg/http/types"
	goabnf "github.com/pandatix/go-abnf"
	"strconv"
	"strings"
)

//go:embed grammar.txt
var grammar []byte

var ContentTypeGrammar *goabnf.Grammar

var (
	ErrNilContentType              = errors.New("nil content type")
	ErrInvalidQuotedParameterValue = errors.New("invalid quoted parameter value")
)

func ParseContentType(data []byte) (*motmedelHttpTypes.ContentType, error) {
	paths, err := parsing_utils.GetParsedDataPaths(ContentTypeGrammar, data)
	if err != nil {
		return nil, motmedelErrors.MakeError(fmt.Errorf("get parsed data paths: %w", err), data)
	}
	if len(paths) == 0 {
		return nil, motmedelErrors.MakeErrorWithStackTrace(motmedelErrors.ErrSyntaxError, data)
	}

	var contentType motmedelHttpTypes.ContentType

	interestingPaths := parsing_utils.SearchPath(
		paths[0],
		[]string{"type", "subtype", "parameter"}, 2, false,
	)
	for _, interestingPath := range interestingPaths {
		value := string(parsing_utils.ExtractPathValue(data, interestingPath))
		switch interestingPath.MatchRule {
		case "type":
			contentType.Type = value
		case "subtype":
			contentType.Subtype = value
		case "parameter":
			key, parameterValue, _ := strings.Cut(value, "=")

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
					return nil, motmedelErrors.MakeErrorWithStackTrace(
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

			contentType.Parameters = append(contentType.Parameters, [2]string{key, parameterValue})
		}
	}

	return &contentType, nil
}

func init() {
	var err error
	ContentTypeGrammar, err = goabnf.ParseABNF(grammar)
	if err != nil {
		panic(fmt.Sprintf("goabnf parse abnf (content type grammar): %v", err))
	}
}
