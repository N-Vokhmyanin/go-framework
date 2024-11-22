package cron

import (
	"context"
	"github.com/N-Vokhmyanin/go-framework/logger"
	"time"
)

type Schedule interface {
	Next(t time.Time) time.Time
}

type Task interface {
	Name() string
	Schedule() Schedule
	Handle(ctx context.Context, log logger.Logger) error
}

type Middleware func(ctx context.Context, log logger.Logger, task Task) error

type Service interface {
	Add(tasks ...Task)
	Middleware(middlewares ...Middleware)
	Tasks() map[string]Task
	Call(name string) error
}
