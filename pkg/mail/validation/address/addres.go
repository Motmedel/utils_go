package address

import (
	"fmt"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelMailValidationAddressErrors "github.com/Motmedel/utils_go/pkg/mail/validation/address/errors"
	"github.com/Motmedel/utils_go/pkg/net/domain_breakdown"
	motmedelNetErrors "github.com/Motmedel/utils_go/pkg/net/errors"
	"net/mail"
	"strings"
)

func Validate(addressString string) error {
	trimmedAddressString := strings.TrimSpace(addressString)
	if trimmedAddressString == "" {
		return fmt.Errorf(
			"%w: %w",
			motmedelErrors.ErrValidationError,
			motmedelMailValidationAddressErrors.ErrEmptyAddress,
		)
	}

	address, err := mail.ParseAddress(trimmedAddressString)
	if err != nil {
		return fmt.Errorf("%w: parse address: %w", motmedelErrors.ErrValidationError, err)
	}

	exactAddress := address.Name == "" && address.Address == trimmedAddressString
	if !exactAddress {
		return fmt.Errorf(
			"%w: %w",
			motmedelErrors.ErrValidationError,
			motmedelMailValidationAddressErrors.ErrAddressMismatch,
		)
	}

	_, domain, found := strings.Cut(trimmedAddressString, "@")
	if !found {
		return fmt.Errorf("%w: %w", motmedelErrors.ErrValidationError, motmedelErrors.ErrBadSplit)
	}

	domainBreakdown := domain_breakdown.GetDomainBreakdown(domain)
	if domainBreakdown == nil {
		return fmt.Errorf("%w: %w", motmedelErrors.ErrValidationError, motmedelNetErrors.ErrNilDomainBreakdown)
	}

	return nil
}
