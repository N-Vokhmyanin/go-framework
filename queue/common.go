package queue

import (
	"context"
	"fmt"
	"github.com/N-Vokhmyanin/go-framework/contracts"
	"github.com/N-Vokhmyanin/go-framework/logger"
	"strings"
)

type jobHandlerFunc = func(ctx context.Context, log logger.Logger, i JobInteract) error

type simpleQueue struct {
	name    string
	workers uint
}

var _ Queue = (*simpleQueue)(nil)

//goland:noinspection GoUnusedExportedFunction
func SimpleQueue(name string, workers uint) Queue {
	return &simpleQueue{
		name:    name,
		workers: workers,
	}
}

func (q *simpleQueue) Name() string {
	return q.name
}

func (q *simpleQueue) Workers() uint {
	return q.workers
}

type simpleHandler struct {
	name  string
	queue string
	fn    jobHandlerFunc
}

var _ Handler = (*simpleHandler)(nil)

//goland:noinspection GoUnusedExportedFunction
func SimpleHandler(name, queue string, fn jobHandlerFunc) Handler {
	return &simpleHandler{
		name:  name,
		queue: queue,
		fn:    fn,
	}
}

func (h *simpleHandler) Name() string {
	return h.name
}

func (h *simpleHandler) Queue() string {
	return h.queue
}

func (h *simpleHandler) Handle(ctx context.Context, log logger.Logger, i JobInteract) error {
	return h.fn(ctx, log, i)
}

type SimpleQueues interface {
	Add(name string, workers uint)
	Config(c contracts.ConfigSet, prefix string)
	Register(queueManager Manager)
}

type simpleQueues []*simpleQueue

func NewSimpleQueues(names ...string) SimpleQueues {
	queues := &simpleQueues{}
	for _, name := range names {
		queues.Add(name, 1)
	}
	return &simpleQueues{}
}

func (queues *simpleQueues) Add(name string, workers uint) {
	*queues = append(
		*queues, &simpleQueue{
			name:    name,
			workers: workers,
		},
	)
}

func (queues simpleQueues) Config(c contracts.ConfigSet, prefix string) {
	for _, queue := range queues {
		name := queue.name
		if prefix != "" {
			if !strings.HasPrefix(name, prefix) {
				continue
			}
			name = strings.Replace(name, prefix, "", 1)
		}
		c.UintVar(
			&queue.workers,
			fmt.Sprintf("QUEUE_%s_WORKERS", strings.ToUpper(strings.ReplaceAll(name, "-", "_"))),
			queue.workers,
			fmt.Sprintf("workers for queue %s", name),
		)
	}
}

func (queues simpleQueues) Register(queueManager Manager) {
	for _, queue := range queues {
		queueManager.Queue(queue)
	}
}
