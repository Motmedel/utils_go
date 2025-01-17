package retry_after

import (
	_ "embed"
	"errors"
	"fmt"
	"github.com/Motmedel/parsing_utils/pkg/parsing_utils"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	goabnf "github.com/pandatix/go-abnf"
	"strconv"
	"time"
)

//go:embed grammar.txt
var grammar []byte

var RetryAfterGrammar *goabnf.Grammar

var (
	ErrEmptyHttpDate     = errors.New("empty http date")
	ErrEmptyDelaySeconds = errors.New("empty delay seconds")
	ErrNoPathMatch       = errors.New("neither HTTP-date or delay-seconds matched")
)

func ParseRetryAfter(data []byte) (*RetryAfter, error) {
	paths, err := goabnf.Parse(data, RetryAfterGrammar, "root")
	if err != nil {
		return nil, &motmedelErrors.InputError{
			Message: "An error occurred when parsing data as retry after.",
			Cause:   err,
			Input:   data,
		}
	}
	if len(paths) == 0 {
		return nil, nil
	}

	retryAfter := &RetryAfter{Raw: string(data)}

	path := paths[0]

	httpDatePath := parsing_utils.SearchPathSingleName(path, "HTTP-date", 2, false)
	if httpDatePath != nil {
		httpDateString := string(parsing_utils.ExtractPathValue(data, httpDatePath))
		if httpDateString == "" {
			return nil, ErrEmptyHttpDate
		}

		httpDate, err := time.Parse(time.RFC1123, httpDateString)
		if err != nil {
			return nil, &motmedelErrors.InputError{
				Message: "An error occurred when parsing an http date string as a RFC1123 timestamp.",
				Cause:   err,
				Input:   httpDateString,
			}
		}

		retryAfter.WaitTime = httpDate

		return retryAfter, nil
	}

	delaySecondsPath := parsing_utils.SearchPathSingleName(path, "delay-seconds", 2, false)
	if delaySecondsPath != nil {
		delaySecondsString := string(parsing_utils.ExtractPathValue(data, delaySecondsPath))
		if delaySecondsString == "" {
			return nil, ErrEmptyDelaySeconds
		}

		delaySeconds, err := strconv.Atoi(delaySecondsString)
		if err != nil {
			return nil, &motmedelErrors.InputError{
				Message: "An error occurred when parsing a delay seconds string as an integer.",
				Cause:   err,
				Input:   delaySecondsString,
			}
		}

		retryAfter.WaitTime = delaySeconds

		return retryAfter, nil
	}

	return nil, ErrNoPathMatch
}

func init() {
	var err error
	RetryAfterGrammar, err = goabnf.ParseABNF(grammar)
	if err != nil {
		panic(fmt.Sprintf("an error occurred when parsing Retry-After grammar: %v", err))
	}
}
