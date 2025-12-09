package retry_after

import (
	_ "embed"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/Motmedel/parsing_utils/pkg/parsing_utils"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelHttpTypes "github.com/Motmedel/utils_go/pkg/http/types"
	goabnf "github.com/pandatix/go-abnf"
)

//go:embed grammar.txt
var grammar []byte

var RetryAfterGrammar *goabnf.Grammar

var (
	ErrNilRetryAfter       = errors.New("nil retry after")
	ErrEmptyHttpDate       = errors.New("empty http date")
	ErrInvalidHttpDate     = errors.New("invalid http date")
	ErrEmptyDelaySeconds   = errors.New("empty delay seconds")
	ErrInvalidDelaySeconds = errors.New("invalid delay seconds")
	ErrNoPathMatch         = errors.New("neither HTTP-date or delay-seconds matched")
)

func Parse(data []byte) (*motmedelHttpTypes.RetryAfter, error) {
	paths, err := parsing_utils.GetParsedDataPaths(RetryAfterGrammar, data)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("get parsed data paths: %w", err), data)
	}
	if len(paths) == 0 {
		return nil, motmedelErrors.NewWithTrace(motmedelErrors.ErrSyntaxError, data)
	}

	retryAfter := &motmedelHttpTypes.RetryAfter{Raw: string(data)}

	path := paths[0]

	httpDatePath := parsing_utils.SearchPathSingleName(path, "HTTP-date", 2, false)
	if httpDatePath != nil {
		httpDateString := string(parsing_utils.ExtractPathValue(data, httpDatePath))
		if httpDateString == "" {
			return nil, motmedelErrors.NewWithTrace(
				fmt.Errorf("%w: %w", motmedelErrors.ErrSemanticError, ErrEmptyHttpDate),
			)
		}

		httpDate, err := time.Parse(time.RFC1123, httpDateString)
		if err != nil {
			return nil, motmedelErrors.NewWithTrace(
				fmt.Errorf(
					"%w: %w: time parse rfc1123: %w",
					motmedelErrors.ErrSemanticError,
					ErrInvalidHttpDate,
					err,
				),
				httpDateString,
			)
		}

		retryAfter.WaitTime = httpDate

		return retryAfter, nil
	}

	delaySecondsPath := parsing_utils.SearchPathSingleName(path, "delay-seconds", 2, false)
	if delaySecondsPath != nil {
		delaySecondsString := string(parsing_utils.ExtractPathValue(data, delaySecondsPath))
		if delaySecondsString == "" {
			return nil, motmedelErrors.NewWithTrace(
				fmt.Errorf("%w: %w", motmedelErrors.ErrSemanticError, ErrEmptyDelaySeconds),
			)
		}

		delaySeconds, err := strconv.Atoi(delaySecondsString)
		if err != nil {
			return nil, motmedelErrors.NewWithTrace(
				fmt.Errorf(
					"%w: %w: strconv atoi: %w",
					motmedelErrors.ErrSemanticError,
					ErrInvalidDelaySeconds,
					err,
				),
				delaySecondsString,
			)
		}

		retryAfter.WaitTime = time.Duration(delaySeconds) * time.Second

		return retryAfter, nil
	}

	return nil, motmedelErrors.NewWithTrace(ErrNoPathMatch)
}

func init() {
	var err error
	RetryAfterGrammar, err = goabnf.ParseABNF(grammar)
	if err != nil {
		panic(fmt.Sprintf("goabnf parse abnf (retry after grammar): %v", err))
	}
}
