package queue

import (
	"context"
	"github.com/N-Vokhmyanin/go-framework/logger"
)

type handlerMiddlewareWrapper struct {
	Handler
	middleware Middleware
}

func newHandlerWithMiddlewares(handler Handler, middlewares []Middleware) Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = &handlerMiddlewareWrapper{
			Handler:    handler,
			middleware: middlewares[i],
		}
	}
	return handler
}

func (w *handlerMiddlewareWrapper) Handle(ctx context.Context, log logger.Logger, i JobInteract) error {
	return w.middleware(ctx, log, i, w.Handler)
}
