package utils_go

import "reflect"

func Map[InputType any, OutputType any](inputSlice []InputType, f func(InputType) OutputType) []OutputType {
	outputSlice := make([]OutputType, len(inputSlice))
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

func MapFilter[InputType any, OutputType any](inputSlice []InputType, f func(InputType) OutputType) []OutputType {
	var outputSlice []OutputType
	for _, inputValue := range inputSlice {
		if outputValue := f(inputValue); !reflect.ValueOf(outputValue).IsZero() {
			outputSlice = append(outputSlice, outputValue)
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
