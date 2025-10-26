package reflect

import (
	"reflect"
	"strings"
)

func RemoveIndirection(reflectType reflect.Type) reflect.Type {
	kind := reflectType.Kind()
	for kind == reflect.Ptr {
		reflectType = reflectType.Elem()
		kind = reflectType.Kind()
	}
	return reflectType
}

func GetTypeName(t reflect.Type) (string, bool) {
	name := t.Name()

	idx := strings.Index(name, "[")
	if idx == -1 {
		return name, false
	}

	// The type uses generics. Remove the bracket part to get the name.
	return name[:idx], true
}
