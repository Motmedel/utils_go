package utils_go

func _map[InputType any, OutputType any](
	inputSlice []InputType,
	outputSlice []OutputType,
	f func(InputType) OutputType,
) []OutputType {
	for i, inputValue := range inputSlice {
		outputSlice[i] = f(inputValue)
	}
	return outputSlice
}

func Map[InputType any, OutputType any](inputSlice []InputType, f func(InputType) OutputType) []OutputType {
	outputSlice := make([]OutputType, len(inputSlice), len(inputSlice))
	return _map[InputType, OutputType](inputSlice, outputSlice, f)
}

func Filter[T any](s []T, f func(T) bool) []T {
	var r []T
	for _, v := range s {
		if f(v) {
			r = append(r, v)
		}
	}
	return r
}

func MapFilter[InputType any, OutputType any](inputSlice []InputType, f func(InputType) OutputType) []OutputType {
	var outputSlice []OutputType
	return _map[InputType, OutputType](inputSlice, outputSlice, f)
}

func Set[T comparable](input []T) []T {
	uniqueMap := make(map[T]bool)
	var result []T
	for _, str := range input {
		if _, ok := uniqueMap[str]; !ok {
			uniqueMap[str] = true
			result = append(result, str)
		}
	}
	return result
}
