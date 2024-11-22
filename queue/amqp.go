package queue

import (
	"context"
	"github.com/N-Vokhmyanin/go-framework/cache"
	"github.com/N-Vokhmyanin/go-framework/contracts"
	"github.com/N-Vokhmyanin/go-framework/errors"
	health "github.com/N-Vokhmyanin/go-framework/health/contracts"
	"github.com/N-Vokhmyanin/go-framework/logger"
	"github.com/N-Vokhmyanin/go-framework/tracer/trace"
	"go.opentelemetry.io/otel/attribute"
	grpcHealthV1 "google.golang.org/grpc/health/grpc_health_v1"
	"sync"
	"time"
)

type amqpManager struct {
	sync.Mutex
	uri         string
	log         logger.Logger
	dp          contracts.Dispatcher
	cache       cache.CacheInterface
	connectors  map[string]*amqpConnector
	middlewares []Middleware

	stoppingTimeout time.Duration
}

var _ Manager = (*amqpManager)(nil)
var _ health.Service = (*amqpManager)(nil)
var _ contracts.CanStart = (*amqpManager)(nil)
var _ contracts.CanStop = (*amqpManager)(nil)

func NewAmqpManager(
	uri string,
	log logger.Logger,
	dp contracts.Dispatcher,
	ch cache.CacheInterface,
	stoppingTimeout time.Duration,
) Manager {
	return &amqpManager{
		uri:        uri,
		dp:         dp,
		cache:      ch,
		log:        log.With(logger.WithComponent, "queue.amqp"),
		connectors: make(map[string]*amqpConnector),

		stoppingTimeout: stoppingTimeout,
	}
}

func (s *amqpManager) Queue(queues ...Queue) {
	s.Lock()
	defer s.Unlock()

	for _, queue := range queues {
		if _, ok := s.connectors[queue.Name()]; ok {
			s.log.Warnw("queue already registered", "name", queue.Name())
		} else {
			s.connectors[queue.Name()] = NewAmqpConnector(s.uri, s, queue, s.log, s.dp, s.cache, s.stoppingTimeout)
		}
	}
}

func (s *amqpManager) MessagesCount(queueName string) (uint, error) {
	s.Lock()
	defer s.Unlock()

	queueConnector, ok := s.connectors[queueName]
	if !ok {
		return 0, errors.Errorf("queue '%s' not registered", queueName)
	}

	info, err := queueConnector.inspect()
	if err != nil {
		return 0, err
	}
	if info.Messages > 0 {
		return uint(info.Messages), nil
	}

	return 0, nil
}

func (s *amqpManager) Handler(handlers ...Handler) {
	s.Lock()
	defer s.Unlock()

	for _, handler := range handlers {
		logKV := []interface{}{
			"name",
			handler.Name(),
			"queue",
			handler.Queue(),
		}
		if queue, ok := s.connectors[handler.Queue()]; !ok {
			s.log.Warnw("unknown queue for handler", logKV...)
		} else {
			queue.Handler(handler)
		}
	}
}

func (s *amqpManager) Middleware(middlewares ...Middleware) {
	s.Lock()
	defer s.Unlock()

	for _, middleware := range middlewares {
		if middleware != nil {
			s.middlewares = append(s.middlewares, middleware)
		}
	}
}

func (s *amqpManager) GetMiddlewares() []Middleware {
	s.Lock()
	defer s.Unlock()

	return s.middlewares
}

func (s *amqpManager) Push(ctx context.Context, job Job, opts ...JobOptionFunc) (err error) {
	ctx, span := trace.Start(ctx, "amqpManager.Push")
	span.SetAttributes(
		attribute.String("queue", job.Queue()),
		attribute.String("name", job.Name()),
	)
	defer func() { trace.End(span, err) }()

	if queue, ok := s.connectors[job.Queue()]; !ok {
		return ErrUnknownQueue{Name: job.Queue()}
	} else {
		return queue.Push(ctx, job, opts...)
	}
}

func (s *amqpManager) HealthStatus(context.Context) grpcHealthV1.HealthCheckResponse_ServingStatus {
	for _, connector := range s.connectors {
		if !connector.isConnected {
			return grpcHealthV1.HealthCheckResponse_NOT_SERVING
		}
	}
	return grpcHealthV1.HealthCheckResponse_SERVING
}

func (s *amqpManager) StartService() {
	for _, q := range s.connectors {
		go q.StartService()
	}
}

func (s *amqpManager) StopService() {
	for _, queue := range s.connectors {
		queue.StopService()
	}
}
