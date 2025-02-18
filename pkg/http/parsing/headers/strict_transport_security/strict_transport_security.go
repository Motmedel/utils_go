package strict_transport_security

import (
	_ "embed"
	"github.com/Motmedel/parsing_utils/pkg/parsing_utils"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelHttpTypes "github.com/Motmedel/utils_go/pkg/http/types"
	goabnf "github.com/pandatix/go-abnf"
	"regexp"
	"strconv"
	"strings"
)

//go:embed grammar.txt
var grammar []byte

var StrictTransportSecurityGrammar *goabnf.Grammar

const (
	maxAgeDirectiveName = "max-age"
)

var digitRegexp = regexp.MustCompile(`^\d+$`)

func ParseStrictTransportSecurity(data []byte) (*motmedelHttpTypes.StrictTransportSecurityPolicy, error) {
	paths, err := goabnf.Parse(data, StrictTransportSecurityGrammar, "root")
	if err != nil {
		return nil, &motmedelErrors.Error{
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
			//return nil, &motmedelErrors.Error{
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
					return nil, &motmedelErrors.Error{
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
			//	Error: motmedelErrors.Error{
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
				//	Error: motmedelErrors.Error{
				//		Message: "A max-age directive value is not only digits.",
				//		Input:   directiveStringValue,
				//	},
				//}
			}

			maxAgeNumber, err := strconv.Atoi(directiveStringValue)
			if err != nil {
				return nil, &motmedelErrors.Error{
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
				//	Error: motmedelErrors.Error{
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
		//	Error: motmedelErrors.Error{
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

func init() {
	var err error
	StrictTransportSecurityGrammar, err = goabnf.ParseABNF(grammar)
	if err != nil {
		panic(err)
	}
}
