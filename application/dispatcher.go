package application

import (
	"context"
	"github.com/N-Vokhmyanin/go-framework/contracts"
	"reflect"
	"sync"
)

var ctxType = reflect.TypeOf((*context.Context)(nil)).Elem()

type dispatcher struct {
	sync.Mutex
	listeners map[string][]interface{}
}

func NewDispatcher() contracts.Dispatcher {
	return &dispatcher{
		listeners: make(map[string][]interface{}),
	}
}

func (d *dispatcher) Listen(l interface{}) {
	d.Lock()
	defer d.Unlock()

	reflectType := reflect.TypeOf(l)
	if reflectType.Kind() != reflect.Func {
		panic("the listener must be a function")
	}
	if reflectType.NumIn() != 2 {
		panic("the listener must have 2 arguments (context, EventStruct)")
	}

	contextType := reflectType.In(0)
	if contextType.Kind() != reflect.Interface || !contextType.Implements(ctxType) {
		panic("listener first argument must be a context")
	}
	argumentType := reflectType.In(1)
	if argumentType.Kind() != reflect.Struct {
		panic("listener second argument must be a struct")
	}
	if _, ok := d.listeners[argumentType.String()]; !ok {
		d.listeners[argumentType.String()] = make([]interface{}, 0)
	}
	d.listeners[argumentType.String()] = append(d.listeners[argumentType.String()], l)
}

func (d *dispatcher) Fire(ctx context.Context, e interface{}) {
	reflectType := reflect.TypeOf(e)
	if reflectType.Kind() != reflect.Struct {
		panic("dispatched event must be a struct")
	}
	listeners, ok := d.listeners[reflectType.String()]
	if !ok || len(listeners) == 0 {
		return
	}
	callValues := []reflect.Value{
		reflect.ValueOf(ctx),
		reflect.ValueOf(e),
	}
	for _, listener := range listeners {
		reflect.ValueOf(listener).Call(callValues)
	}
}
