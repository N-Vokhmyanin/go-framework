package maps

func ToSlice[M ~map[K]V, K comparable, V any, T any](in M, fn func(K, V) T) (out []T) {
	var idx int
	out = make([]T, len(in))
	for k, v := range in {
		out[idx] = fn(k, v)
		idx++
	}
	return out
}
