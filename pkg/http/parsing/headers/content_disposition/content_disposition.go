package content_disposition

import (
	_ "embed"
	"errors"
	"fmt"
	"github.com/Motmedel/parsing_utils/pkg/parsing_utils"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelHttpTypes "github.com/Motmedel/utils_go/pkg/http/types"
	goabnf "github.com/pandatix/go-abnf"
	"strconv"
	"strings"
)

var ContentDispositionGrammar *goabnf.Grammar

//go:embed grammar.txt
var grammar []byte

var (
	ErrSyntaxError                        = errors.New("syntax error")
	ErrSemanticError                      = errors.New("semantic error")
	ErrNilGrammar                         = errors.New("nil grammar")
	ErrNoFilenameLabel                    = errors.New("no filename label")
	ErrNilFilenameLabelPath               = errors.New("nil filename label path")
	ErrNilExtensionLabelPath              = errors.New("nil extension label path")
	ErrNilFilenameValuePath               = errors.New("nil filename value path")
	ErrNilFilenameExtValuePath            = errors.New("nil filename ext value path")
	ErrNotEnoughExtensionSubpaths         = errors.New("not enough extension subpaths")
	ErrNilExtensionValuePath              = errors.New("nil extension value path")
	ErrEmptyExtensionLabel                = errors.New("empty extension label")
	ErrEmptyExtensionValue                = errors.New("empty extension value")
	ErrDuplicateLabel                     = errors.New("duplicate label")
	ErrUnexpectedInterestingPathMatchRule = errors.New("unexpected interesting path match rule")
	ErrUnexpectedFilenameLabel            = errors.New("unexpected filename label")
)

type SemanticError struct {
	motmedelErrors.InputError
}

func (semanticError *SemanticError) Is(target error) bool {
	// TODO: Not sure if this works?
	_, ok := target.(*SemanticError)
	return ok
}

func (semanticError *SemanticError) Error() string {
	return ErrSemanticError.Error()
}

func getValue(data []byte, path *goabnf.Path) (string, error) {
	if path == nil {
		return "", nil
	}

	var value string

	quotedStringPath := parsing_utils.SearchPathSingleName(
		path,
		"quoted-string",
		1,
		false,
	)
	if quotedStringPath != nil {
		var err error
		quotedString := string(parsing_utils.ExtractPathValue(data, quotedStringPath))
		value, err = strconv.Unquote(quotedString)
		if err != nil {
			return "", &motmedelErrors.InputError{
				Message: "An error occurred when unquoting a quoted-string.",
				Cause:   err,
				Input:   quotedString,
			}
		}
	} else {
		value = string(parsing_utils.ExtractPathValue(data, path))
	}

	return value, nil
}

