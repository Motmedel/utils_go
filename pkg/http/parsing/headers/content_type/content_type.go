package content_type

import (
	_ "embed"
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

func ParseContentType(data []byte) (*motmedelHttpTypes.ContentType, error) {
	paths, err := goabnf.Parse(data, ContentTypeGrammar, "root")
	if err != nil {
		return nil, &motmedelErrors.Error{
			Message: "An error occurred when parsing data as a content type.",
			Cause:   err,
			Input:   data,
		}
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
					return nil, &motmedelErrors.Error{
						Message: "An error occurred when unquoting a quoted-string.",
						Cause:   err,
						Input:   quotedString,
					}
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
		panic(err)
	}
}
