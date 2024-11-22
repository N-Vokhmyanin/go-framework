package maps

import "golang.org/x/exp/constraints"

func Merge[KEY constraints.Ordered, VAL any](in ...map[KEY]VAL) (out map[KEY]VAL) {
	out = make(map[KEY]VAL)
	for _, m := range in {
		for key, val := range m {
			out[key] = val
		}
	}
	return out
}
