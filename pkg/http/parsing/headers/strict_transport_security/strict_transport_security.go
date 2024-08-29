package strict_transport_security

import (
	"github.com/Motmedel/parsing_utils/parsing_utils"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelHttpTypes "github.com/Motmedel/utils_go/pkg/http/types"
	goabnf "github.com/pandatix/go-abnf"
	"regexp"
	"strconv"
	"strings"
)

var StrictTransportSecurityGrammar *goabnf.Grammar

const (
	maxAgeDirectiveName = "max-age"
)

var digitRegexp = regexp.MustCompile(`^\d+$`)

func ParseStrictTransportSecurity(data []byte) (*motmedelHttpTypes.StrictTransportSecurityPolicy, error) {
	paths, err := goabnf.Parse(data, StrictTransportSecurityGrammar, "root")
	if err != nil {
		return nil, &motmedelErrors.InputError{
			Message: "An error occurred when parsing data as a strict transport security policy.",
			Cause:   err,
			Input:   data,
		}
	}
	if len(paths) == 0 {
		return nil, nil
	}

	directiveNameSet := make(map[string]struct{})

	var strictTransportPolicy motmedelHttpTypes.StrictTransportSecurityPolicy

	interestingPaths := parsing_utils.SearchPath(paths[0], []string{"directive"}, 2, false)
	for _, interestingPath := range interestingPaths {
		directiveNamePath := parsing_utils.SearchPathSingleName(
			interestingPath,
			"directive-name",
			1,
			false,
		)
		if directiveNamePath == nil {
			return nil, nil
			//return nil, &motmedelErrors.InputError{
			//	Message: "No directive name could be found in an interesting path.",
			//	Input:   interestingPath,
			//}
		}
		directiveName := string(parsing_utils.ExtractPathValue(data, directiveNamePath))

		var directiveStringValue string

		directiveValuePath := parsing_utils.SearchPathSingleName(
			interestingPath,
			"directive-value",
			1,
			false,
		)
		if directiveValuePath != nil {
			quotedStringPath := parsing_utils.SearchPathSingleName(
				directiveValuePath,
				"quoted-string",
				1,
				false,
			)
			if quotedStringPath != nil {
				var err error
				quotedString := string(parsing_utils.ExtractPathValue(data, quotedStringPath))
				directiveStringValue, err = strconv.Unquote(quotedString)
				if err != nil {
					return nil, &motmedelErrors.InputError{
						Message: "An error occurred when unquoting a quoted-string.",
						Cause:   err,
						Input:   quotedString,
					}
				}
			} else {
				directiveStringValue = string(parsing_utils.ExtractPathValue(data, directiveValuePath))
			}
		}

		lowercaseDirectiveName := strings.ToLower(directiveName)
		if _, ok := directiveNameSet[lowercaseDirectiveName]; ok {
			return nil, nil
			//return nil, &MultipleSameNameDirectivesError{
			//	InputError: motmedelErrors.InputError{
			//		Message: "Multiple directives with the same name were encountered.",
			//		Input:   lowercaseDirectiveName,
			//	},
			//}
		}
		directiveNameSet[lowercaseDirectiveName] = struct{}{}

		switch lowercaseDirectiveName {
		case maxAgeDirectiveName:
			if !digitRegexp.MatchString(directiveStringValue) {
				return nil, nil
				//return nil, &BadMaxAgeFormatError{
				//	InputError: motmedelErrors.InputError{
				//		Message: "A max-age directive value is not only digits.",
				//		Input:   directiveStringValue,
				//	},
				//}
			}

			maxAgeNumber, err := strconv.Atoi(directiveStringValue)
			if err != nil {
				return nil, &motmedelErrors.InputError{
					Message: "An error occurred when parsing a max-age value as an integer.",
					Cause:   err,
					Input:   directiveStringValue,
				}
			}

			strictTransportPolicy.MaxAga = maxAgeNumber
		case "includesubdomains":
			if directiveValuePath != nil {
				return nil, nil
				//return nil, &NonValuelessIncludeSubdomainsError{
				//	InputError: motmedelErrors.InputError{
				//		Message: "An includeSubdomains directive was encounter that is not valueless.",
				//		Input:   directiveStringValue,
				//	},
				//}
			}
			strictTransportPolicy.IncludeSubdomains = true
		}
	}

	if _, ok := directiveNameSet[maxAgeDirectiveName]; !ok {
		return nil, nil
		//return nil, &MissingMaxAgeError{
		//	InputError: motmedelErrors.InputError{
		//		Message: "The required max-age directive is missing.",
		//		Input:   data,
		//	},
		//}
	}

	if len(directiveNameSet) == 0 {
		return nil, nil
	}

	strictTransportPolicy.Raw = string(data)

	return &strictTransportPolicy, nil
}

