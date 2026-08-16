package errors

import "errors"

var (
	ErrOddDictArguments             = errors.New("dict expects an even number of arguments")
	ErrNonStringDictKey             = errors.New("dict keys must be strings")
	ErrBodylessMethodContentType    = errors.New("the content type is not supported for a method that takes no body")
	ErrUnsupportedOutputContentType = errors.New("the output content type is not supported")
	ErrBinaryOutputWithOutputType   = errors.New("an output type cannot be combined with a binary output content type")
	ErrOptionalBinaryOutput         = errors.New("an optional output is not supported for a binary output content type")
)
