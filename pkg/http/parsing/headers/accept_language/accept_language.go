package accept_language

import (
	_ "embed"
	"errors"
	"fmt"
	"strconv"

	"github.com/Motmedel/parsing_utils/pkg/parsing_utils"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelHttpTypes "github.com/Motmedel/utils_go/pkg/http/types"
	goabnf "github.com/pandatix/go-abnf"
)

//go:embed grammar.txt
var grammar []byte

var Grammar *goabnf.Grammar

var (
	ErrNilAcceptLanguage = errors.New("nil accept language")
	ErrNilPrimarySubtag  = errors.New("nil primary subtag")
	ErrInvalidQvalue     = errors.New("invalid qvalue")
)

func Parse(data []byte) (*motmedelHttpTypes.AcceptLanguage, error) {
	paths, err := parsing_utils.GetParsedDataPaths(Grammar, data)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("get parsed data paths: %w", err), data)
	}
	if len(paths) == 0 {
		return nil, motmedelErrors.NewWithTrace(motmedelErrors.ErrSyntaxError, data)
	}

	var acceptLanguage motmedelHttpTypes.AcceptLanguage

	interestingPaths := parsing_utils.SearchPath(
		paths[0],
		[]string{"element"}, 2, false,
	)
	for _, interestingPath := range interestingPaths {
		primarySubtagPath := parsing_utils.SearchPathSingleName(
			interestingPath,
			"Primary-subtag",
			2,
			false,
		)
		if primarySubtagPath == nil {
			return nil, motmedelErrors.NewWithTrace(
				fmt.Errorf("%w: %w", motmedelErrors.ErrSemanticError, ErrNilPrimarySubtag),
			)
		}
		primarySubtag := parsing_utils.ExtractPathValue(data, primarySubtagPath)

		subtagPath := parsing_utils.SearchPathSingleName(
			interestingPath,
			"Subtag",
			2,
			false,
		)
		var subtag []byte
		if subtagPath != nil {
			subtag = parsing_utils.ExtractPathValue(data, subtagPath)
		}

		var qualityValue float32 = 1.0
		qvaluePath := parsing_utils.SearchPathSingleName(
			interestingPath,
			"qvalue",
			1,
			false,
		)
		if qvaluePath != nil {
			qvalueString := string(parsing_utils.ExtractPathValue(data, qvaluePath))
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
	Grammar, err = goabnf.ParseABNF(grammar)
	if err != nil {
		panic(fmt.Sprintf("goabnf parse abnf (accept encoding grammar): %v", err))
	}
}
