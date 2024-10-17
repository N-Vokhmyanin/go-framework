package ctxapp

import (
	"context"
	"github.com/N-Vokhmyanin/go-framework/utils/di"
)

func Get[T any](ctx context.Context) (out T) {
	app := Extract(ctx)
	if app == nil {
		return
	}
	return di.Get[T](app)
}

func Make[T any](ctx context.Context, resolver any) (out T) {
	app := Extract(ctx)
	if app == nil {
		return
	}
	return di.Make[T](app, resolver)
}
