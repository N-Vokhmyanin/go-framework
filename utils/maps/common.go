package maps

func Keys[K comparable, E any, M ~map[K]E](in M) (out []K) {
	if len(in) == 0 {
		return nil
	}
	out = make([]K, len(in))
	var i int
	for key := range in {
		out[i] = key
		i += 1
	}
	return out
}

func Values[K comparable, E any, M ~map[K]E](in M) (out []E) {
	if len(in) == 0 {
		return nil
	}
	out = make([]E, len(in))
	var i int
	for _, val := range in {
		out[i] = val
		i += 1
	}
	return out
}
