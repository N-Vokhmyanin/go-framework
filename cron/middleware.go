package cron

import (
	"context"
	"github.com/N-Vokhmyanin/go-framework/logger"
)

type taskMiddlewareWrapper struct {
	Task
	middleware Middleware
}

func newTaskWithMiddlewares(task Task, middlewares ...Middleware) Task {
	for i := len(middlewares) - 1; i >= 0; i-- {
		task = &taskMiddlewareWrapper{
			Task:       task,
			middleware: middlewares[i],
		}
	}
	return task
}

func (t *taskMiddlewareWrapper) Handle(ctx context.Context, log logger.Logger) error {
	return t.middleware(ctx, log, t.Task)
}