/*
root=Strict-Transport-Security
Strict-Transport-Security = "Strict-Transport-Security" ":" [ directive ]  *( ";" [ directive ] )
directive = directive-name *LWS [ "=" *LWS directive-value *LWS ]
LWS = [CRLF] 1*( SP | HT )
directive-name = token
directive-value = token / quoted-string
token=1*tchar
tchar="!"/"#"/"$"/"%"/"&"/"'"/"*"/"+"/"-"/"."/"^"/"_"/"`"/"|"/"~"/DIGIT/ALPHA
quoted-string=DQUOTE *(qdtext/quoted-pair) DQUOTE
qdtext=HTAB/SP/%x21/%x23-5B/%x5D-7E
quoted-pair="\" (HTAB/SP/VCHAR)
*/

var grammar = []uint8{114, 111, 111, 116, 61, 83, 116, 114, 105, 99, 116, 45, 84, 114, 97, 110, 115, 112, 111, 114, 116, 45, 83, 101, 99, 117, 114, 105, 116, 121, 13, 10, 83, 116, 114, 105, 99, 116, 45, 84, 114, 97, 110, 115, 112, 111, 114, 116, 45, 83, 101, 99, 117, 114, 105, 116, 121, 32, 61, 32, 91, 32, 100, 105, 114, 101, 99, 116, 105, 118, 101, 32, 93, 32, 42, 40, 32, 42, 76, 87, 83, 32, 34, 59, 34, 32, 42, 76, 87, 83, 32, 91, 32, 100, 105, 114, 101, 99, 116, 105, 118, 101, 32, 93, 32, 41, 13, 10, 100, 105, 114, 101, 99, 116, 105, 118, 101, 32, 61, 32, 100, 105, 114, 101, 99, 116, 105, 118, 101, 45, 110, 97, 109, 101, 32, 42, 76, 87, 83, 32, 91, 32, 34, 61, 34, 32, 42, 76, 87, 83, 32, 100, 105, 114, 101, 99, 116, 105, 118, 101, 45, 118, 97, 108, 117, 101, 32, 42, 76, 87, 83, 32, 93, 13, 10, 76, 87, 83, 32, 61, 32, 91, 67, 82, 76, 70, 93, 32, 49, 42, 40, 83, 80, 47, 72, 84, 65, 66, 41, 13, 10, 100, 105, 114, 101, 99, 116, 105, 118, 101, 45, 110, 97, 109, 101, 32, 61, 32, 116, 111, 107, 101, 110, 13, 10, 100, 105, 114, 101, 99, 116, 105, 118, 101, 45, 118, 97, 108, 117, 101, 32, 61, 32, 116, 111, 107, 101, 110, 47, 113, 117, 111, 116, 101, 100, 45, 115, 116, 114, 105, 110, 103, 13, 10, 116, 111, 107, 101, 110, 61, 49, 42, 116, 99, 104, 97, 114, 13, 10, 116, 99, 104, 97, 114, 61, 34, 33, 34, 47, 34, 35, 34, 47, 34, 36, 34, 47, 34, 37, 34, 47, 34, 38, 34, 47, 34, 39, 34, 47, 34, 42, 34, 47, 34, 43, 34, 47, 34, 45, 34, 47, 34, 46, 34, 47, 34, 94, 34, 47, 34, 95, 34, 47, 34, 96, 34, 47, 34, 124, 34, 47, 34, 126, 34, 47, 68, 73, 71, 73, 84, 47, 65, 76, 80, 72, 65, 13, 10, 113, 117, 111, 116, 101, 100, 45, 115, 116, 114, 105, 110, 103, 61, 68, 81, 85, 79, 84, 69, 32, 42, 40, 113, 100, 116, 101, 120, 116, 47, 113, 117, 111, 116, 101, 100, 45, 112, 97, 105, 114, 41, 32, 68, 81, 85, 79, 84, 69, 13, 10, 113, 100, 116, 101, 120, 116, 61, 72, 84, 65, 66, 47, 83, 80, 47, 37, 120, 50, 49, 47, 37, 120, 50, 51, 45, 53, 66, 47, 37, 120, 53, 68, 45, 55, 69, 13, 10, 113, 117, 111, 116, 101, 100, 45, 112, 97, 105, 114, 61, 34, 92, 34, 32, 40, 72, 84, 65, 66, 47, 83, 80, 47, 86, 67, 72, 65, 82, 41, 13, 10, 13, 10}

func init() {
	var err error
	StrictTransportSecurityGrammar, err = goabnf.ParseABNF(grammar)
	if err != nil {
		panic(err)
	}
}
