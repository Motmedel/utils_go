package dmarc

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Motmedel/parsing_utils/pkg/parsing_utils"
	dnsTypes "github.com/Motmedel/utils_go/pkg/dns/types"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
)

var (
	ErrUnexpectedKey   = errors.New("unexpected key")
	ErrMultipleSameKey = errors.New("multiple keys with the same name")
)

var keyValueNames = []string{
	"dmarc-request",
	"dmarc-srequest",
	"dmarc-auri",
	"dmarc-furi",
	"dmarc-adkim",
	"dmarc-aspf",
	"dmarc-ainterval",
	"dmarc-fo",
	"dmarc-rfmt",
	"dmarc-percent",
}

// caseInsensitiveTags lists tags whose values are case-insensitive per the
// DMARC ABNF. They are lowercased on parse so analysis can use exact
// comparisons.
var caseInsensitiveTags = map[string]bool{
	"p":     true,
	"sp":    true,
	"adkim": true,
	"aspf":  true,
	"fo":    true,
	"rf":    true,
}

func ParseDmarcRecord(data []byte) (*dnsTypes.DmarcRecord, error) {
	paths, err := parsing_utils.GetParsedDataPaths(DmarcGrammar, data)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("get parsed data paths: %w", err))
	}
	if len(paths) == 0 {
		return nil, motmedelErrors.NewWithTrace(motmedelErrors.ErrSyntaxError)
	}

	record := &dnsTypes.DmarcRecord{Raw: string(data)}
	fields := map[string]*string{
		"p":     &record.P,
		"sp":    &record.Sp,
		"rua":   &record.Rua,
		"ruf":   &record.Ruf,
		"adkim": &record.Adkim,
		"aspf":  &record.Aspf,
		"ri":    &record.Ri,
		"fo":    &record.Fo,
		"rf":    &record.Rf,
		"pct":   &record.Pct,
	}

	for _, termPath := range parsing_utils.SearchPath(paths[0], keyValueNames, 1, false) {
		keyValuePair := strings.SplitN(string(parsing_utils.ExtractPathValue(data, termPath)), "=", 2)
		key := strings.ToLower(strings.TrimSpace(keyValuePair[0]))
		value := strings.TrimSpace(keyValuePair[1])

		field, ok := fields[key]
		if !ok {
			return nil, motmedelErrors.NewWithTrace(
				fmt.Errorf("%w: %w: %s", motmedelErrors.ErrSemanticError, ErrUnexpectedKey, key),
				key,
			)
		}

		if *field != "" {
			return nil, motmedelErrors.NewWithTrace(
				fmt.Errorf("%w: %w: %s", motmedelErrors.ErrSemanticError, ErrMultipleSameKey, key),
				key,
			)
		}

		if caseInsensitiveTags[key] {
			value = strings.ToLower(value)
		}
		*field = value
	}

	return record, nil
}
