package utils_go

func Map[InputType any, OutputType any](inputSlice []InputType, f func(InputType) OutputType) []OutputType {
	outputSlice := make([]OutputType, len(inputSlice), len(inputSlice))
	for i, inputValue := range inputSlice {
		outputSlice[i] = f(inputValue)
	}
	return outputSlice
}

func Filter[T any](inputSlice []T, f func(T) bool) []T {
	var outputSlice []T
	for _, inputValue := range inputSlice {
		if f(inputValue) {
			outputSlice = append(outputSlice, inputValue)
		}
	}
	return outputSlice
}

func MapFilter[InputType any, OutputType any](inputSlice []InputType, f func(InputType) *OutputType) []OutputType {
	var outputSlice []OutputType
	for i, inputValue := range inputSlice {
		if outputValue := f(inputValue); outputValue != nil {
			outputSlice[i] = *outputValue
		}
	}
	return outputSlice
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
