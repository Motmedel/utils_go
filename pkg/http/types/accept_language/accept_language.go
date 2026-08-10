package accept_language

import (
	_ "embed"
	"errors"
	"fmt"
	"strconv"

	"github.com/Motmedel/utils_go/pkg/errors/types/nil_error"

	"github.com/Motmedel/utils_go/pkg/abnf"
	abnfUtils "github.com/Motmedel/utils_go/pkg/abnf/utils"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelHttpTypes "github.com/Motmedel/utils_go/pkg/http/types"
)

//go:embed grammar.abnf
var grammar []byte

var Grammar *abnf.Grammar

var (
	ErrInvalidQvalue = errors.New("invalid qvalue")
)

func Parse(data []byte) (*motmedelHttpTypes.AcceptLanguage, error) {
	paths, err := abnfUtils.GetParsedDataPaths(Grammar, data)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("get parsed data paths: %w", err), data)
	}
	if len(paths) == 0 {
		return nil, motmedelErrors.NewWithTrace(motmedelErrors.ErrSyntaxError, data)
	}

	var acceptLanguage motmedelHttpTypes.AcceptLanguage

	interestingPaths := abnfUtils.SearchPath(
		paths[0],
		[]string{"element"}, 2, false,
	)
	for _, interestingPath := range interestingPaths {
		primarySubtagPath := abnfUtils.SearchPathSingleName(
			interestingPath,
			"Primary-subtag",
			2,
			false,
		)
		if primarySubtagPath == nil {
			return nil, motmedelErrors.NewWithTrace(
				fmt.Errorf("%w: %w", motmedelErrors.ErrSemanticError, nil_error.New("primary subtag")),
			)
		}
		primarySubtag := abnfUtils.ExtractPathValue(data, primarySubtagPath)

		subtagPath := abnfUtils.SearchPathSingleName(
			interestingPath,
			"Subtag",
			2,
			false,
		)
		var subtag []byte
		if subtagPath != nil {
			subtag = abnfUtils.ExtractPathValue(data, subtagPath)
		}

		var qualityValue float32 = 1.0
		qvaluePath := abnfUtils.SearchPathSingleName(
			interestingPath,
			"qvalue",
			1,
			false,
		)
		if qvaluePath != nil {
			qvalueString := string(abnfUtils.ExtractPathValue(data, qvaluePath))
			bitSize := 32
			parsedQualityValue, err := strconv.ParseFloat(qvalueString, bitSize)
			if err != nil {
				return nil, motmedelErrors.NewWithTrace(
					fmt.Errorf(
						"%w: %w: strvconv parse float (qvalue): %w",
						motmedelErrors.ErrSemanticError,
						ErrInvalidQvalue,
						err,
					),
					qvaluePath, bitSize,
				)
			}

			qualityValue = float32(parsedQualityValue)
		}

		acceptLanguage.LanguageQs = append(
			acceptLanguage.LanguageQs,
			&motmedelHttpTypes.LanguageQ{
				Tag: &motmedelHttpTypes.LanguageTag{
					PrimarySubtag: string(primarySubtag),
					Subtag:        string(subtag),
				},
				Q: qualityValue,
			},
		)
	}

	acceptLanguage.Raw = string(data)

	return &acceptLanguage, nil
}

func init() {
	var err error
	Grammar, err = abnf.ParseABNF(grammar)
	if err != nil {
		panic(fmt.Sprintf("goabnf parse abnf (accept encoding grammar): %v", err))
	}
}
