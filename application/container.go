package application

import (
	"github.com/N-Vokhmyanin/go-framework/contracts"
	"reflect"
)

type container struct {
	binds map[reflect.Type]*binding
}

type binding struct {
	singleton bool
	resolver  interface{} // resolver function
	instance  interface{} // instance stored for singleton bindings
}

var _ contracts.Container = (*container)(nil)

func newContainer() contracts.Container {
	return &container{
		binds: map[reflect.Type]*binding{},
	}
}

func (c *container) invoke(function interface{}) interface{} {
	return reflect.ValueOf(function).Call(c.arguments(function))[0].Interface()
}

func (c *container) bind(resolver interface{}, singleton bool) {
	if reflect.TypeOf(resolver).Kind() != reflect.Func {
		panic("the resolver must be a function")
	}

	for i := 0; i < reflect.TypeOf(resolver).NumOut(); i++ {
		var instance interface{}
		c.binds[reflect.TypeOf(resolver).Out(i)] = &binding{
			singleton: singleton,
			resolver:  resolver,
			instance:  instance,
		}
	}
}

func (c *container) arguments(function interface{}) []reflect.Value {
	argumentsCount := reflect.TypeOf(function).NumIn()
	arguments := make([]reflect.Value, argumentsCount)

	for i := 0; i < argumentsCount; i++ {
		abstraction := reflect.TypeOf(function).In(i)

		if concrete, ok := c.binds[abstraction]; ok {
			arguments[i] = reflect.ValueOf(c.resolve(concrete))
		} else {
			arguments[i] = reflect.New(abstraction).Elem()
		}
	}

	return arguments
}

func (c *container) resolve(concrete *binding) interface{} {
	if concrete.instance != nil {
		return concrete.instance
	}
	instance := c.invoke(concrete.resolver)
	if concrete.singleton {
		concrete.instance = instance
	}
	return instance
}

func (c *container) Instances() []interface{} {
	var instances []interface{}
	for _, bind := range c.binds {
		if bind.singleton {
			instance := c.resolve(bind)
			if reflect.TypeOf(instance).Kind() != reflect.Func {
				instances = append(instances, instance)
			}
		}
	}
	return instances
}

func (c *container) Singleton(resolver interface{}) {
	c.bind(resolver, true)
}

func (c *container) Transient(resolver interface{}) {
	c.bind(resolver, false)
}

func (c *container) Make(receiver interface{}) (res []interface{}) {
	if reflect.TypeOf(receiver) == nil {
		panic("cannot detect type of the receiver, make sure your are passing reference of the object")
	}

	if reflect.TypeOf(receiver).Kind() == reflect.Ptr {
		abstraction := reflect.TypeOf(receiver).Elem()
		if concrete, ok := c.binds[abstraction]; ok {
			reflect.ValueOf(receiver).Elem().Set(reflect.ValueOf(c.resolve(concrete)))
			return nil
		}
		panic("no concrete found for the abstraction " + abstraction.String())
	}

	if reflect.TypeOf(receiver).Kind() == reflect.Func {
		arguments := c.arguments(receiver)
		values := reflect.ValueOf(receiver).Call(arguments)
		for _, value := range values {
			if value.CanInterface() {
				res = append(res, value.Interface())
			}
		}
		return res
	}

	panic("the receiver must be either a reference or a callback")
}