func ParseContentDisposition(data []byte) (*motmedelHttpTypes.ContentDisposition, error) {
	if len(data) == 0 {
		return nil, nil
	}

	if ContentDispositionGrammar == nil {
		return nil, ErrNilGrammar
	}

	paths, err := goabnf.Parse(data, ContentDispositionGrammar, "root")
	if err != nil {
		return nil, &motmedelErrors.InputError{
			Message: "An error occurred when parsing data as a content disposition.",
			Cause:   err,
			Input:   data,
		}
	}
	if len(paths) == 0 {
		return nil, ErrSyntaxError
	}

	contentDisposition := motmedelHttpTypes.ContentDisposition{
		ExtensionParameters: make(map[string]string),
	}

	interestingPaths := parsing_utils.SearchPath(
		paths[0],
		[]string{"disposition-type", "filename-parm", "disp-ext-parm"},
		4,
		false,
	)

	for _, interestingPath := range interestingPaths {

		interestingPathMatchRule := interestingPath.MatchRule
		switch interestingPathMatchRule {
		case "disposition-type":
			contentDisposition.DispositionType = strings.ToLower(string(parsing_utils.ExtractPathValue(data, interestingPath)))
		case "filename-parm":
			subpaths := interestingPath.Subpaths
			if len(subpaths) < 1 {
				return nil, &SemanticError{
					InputError: motmedelErrors.InputError{
						Message: "No label was found for a filename parameter.",
						Cause:   ErrNoFilenameLabel,
						Input:   subpaths,
					},
				}
			}

			labelPath := subpaths[0]
			if labelPath == nil {
				return nil, &SemanticError{
					InputError: motmedelErrors.InputError{
						Message: "A filename label path is nil.",
						Cause:   ErrNilFilenameLabelPath,
					},
				}
			}

			filenameLabel := strings.ToLower(string(parsing_utils.ExtractPathValue(data, subpaths[0])))
			switch filenameLabel {
			case "filename":
				if contentDisposition.FilenameParameter != "" {
					return nil, &SemanticError{
						InputError: motmedelErrors.InputError{
							Message: fmt.Sprintf("A duplicate %s label was observed.", filenameLabel),
							Cause:   ErrDuplicateLabel,
							Input:   filenameLabel,
						},
					}
				}

				filenameValuePath := parsing_utils.SearchPathSingle(
					interestingPath,
					[]string{"value"},
					1,
					false,
				)
				if filenameValuePath == nil {
					return nil, &SemanticError{
						InputError: motmedelErrors.InputError{
							Message: "No value path was found for the filename parameter.",
							Cause:   ErrNilFilenameValuePath,
						},
					}
				}

				value, err := getValue(data, filenameValuePath)
				if err != nil {
					return nil, &motmedelErrors.InputError{
						Message: "An error occurred when obtaining a parameter value.",
						Cause:   err,
						Input:   filenameValuePath,
					}
				}

				contentDisposition.FilenameParameter = value
			case "filename*":
				if contentDisposition.FilenameParameterAsterisk != "" {
					return nil, &SemanticError{
						InputError: motmedelErrors.InputError{
							Message: fmt.Sprintf("A duplicate %s label was observed.", filenameLabel),
							Cause:   ErrDuplicateLabel,
							Input:   filenameLabel,
						},
					}
				}

				filenameAsteriskExtValuePath := parsing_utils.SearchPathSingle(
					interestingPath,
					[]string{"ext-value"},
					1,
					false,
				)
				if filenameAsteriskExtValuePath == nil {
					return nil, &SemanticError{
						InputError: motmedelErrors.InputError{
							Message: "No value path was found for the filename* parameter.",
							Cause:   ErrNilFilenameExtValuePath,
						},
					}
				}

				contentDisposition.FilenameParameterAsterisk = string(parsing_utils.ExtractPathValue(data, filenameAsteriskExtValuePath))
			default:
				return nil, &motmedelErrors.InputError{
					Message: "An unexpected filename label was observed.",
					Cause:   ErrUnexpectedFilenameLabel,
					Input:   filenameLabel,
				}
			}
		case "disp-ext-parm":
			subpaths := interestingPath.Subpaths
			if len(subpaths) != 3 {
				return nil, &SemanticError{
					InputError: motmedelErrors.InputError{
						Message: "Not enough extension subpaths are present.",
						Cause:   ErrNotEnoughExtensionSubpaths,
						Input:   subpaths,
					},
				}
			}

			labelPath := subpaths[0]
			if labelPath == nil {
				return nil, &SemanticError{
					InputError: motmedelErrors.InputError{
						Message: "An extension label path is nil.",
						Cause:   ErrNilExtensionLabelPath,
					},
				}
			}
			label := strings.ToLower(string(parsing_utils.ExtractPathValue(data, labelPath)))
			if label == "" {
				return nil, &SemanticError{
					InputError: motmedelErrors.InputError{
						Message: "An extension label is empty.",
						Cause:   ErrEmptyExtensionLabel,
					},
				}
			}

			if _, ok := contentDisposition.ExtensionParameters[label]; ok {
				return nil, &SemanticError{
					InputError: motmedelErrors.InputError{
						Message: fmt.Sprintf("A duplicate %s label was observed.", label),
						Cause:   ErrDuplicateLabel,
						Input:   label,
					},
				}
			}

			valuePath := subpaths[2]
			if valuePath == nil {
				return nil, &SemanticError{
					InputError: motmedelErrors.InputError{
						Message: "An extension value path is nil.",
						Cause:   ErrNilExtensionValuePath,
					},
				}
			}

			value, err := getValue(data, valuePath)
			if err != nil {
				return nil, &motmedelErrors.InputError{
					Message: "An error occurred when obtaining a parameter value.",
					Cause:   err,
					Input:   valuePath,
				}
			}

			if value == "" {
				return nil, &SemanticError{
					InputError: motmedelErrors.InputError{
						Message: "An extension value is empty.",
						Cause:   ErrEmptyExtensionValue,
					},
				}
			}

			contentDisposition.ExtensionParameters[label] = value
		default:
			return nil, &motmedelErrors.InputError{
				Message: "An unexpected interesting path match rule was observed.",
				Cause:   ErrUnexpectedInterestingPathMatchRule,
				Input:   interestingPathMatchRule,
			}
		}
	}

	if len(contentDisposition.ExtensionParameters) == 0 {
		contentDisposition.ExtensionParameters = nil
	}

	return &contentDisposition, nil
}

func init() {
	var err error
	ContentDispositionGrammar, err = goabnf.ParseABNF(grammar)
	if err != nil {
		panic(fmt.Sprintf("An error occurred when parsing the grammar: %v", err))
	}
}
