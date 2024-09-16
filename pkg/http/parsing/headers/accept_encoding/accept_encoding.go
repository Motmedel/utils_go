package accept_encoding

import (
	"errors"
	"github.com/Motmedel/parsing_utils/parsing_utils"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelHttpTypes "github.com/Motmedel/utils_go/pkg/http/types"
	goabnf "github.com/pandatix/go-abnf"
	"strconv"
)

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

/*
root=Accept-Encoding
Accept-Encoding = [ element ] *( OWS "," OWS [ element ] )
element = codings [ weight ]
codings = content-coding / "identity" / "*"
content-coding=token
token=1*tchar
tchar="!"/"#"/"$"/"%"/"&"/"'"/"*"/"+"/"-"/"."/"^"/"_"/"`"/"|"/"~"/DIGIT/ALPHA
weight = OWS ";" OWS "q=" qvalue
OWS=*(SP/HTAB)
qvalue = ( "0" [ "." 0*3DIGIT ] ) / ( "1" [ "." 0*3("0") ] )
*/

var grammar = []uint8{114, 111, 111, 116, 61, 65, 99, 99, 101, 112, 116, 45, 69, 110, 99, 111, 100, 105, 110, 103, 13, 10, 65, 99, 99, 101, 112, 116, 45, 69, 110, 99, 111, 100, 105, 110, 103, 32, 61, 32, 91, 32, 101, 108, 101, 109, 101, 110, 116, 32, 93, 32, 42, 40, 32, 79, 87, 83, 32, 34, 44, 34, 32, 79, 87, 83, 32, 91, 32, 101, 108, 101, 109, 101, 110, 116, 32, 93, 32, 41, 13, 10, 101, 108, 101, 109, 101, 110, 116, 32, 61, 32, 99, 111, 100, 105, 110, 103, 115, 32, 91, 32, 119, 101, 105, 103, 104, 116, 32, 93, 13, 10, 99, 111, 100, 105, 110, 103, 115, 32, 61, 32, 99, 111, 110, 116, 101, 110, 116, 45, 99, 111, 100, 105, 110, 103, 32, 47, 32, 34, 105, 100, 101, 110, 116, 105, 116, 121, 34, 32, 47, 32, 34, 42, 34, 13, 10, 99, 111, 110, 116, 101, 110, 116, 45, 99, 111, 100, 105, 110, 103, 61, 116, 111, 107, 101, 110, 13, 10, 116, 111, 107, 101, 110, 61, 49, 42, 116, 99, 104, 97, 114, 13, 10, 116, 99, 104, 97, 114, 61, 34, 33, 34, 47, 34, 35, 34, 47, 34, 36, 34, 47, 34, 37, 34, 47, 34, 38, 34, 47, 34, 39, 34, 47, 34, 42, 34, 47, 34, 43, 34, 47, 34, 45, 34, 47, 34, 46, 34, 47, 34, 94, 34, 47, 34, 95, 34, 47, 34, 96, 34, 47, 34, 124, 34, 47, 34, 126, 34, 47, 68, 73, 71, 73, 84, 47, 65, 76, 80, 72, 65, 13, 10, 119, 101, 105, 103, 104, 116, 32, 61, 32, 79, 87, 83, 32, 34, 59, 34, 32, 79, 87, 83, 32, 34, 113, 61, 34, 32, 113, 118, 97, 108, 117, 101, 13, 10, 79, 87, 83, 61, 42, 40, 83, 80, 47, 72, 84, 65, 66, 41, 13, 10, 113, 118, 97, 108, 117, 101, 32, 61, 32, 40, 32, 34, 48, 34, 32, 91, 32, 34, 46, 34, 32, 48, 42, 51, 68, 73, 71, 73, 84, 32, 93, 32, 41, 32, 47, 32, 40, 32, 34, 49, 34, 32, 91, 32, 34, 46, 34, 32, 48, 42, 51, 40, 34, 48, 34, 41, 32, 93, 32, 41, 13, 10}

func init() {
	var err error
	AcceptEncodingGrammar, err = goabnf.ParseABNF(grammar)
	if err != nil {
		panic(err)
	}
}
