package generic_type_info

import "github.com/Motmedel/utils_go/pkg/type_generation/types/shape"

type GenericTypeInfo struct {
	TypeParameterNames           []string
	FieldNameToShapes            map[string][]shape.Shape
	TypeParameterNameToFieldName map[string]string
}
