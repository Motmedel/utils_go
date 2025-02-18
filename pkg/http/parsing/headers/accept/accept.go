package accept

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

var AcceptGrammar *goabnf.Grammar

var (
	ErrCouldNotSplitParameter = errors.New("could not split parameter")
	ErrNilQValuePath          = errors.New("nil qvalue path")
	ErrParameterNamedQ        = errors.New("parameter named q")
)

func ParseAccept(data []byte) (*motmedelHttpTypes.Accept, error) {
	paths, err := goabnf.Parse(data, AcceptGrammar, "root")
	if err != nil {
		return nil, &motmedelErrors.Error{
			Message: "An error occurred when parsing data as an accept value.",
			Cause:   err,
			Input:   data,
		}
	}

	if len(paths) == 0 {
		return nil, nil
	}

	accept := &motmedelHttpTypes.Accept{Raw: string(data)}

	interestingPaths := parsing_utils.SearchPath(
		paths[0],
		[]string{"media-range"},
		2,
		false,
	)

	for _, interestingPath := range interestingPaths {
		if interestingPath == nil {
			continue
		}

		mediaRange := &motmedelHttpTypes.MediaRange{Weight: 1.0}

		mediaRangeType := "*"
		mediaRangeSubtype := "*"

		if typePath := parsing_utils.SearchPathSingle(interestingPath, []string{"type"}, 1, false); typePath != nil {
			mediaRangeType = string(parsing_utils.ExtractPathValue(data, typePath))
		}

		if subtypePath := parsing_utils.SearchPathSingle(interestingPath, []string{"subtype"}, 1, false); subtypePath != nil {
			mediaRangeSubtype = string(parsing_utils.ExtractPathValue(data, subtypePath))
		}

		mediaRange.Type = mediaRangeType
		mediaRange.Subtype = mediaRangeSubtype

		parameterPaths := parsing_utils.SearchPath(interestingPath, []string{"weight", "parameter"}, 2, false)
		for i, parameterPath := range parameterPaths {
			if parameterPath.MatchRule == "weight" {
				if i != len(parameterPaths)-1 {
					return nil, ErrParameterNamedQ
				}

				qValuePath := parsing_utils.SearchPathSingle(interestingPath, []string{"qvalue"}, 1, false)
				if qValuePath == nil {
					return nil, ErrNilQValuePath
				}

				qValueString := string(parsing_utils.ExtractPathValue(data, qValuePath))

				parsedWeight, err := strconv.ParseFloat(qValueString, 32)
				if err != nil {
					return nil, &motmedelErrors.Error{
						Message: "An error occurred when parsing a weight string as a float.",
						Cause:   err,
						Input:   qValueString,
					}
				}

				mediaRange.Weight = float32(parsedWeight)
			} else {
				parameterString := string(parsing_utils.ExtractPathValue(data, parameterPath))
				key, value, found := strings.Cut(parameterString, "=")
				if !found {
					return nil, &motmedelErrors.Error{
						Message: "A parameter value could not be split.",
						Cause:   ErrCouldNotSplitParameter,
						Input:   parameterString,
					}
				}

				mediaRange.Parameters = append(mediaRange.Parameters, [2]string{key, value})
			}
		}

		accept.MediaRanges = append(accept.MediaRanges, mediaRange)
	}

	return accept, nil
}

func init() {
	var err error
	AcceptGrammar, err = goabnf.ParseABNF(grammar)
	if err != nil {
		panic(fmt.Sprintf("an error occurred when parsing the grammar: %v", err))
	}
}
