package queue

import (
	"context"
	"github.com/N-Vokhmyanin/go-framework/logger"
)

type Queue interface {
	Name() string
	Workers() uint
}

type Job interface {
	Name() string
	Queue() string
	Body() ([]byte, error)
}

type OnceJob interface {
	Job
	Hash() string
}

type JobWithOptions interface {
	Options() []JobOptionFunc
}

type JobInteract interface {
	Unmarshal(j interface{}) error
	WithResult(r map[string]string) error
	GetResult(key string) string
	Release(delay uint) error
	Fail(err error) error
	Delete() error
	Attempts() uint
	Body() []byte
	IsFailed() bool
}

type Handler interface {
	Name() string
	Queue() string
	Handle(ctx context.Context, log logger.Logger, i JobInteract) error
}

type Middleware func(ctx context.Context, log logger.Logger, i JobInteract, handler Handler) error

type Manager interface {
	Connection
	Queue(queues ...Queue)
	MessagesCount(queueName string) (uint, error)
	Middleware(middlewares ...Middleware)
	GetMiddlewares() []Middleware
}

type Connection interface {
	Handler(handlers ...Handler)
	Push(ctx context.Context, job Job, opts ...JobOptionFunc) error
}
