package utils_go

func Map[T any, U any](s []T, f func(T) U) []U {
	r := make([]U, len(s), len(s))
	for i, v := range s {
		r[i] = f(v)
	}
	return r
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
