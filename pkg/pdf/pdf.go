package pdf

import (
	"bytes"

	"github.com/Motmedel/utils_go/pkg/types/file_validator"
)

func IsSigned(documentData []byte) bool {
	return bytes.Contains(documentData, []byte("/Type /Sig")) && bytes.Contains(documentData, []byte("/ByteRange"))
}

func NewPdfFileValidator() *file_validator.Validator {
	return &file_validator.Validator{
		ExpectedFileExtension: ".pdf",
		ExpectedContentType:   "application/pdf",
	}
}
