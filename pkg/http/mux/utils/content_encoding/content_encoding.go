package content_encoding

import (
	motmedelHttpTypes "github.com/Motmedel/utils_go/pkg/http/types"
	"strings"
)

const AcceptContentIdentityIdentifier = "identity"

func GetMatchingContentEncoding(
	acceptableEncodings []*motmedelHttpTypes.Encoding,
	supportedEncodings []string,
) string {
	if len(acceptableEncodings) == 0 {
		return AcceptContentIdentityIdentifier
	}

	disallowIdentity := false

	for _, acceptableEncoding := range acceptableEncodings {
		coding := strings.ToLower(acceptableEncoding.Coding)
		qualityValue := acceptableEncoding.QualityValue

		if coding == "*" {
			if qualityValue == 0 {
				disallowIdentity = true
			} else {
				if len(supportedEncodings) != 0 {
					return supportedEncodings[0]
				} else {
					if !disallowIdentity {
						return AcceptContentIdentityIdentifier
					}
				}
			}
		}

		if coding == AcceptContentIdentityIdentifier {
			if qualityValue == 0 {
				disallowIdentity = true
			} else {
				return AcceptContentIdentityIdentifier
			}
		}

		if qualityValue == 0 {
			continue
		}

		for _, supportedEncoding := range supportedEncodings {
			if acceptableEncoding.Coding == supportedEncoding {
				return supportedEncoding
			}
		}
	}

	if !disallowIdentity {
		return AcceptContentIdentityIdentifier
	} else {
		return ""
	}
}
