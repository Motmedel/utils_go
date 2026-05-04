package spf

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/Motmedel/parsing_utils/pkg/parsing_utils"
	dnsTypes "github.com/Motmedel/utils_go/pkg/dns/types"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelNet "github.com/Motmedel/utils_go/pkg/net"
)

var ErrUnexpectedMatchRule = errors.New("unexpected matching rule")

func extractLabelValues[T dnsTypes.SpfTermPtr](record *dnsTypes.SpfRecord, passOnly bool, labels ...string) []string {
	var values []string

	for _, term := range record.Terms {
		if _, ok := term.(T); ok {
			switch typedTerm := term.(type) {
			case *dnsTypes.SpfDirective:
				for _, label := range labels {
					if passOnly && (typedTerm.Qualifier != "" && typedTerm.Qualifier != "+") {
						continue
					}

					if strings.ToLower(typedTerm.Mechanism.Label) == label {
						values = append(values, typedTerm.Mechanism.Value)
					}
				}
			case *dnsTypes.SpfModifier:
				for _, label := range labels {
					if strings.ToLower(typedTerm.Label) == label {
						values = append(values, typedTerm.Value)
					}
				}
			}
		}
	}

	return values
}

func ExtractIncludeValues(record *dnsTypes.SpfRecord) []string {
	if record == nil {
		return nil
	}

	return extractLabelValues[*dnsTypes.SpfDirective](record, false, "include")
}

func ExtractRedirectValues(record *dnsTypes.SpfRecord) []string {
	if record == nil {
		return nil
	}

	return extractLabelValues[*dnsTypes.SpfModifier](record, false, "redirect")
}

func ExtractNetworks(record *dnsTypes.SpfRecord, passOnly bool) []*net.IPNet {
	if record == nil {
		return nil
	}

	var networks []*net.IPNet

	for _, networkString := range extractLabelValues[*dnsTypes.SpfDirective](record, passOnly, "ip4", "ip6") {
		if network, _ := motmedelNet.ParseAddressNet(networkString); network != nil {
			networks = append(networks, network)
		}
	}

	return networks
}

func ParseSpfRecord(data []byte) (*dnsTypes.SpfRecord, error) {
	paths, err := parsing_utils.GetParsedDataPaths(SpfGrammar, data)
	if err != nil {
		return nil, fmt.Errorf("get parsed data paths: %w", err)
	}
	if len(paths) == 0 {
		return nil, motmedelErrors.NewWithTrace(motmedelErrors.ErrSyntaxError)
	}

	var record dnsTypes.SpfRecord
	record.Raw = string(data)

	var terms []any
	for i, termPath := range parsing_utils.SearchPath(paths[0], []string{"directive", "modifier"}, 2, false) {
		switch matchRule := termPath.MatchRule; matchRule {
		case "directive":
			directive := dnsTypes.SpfDirective{Index: i}

			qualifierPath := parsing_utils.SearchPathSingleName(termPath, "qualifier", 1, false)
			if qualifierPath != nil {
				directive.Qualifier = string(parsing_utils.ExtractPathValue(data, qualifierPath))
			}
			mechanismPath := parsing_utils.SearchPathSingleName(termPath, "mechanism", 1, false)
			if mechanismPath != nil {
				mechanismPair := strings.SplitN(string(parsing_utils.ExtractPathValue(data, mechanismPath)), ":", 2)
				directive.Mechanism = &dnsTypes.SpfMechanism{Label: mechanismPair[0]}
				if len(mechanismPair) == 2 {
					directive.Mechanism.Value = mechanismPair[1]
				}
			}

			terms = append(terms, &directive)
		case "modifier":
			modifierPair := strings.SplitN(string(parsing_utils.ExtractPathValue(data, termPath)), "=", 2)
			// NOTE: According to the grammar there should always be two elements.
			terms = append(terms, &dnsTypes.SpfModifier{Index: i, Label: modifierPair[0], Value: modifierPair[1]})
		default:
			return nil, motmedelErrors.NewWithTrace(
				fmt.Errorf("%w: %w: %s", motmedelErrors.ErrSemanticError, ErrUnexpectedMatchRule, matchRule),
			)
		}
	}

	record.Terms = terms

	return &record, nil
}
