package utils_go

func Map[T any, U any](s []T, f func(T) U) []U {
	r := make([]U, len(s), len(s))
	for i, v := range s {
		r[i] = f(v)
	}
	return r
}
