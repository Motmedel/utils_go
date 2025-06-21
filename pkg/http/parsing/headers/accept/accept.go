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
	ErrNilAccept              = errors.New("nil accept")
	ErrParameterNamedQ        = errors.New("parameter named q")
	ErrNilQValuePath          = errors.New("nil qvalue path")
	ErrInvalidQvalue          = errors.New("invalid qvalue")
	ErrCouldNotSplitParameter = errors.New("could not split parameter")
)

func ParseAccept(data []byte) (*motmedelHttpTypes.Accept, error) {
	paths, err := parsing_utils.GetParsedDataPaths(AcceptGrammar, data)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("get parsed data paths: %w", err), data)
	}
	if len(paths) == 0 {
		return nil, motmedelErrors.NewWithTrace(motmedelErrors.ErrSyntaxError, data)
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
					return nil, motmedelErrors.NewWithTrace(
						fmt.Errorf("%w: %w", motmedelErrors.ErrSemanticError, ErrParameterNamedQ),
					)
				}

				qValuePath := parsing_utils.SearchPathSingle(interestingPath, []string{"qvalue"}, 1, false)
				if qValuePath == nil {
					return nil, motmedelErrors.NewWithTrace(
						fmt.Errorf("%w: %w", motmedelErrors.ErrSemanticError, ErrNilQValuePath),
					)
				}

				qValueString := string(parsing_utils.ExtractPathValue(data, qValuePath))
				bitSize := 32
				parsedWeight, err := strconv.ParseFloat(qValueString, bitSize)
				if err != nil {
					return nil, motmedelErrors.NewWithTrace(
						fmt.Errorf(
							"%w: %w: strvconv parse float: %w",
							motmedelErrors.ErrSemanticError, ErrInvalidQvalue, err,
						),
						qValueString, bitSize,
					)
				}

				mediaRange.Weight = float32(parsedWeight)
			} else {
				parameterString := string(parsing_utils.ExtractPathValue(data, parameterPath))
				separator := "='"
				key, value, found := strings.Cut(parameterString, "=")
				if !found {
					return nil, motmedelErrors.NewWithTrace(
						fmt.Errorf("%w: %w", motmedelErrors.ErrSemanticError, ErrCouldNotSplitParameter),
						parameterString,
						separator,
					)
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
		panic(fmt.Sprintf("goabnf parse abnf (accept grammar): %v", err))
	}
}
