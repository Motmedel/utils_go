package schema

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Motmedel/jsonschema"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	jsonSchemaStruct "github.com/swaggest/jsonschema-go"
)

var (
	ErrNilCompiler = errors.New("nil compiler")
)

func New[T any](t T) (*jsonschema.Schema, error) {
	var reflector jsonSchemaStruct.Reflector

	structSchema, err := reflector.Reflect(t)
	if err != nil {
		return nil, motmedelErrors.NewWithTrace(fmt.Errorf("reflect: %w", err), t)
	}

	byteSchema, err := json.Marshal(structSchema)
	if err != nil {
		return nil, motmedelErrors.NewWithTrace(fmt.Errorf("json marshal: %w", err), t)
	}

	compiler := jsonschema.NewCompiler()
	if compiler == nil {
		return nil, motmedelErrors.MakeErrorWithStackTrace(ErrNilCompiler)
	}

	schema, err := compiler.Compile(byteSchema)
	if err != nil {
		return nil, motmedelErrors.NewWithTrace(fmt.Errorf("compile: %w", err), byteSchema)
	}

	return schema, nil
}
