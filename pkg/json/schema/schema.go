package schema

import (
	"fmt"

	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelReflect "github.com/Motmedel/utils_go/pkg/reflect"
	"github.com/altshiftab/jsonschema/pkg/jsonschema"
	jsonschemaTypeGeneration "github.com/vphpersson/type_generation/pkg/producers/jsonschema"
)

func New[T any]() (*jsonschema.Schema, error) {
	schemaData, err := jsonschemaTypeGeneration.Convert(motmedelReflect.TypeOf[T]())
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("jsonschema convert: %w", err))
	}

	schema, err := jsonschema.New([]byte(schemaData))
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("jsonschema new: %w", err))
	}

	return schema, nil
}
