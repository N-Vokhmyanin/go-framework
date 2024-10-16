package di

import (
	"reflect"

	"github.com/N-Vokhmyanin/go-framework/contracts"
)

func Make[T any](c contracts.Container, resolver any) T {
	if reflect.TypeOf(resolver).Kind() != reflect.Func {
		panic("the resolver must be a function")
	}
	if reflect.TypeOf(resolver).NumOut() != 1 {
		panic("resolver must return one parameter")
	}
	return c.Make(resolver)[0].(T)
}

func Get[T any](c contracts.Container) T {
	return Make[T](c, func(v T) T {
		return v
	})
}

func Alias[IN any, OUT any](a contracts.Container) {
	a.Singleton(func(in IN) OUT {
		return any(in).(OUT)
	})
}

func AsSingleton[IN any](a contracts.Container, in IN) {
	a.Singleton(func() IN {
		return in
	})
}
