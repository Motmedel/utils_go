package iter

import (
	"iter"
	"maps"
	"reflect"
	"slices"
)

func Concat[V any](sequences ...iter.Seq[V]) iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, sequence := range sequences {
			for element := range sequence {
				if !yield(element) {
					return
				}
			}
		}
	}
}

func Concat2[K any, V any](sequences ...iter.Seq2[K, V]) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for _, sequence := range sequences {
			for kElement, vElement := range sequence {
				if !yield(kElement, V(vElement)) {
					return
				}
			}
		}
	}
}

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

func Set[T comparable](elements []T) []T {
	if len(elements) == 0 {
		return nil
	}

	setMap := make(map[T]struct{})
	for _, element := range elements {
		if _, ok := setMap[element]; !ok {
			setMap[element] = struct{}{}
		}
	}

	return slices.Collect(maps.Keys(setMap))
}
