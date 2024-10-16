package rfl

import "reflect"

func FullTypeName(v any) string {
	rv := reflect.TypeOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	return rv.PkgPath() + "." + rv.Name()
}
