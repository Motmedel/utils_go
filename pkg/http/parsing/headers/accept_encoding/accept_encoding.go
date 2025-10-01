package accept_encoding

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

var AcceptEncodingGrammar *goabnf.Grammar

var (
	ErrNilAcceptEncoding = errors.New("nil accept encoding")
	ErrInvalidQvalue     = errors.New("invalid qvalue")
	ErrNilCodingsPath    = errors.New("nil codings path")
)

func ParseAcceptEncoding(data []byte) (*motmedelHttpTypes.AcceptEncoding, error) {
	paths, err := parsing_utils.GetParsedDataPaths(AcceptEncodingGrammar, data)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("get parsed data paths: %w", err), data)
	}
	if len(paths) == 0 {
		return nil, motmedelErrors.NewWithTrace(motmedelErrors.ErrSyntaxError, data)
	}

	var acceptEncoding motmedelHttpTypes.AcceptEncoding

	interestingPaths := parsing_utils.SearchPath(
		paths[0],
		[]string{"element"}, 2, false,
	)
	for _, interestingPath := range interestingPaths {
		codingsPath := parsing_utils.SearchPathSingleName(
			interestingPath,
			"codings",
			1,
			false,
		)
		if codingsPath == nil {
			return nil, motmedelErrors.NewWithTrace(
				fmt.Errorf("%w: %w", motmedelErrors.ErrSemanticError, ErrNilCodingsPath),
			)
		}
		codingsValue := parsing_utils.ExtractPathValue(data, codingsPath)

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

		acceptEncoding.Encodings = append(
			acceptEncoding.Encodings,
			&motmedelHttpTypes.Encoding{Coding: string(codingsValue), QualityValue: qualityValue},
		)
	}

	acceptEncoding.Raw = string(data)

	return &acceptEncoding, nil
}

func init() {
	var err error
	AcceptEncodingGrammar, err = goabnf.ParseABNF(grammar)
	if err != nil {
		panic(fmt.Sprintf("goabnf parse abnf (accept encoding grammar): %v", err))
	}
}
