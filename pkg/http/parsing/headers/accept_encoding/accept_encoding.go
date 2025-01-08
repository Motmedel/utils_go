package accept_encoding

import (
	_ "embed"
	"errors"
	"github.com/Motmedel/parsing_utils/pkg/parsing_utils"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelHttpTypes "github.com/Motmedel/utils_go/pkg/http/types"
	goabnf "github.com/pandatix/go-abnf"
	"strconv"
)

//go:embed grammar.txt
var grammar []byte

var AcceptEncodingGrammar *goabnf.Grammar

var (
	ErrBadQvalue      = errors.New("qvalue could not be parsed as a float")
	ErrNilCodingsPath = errors.New("the codings path is nil")
)

type BadQvalueError struct {
	motmedelErrors.InputError
}

func (badQvalueError *BadQvalueError) Is(target error) bool {
	return target == ErrBadQvalue
}

func ParseAcceptEncoding(data []byte) (*motmedelHttpTypes.AcceptEncoding, error) {
	paths, err := goabnf.Parse(data, AcceptEncodingGrammar, "root")
	if err != nil {
		return nil, &motmedelErrors.InputError{
			Message: "An error occurred when parsing data as an accept encoding.",
			Cause:   err,
			Input:   data,
		}
	}
	if len(paths) == 0 {
		return nil, nil
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
			return nil, ErrNilCodingsPath
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
			parsedQualityValue, err := strconv.ParseFloat(qvalueString, 32)
			if err != nil {
				return nil, &BadQvalueError{
					InputError: motmedelErrors.InputError{
						Message: "A qvalue could not be parsed as a float.",
						Cause:   err,
						Input:   qvalueString,
					},
				}
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
		panic(err)
	}
}
