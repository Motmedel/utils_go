package content_type

import (
	"github.com/Motmedel/parsing_utils/parsing_utils"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelHttpTypes "github.com/Motmedel/utils_go/pkg/http/types"
	goabnf "github.com/pandatix/go-abnf"
	"strconv"
	"strings"
)

var ContentTypeGrammar *goabnf.Grammar

func ParseContentType(data []byte) (*motmedelHttpTypes.ContentType, error) {
	paths, err := goabnf.Parse(data, ContentTypeGrammar, "root")
	if err != nil {
		return nil, &motmedelErrors.InputError{
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
				parameterValue, err = strconv.Unquote(string(parsing_utils.ExtractPathValue(data, quotedStringPath)))
				if err != nil {
					panic(err)
				}
			}

			contentType.Parameters = append(contentType.Parameters, [2]string{key, parameterValue})
		}
	}

	return &contentType, nil
}

/*
root=Content-Type
Content-Type=media-type
media-type=type "/" subtype *(OWS ";" OWS parameter)
OWS=*(SP/HTAB)
type=token
subtype=token
parameter=token "=" (token/quoted-string)
token=1*tchar
tchar="!"/"#"/"$"/"%"/"&"/"'"/"*"/"+"/"-"/"."/"^"/"_"/"`"/"|"/"~"/DIGIT/ALPHA
quoted-string=DQUOTE *(qdtext/quoted-pair) DQUOTE
qdtext=HTAB/SP/%x21/%x23-5B/%x5D-7E
quoted-pair="\" (HTAB/SP/VCHAR)
*/

var grammar = []uint8{114, 111, 111, 116, 61, 67, 111, 110, 116, 101, 110, 116, 45, 84, 121, 112, 101, 13, 10, 67, 111, 110, 116, 101, 110, 116, 45, 84, 121, 112, 101, 61, 109, 101, 100, 105, 97, 45, 116, 121, 112, 101, 13, 10, 109, 101, 100, 105, 97, 45, 116, 121, 112, 101, 61, 116, 121, 112, 101, 32, 34, 47, 34, 32, 115, 117, 98, 116, 121, 112, 101, 32, 42, 40, 79, 87, 83, 32, 34, 59, 34, 32, 79, 87, 83, 32, 112, 97, 114, 97, 109, 101, 116, 101, 114, 41, 13, 10, 79, 87, 83, 61, 42, 40, 83, 80, 47, 72, 84, 65, 66, 41, 13, 10, 116, 121, 112, 101, 61, 116, 111, 107, 101, 110, 13, 10, 115, 117, 98, 116, 121, 112, 101, 61, 116, 111, 107, 101, 110, 13, 10, 112, 97, 114, 97, 109, 101, 116, 101, 114, 61, 116, 111, 107, 101, 110, 32, 34, 61, 34, 32, 40, 116, 111, 107, 101, 110, 47, 113, 117, 111, 116, 101, 100, 45, 115, 116, 114, 105, 110, 103, 41, 13, 10, 116, 111, 107, 101, 110, 61, 49, 42, 116, 99, 104, 97, 114, 13, 10, 116, 99, 104, 97, 114, 61, 34, 33, 34, 47, 34, 35, 34, 47, 34, 36, 34, 47, 34, 37, 34, 47, 34, 38, 34, 47, 34, 39, 34, 47, 34, 42, 34, 47, 34, 43, 34, 47, 34, 45, 34, 47, 34, 46, 34, 47, 34, 94, 34, 47, 34, 95, 34, 47, 34, 96, 34, 47, 34, 124, 34, 47, 34, 126, 34, 47, 68, 73, 71, 73, 84, 47, 65, 76, 80, 72, 65, 13, 10, 113, 117, 111, 116, 101, 100, 45, 115, 116, 114, 105, 110, 103, 61, 68, 81, 85, 79, 84, 69, 32, 42, 40, 113, 100, 116, 101, 120, 116, 47, 113, 117, 111, 116, 101, 100, 45, 112, 97, 105, 114, 41, 32, 68, 81, 85, 79, 84, 69, 13, 10, 113, 100, 116, 101, 120, 116, 61, 72, 84, 65, 66, 47, 83, 80, 47, 37, 120, 50, 49, 47, 37, 120, 50, 51, 45, 53, 66, 47, 37, 120, 53, 68, 45, 55, 69, 13, 10, 113, 117, 111, 116, 101, 100, 45, 112, 97, 105, 114, 61, 34, 92, 34, 32, 40, 72, 84, 65, 66, 47, 83, 80, 47, 86, 67, 72, 65, 82, 41, 13, 10}

func init() {
	var err error
	ContentTypeGrammar, err = goabnf.ParseABNF(grammar)
	if err != nil {
		panic(err)
	}
}
