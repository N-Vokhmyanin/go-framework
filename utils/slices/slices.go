package slices

import (
	"fmt"
	"strings"

	"golang.org/x/exp/constraints"
	"golang.org/x/exp/slices"
)

// Equal reports whether two slices are equal: the same length and all
// elements equal. If the lengths are different, Equal returns false.
// Otherwise, the elements are compared in increasing index order, and the
// comparison stops at the first unequal pair.
// Floating point NaNs are not considered equal.
func Equal[E comparable](s1, s2 []E) bool {
	return slices.Equal(s1, s2)
}

// EqualFunc reports whether two slices are equal using a comparison
// function on each pair of elements. If the lengths are different,
// EqualFunc returns false. Otherwise, the elements are compared in
// increasing index order, and the comparison stops at the first index
// for which eq returns false.
func EqualFunc[E1, E2 any](s1 []E1, s2 []E2, eq func(E1, E2) bool) bool {
	return slices.EqualFunc(s1, s2, eq)
}

// Compare compares the elements of s1 and s2.
// The elements are compared sequentially, starting at index 0,
// until one element is not equal to the other.
// The result of comparing the first non-matching elements is returned.
// If both slices are equal until one of them ends, the shorter slice is
// considered less than the longer one.
// The result is 0 if s1 == s2, -1 if s1 < s2, and +1 if s1 > s2.
// Comparisons involving floating point NaNs are ignored.
func Compare[E constraints.Ordered](s1, s2 []E) int {
	return slices.Compare(s1, s2)
}

// CompareFunc is like Compare but uses a comparison function
// on each pair of elements. The elements are compared in increasing
// index order, and the comparisons stop after the first time cmp
// returns non-zero.
// The result is the first non-zero result of cmp; if cmp always
// returns 0 the result is 0 if len(s1) == len(s2), -1 if len(s1) < len(s2),
// and +1 if len(s1) > len(s2).
func CompareFunc[E1, E2 any](s1 []E1, s2 []E2, cmp func(E1, E2) int) int {
	return slices.CompareFunc(s1, s2, cmp)
}

// Index returns the index of the first occurrence of v in s,
// or -1 if not present.
func Index[E comparable](s []E, v E) int {
	return slices.Index(s, v)
}

// IndexFunc returns the first index i satisfying f(s[i]),
// or -1 if none do.
func IndexFunc[E any](s []E, f func(E) bool) int {
	return slices.IndexFunc(s, f)
}

// Contains reports whether v is present in s.
func Contains[E comparable](s []E, v E) bool {
	return Index(s, v) >= 0
}

// Insert inserts the values v... into s at index i,
// returning the modified slice.
// In the returned slice r, r[i] == v[0].
// Insert panics if i is out of range.
// This function is O(len(s) + len(v)).
func Insert[S ~[]E, E any](s S, i int, v ...E) S {
	return slices.Insert(s, i, v...)
}

// Delete removes the elements s[i:j] from s, returning the modified slice.
// Delete panics if s[i:j] is not a valid slice of s.
// Delete modifies the contents of the slice s; it does not create a new slice.
// Delete is O(len(s)-j), so if many items must be deleted, it is better to
// make a single call deleting them all together than to delete one at a time.
// Delete might not modify the elements s[len(s)-(j-i):len(s)]. If those
// elements contain pointers you might consider zeroing those elements so that
// objects they reference can be garbage collected.
func Delete[S ~[]E, E any](s S, i, j int) S {
	return slices.Delete(s, i, j)
}

// Clone returns a copy of the slice.
// The elements are copied using assignment, so this is a shallow clone.
func Clone[S ~[]E, E any](s S) S {
	return slices.Clone(s)
}

// Compact replaces consecutive runs of equal elements with a single copy.
// This is like the uniq command found on Unix.
// Compact modifies the contents of the slice s; it does not create a new slice.
func Compact[S ~[]E, E comparable](s S) S {
	return slices.Compact(s)
}

// CompactFunc is like Compact but uses a comparison function.
func CompactFunc[S ~[]E, E any](s S, eq func(E, E) bool) S {
	return slices.CompactFunc(s, eq)
}

// Grow increases the slice's capacity, if necessary, to guarantee space for
// another n elements. After Grow(n), at least n elements can be appended
// to the slice without another allocation. Grow may modify elements of the
// slice between the length and the capacity. If n is negative or too large to
// allocate the memory, Grow panics.
func Grow[S ~[]E, E any](s S, n int) S {
	return slices.Grow(s, n)
}

// Clip removes unused capacity from the slice, returning s[:len(s):len(s)].
func Clip[S ~[]E, E any](s S) S {
	return slices.Clip(s)
}

func Keys[KEY constraints.Ordered, VAL any](in map[KEY]VAL) (out []KEY) {
	out = make([]KEY, len(in))
	i := 0
	for key := range in {
		out[i] = key
		i++
	}
	return out
}

func Values[KEY constraints.Ordered, VAL any](in map[KEY]VAL) (out []VAL) {
	out = make([]VAL, len(in))
	i := 0
	for _, val := range in {
		out[i] = val
		i++
	}
	return out
}

func Map[T1, T2 any](in []T1, fn func(item T1) T2) (out []T2) {
	out = make([]T2, len(in))
	for i, item := range in {
		out[i] = fn(item)
	}
	return out
}

func ToMap[IN any, KEY constraints.Ordered, VAL any](in []IN, fn func(item IN) (KEY, VAL)) (out map[KEY]VAL) {
	out = map[KEY]VAL{}
	for _, item := range in {
		k, v := fn(item)
		out[k] = v
	}
	return out
}

func Filter[T any](in []T, fn func(item T) bool) (out []T) {
	for _, item := range in {
		if fn(item) {
			out = append(out, item)
		}
	}
	return
}

func EqualUnSort[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for _, aItem := range a {
		if !slices.Contains(b, aItem) {
			return false
		}
	}
	return true
}

func ContainsFunc[T comparable](source []T, fn func(item T) bool) bool {
	return slices.IndexFunc(source, fn) >= 0
}

func ContainsAny[T comparable](source []T, items ...T) bool {
	for _, item := range items {
		if slices.Contains(source, item) {
			return true
		}
	}
	return false
}

func Diff[T comparable](a, b []T) (diff []T) {
	for _, aItem := range a {
		if !slices.Contains(b, aItem) {
			diff = append(diff, aItem)
		}
	}
	return diff
}

func UnionAll[T comparable](items ...[]T) (out []T) {
	switch len(items) {
	case 0:
		return out
	case 1:
		return items[0]
	default:
		out = items[0]
		for _, item := range items[1:] {
			for _, el := range item {
				if !slices.Contains(out, el) {
					out = append(out, el)
				}
			}
		}
	}
	return
}

func Join[T constraints.Ordered](in []T, sep string) string {
	strSlice := Map(
		in, func(item T) string {
			return fmt.Sprint(item)
		},
	)
	return strings.Join(strSlice, sep)
}

func OfAny[T any](in []T) []any {
	return Map(
		in, func(item T) any {
			return item
		},
	)
}

func Unique[S ~[]E, E comparable](in S) S {
	if len(in) < 2 {
		return in
	}
	mapVal := make(map[E]bool, len(in))
	out := make(S, len(mapVal))
	for _, elem := range in {
		if !mapVal[elem] {
			out = append(out, elem)
			mapVal[elem] = true
		}
	}
	return out
}
