package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/N-Vokhmyanin/go-framework/cache"
	"github.com/N-Vokhmyanin/go-framework/contracts"
	"github.com/N-Vokhmyanin/go-framework/logger"
	"github.com/N-Vokhmyanin/go-framework/tracer/trace"
	"github.com/eko/gocache/v2/store"
	"github.com/go-redis/redis/v8"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"sync"
	"time"
)

const (
	reconnectDelay = 5 * time.Second

	onceJobDefaultDelay = 500 * time.Millisecond
	onceJobMinDelay     = 1 * time.Millisecond
	onceJobCachePrefix  = "once-job-"
)

type amqpConnector struct {
	sync.Mutex

	initFlag bool
	initLock sync.Mutex

	manager     Manager
	uri         string
	name        string
	log         logger.Logger
	dp          contracts.Dispatcher
	cache       cache.CacheInterface
	workers     map[string]*amqpWorker
	handlers    map[string]Handler
	connection  *amqp.Connection
	channel     *amqp.Channel
	notifyStop  chan bool
	notifyClose chan *amqp.Error
	isConnected bool
	isStopped   bool

	stoppingTimeout time.Duration
}

var _ Connection = (*amqpConnector)(nil)
var _ contracts.CanStart = (*amqpConnector)(nil)
var _ contracts.CanStop = (*amqpConnector)(nil)

//goland:noinspection GoExportedFuncWithUnexportedType
func NewAmqpConnector(
	uri string,
	m Manager,
	q Queue,
	log logger.Logger,
	dp contracts.Dispatcher,
	ch cache.CacheInterface,
	stoppingTimeout time.Duration,
) *amqpConnector {
	queue := amqpConnector{
		uri:        uri,
		manager:    m,
		dp:         dp,
		cache:      ch,
		name:       q.Name(),
		log:        log.With("queue.name", q.Name()),
		workers:    make(map[string]*amqpWorker),
		handlers:   make(map[string]Handler),
		notifyStop: make(chan bool),

		stoppingTimeout: stoppingTimeout,
	}
	var i uint
	for i = 0; i < q.Workers(); i++ {
		queue.newWorker(fmt.Sprintf("worker-%d", i+1))
	}
	return &queue
}

func (q *amqpConnector) Handler(handlers ...Handler) {
	q.Lock()
	defer q.Unlock()

	for _, handler := range handlers {
		logKV := []interface{}{
			"name",
			handler.Name(),
			"queue",
			handler.Queue(),
		}
		if _, ok := q.handlers[handler.Name()]; ok {
			q.log.Warnw("handler already registered", logKV...)
		} else {
			q.handlers[handler.Name()] = handler
		}
	}
}

func (q *amqpConnector) Push(ctx context.Context, job Job, opts ...JobOptionFunc) (err error) {
	ctx, span := trace.Start(ctx, "amqpConnector.Push")
	defer func() { trace.End(span, err) }()

	if !q.initConnection() {
		return ErrNotConnected{}
	}

	var wrapper amqpJobWrapper
	if wrappedJob, ok := job.(*amqpJobWrapper); ok {
		wrapper = *wrappedJob // job already wrapped
	} else if wrapper, err = wrap(WithOptions(job, opts...)); err != nil {
		return err
	}
	if wrapper.Options.Hash != "" {
		return q.once(ctx, wrapper, wrapper.Options.Hash)
	}
	if wrapper.Options.GetDelay() > 0 {
		return q.delay(ctx, wrapper, wrapper.Options.GetDelay())
	}
	return q.push(ctx, wrapper)
}

func (q *amqpConnector) StartService() {
	if !q.initConnection() {
		return
	}
	for _, worker := range q.workers {
		go worker.start()
	}
}

func (q *amqpConnector) StopService() {
	close(q.notifyStop)
	if !q.isConnected {
		return
	}

	var wg sync.WaitGroup
	wg.Add(len(q.workers))
	for _, worker := range q.workers {
		go func(w *amqpWorker) {
			defer wg.Done()
			w.stop()
		}(worker)
	}
	wg.Wait()

	if err := q.channel.Close(); err != nil {
		q.log.Errorw("failed close channel", zap.Error(err))
	} else if err = q.connection.Close(); err != nil {
		q.log.Errorw("failed close connection", zap.Error(err))
	}
	q.isConnected = false
}

func (q *amqpConnector) stream() (<-chan amqp.Delivery, error) {
	if !q.isConnected {
		return nil, ErrNotConnected{}
	}
	return q.channel.Consume(
		q.name,
		"",    // Consumer
		false, // Auto-Ack
		false, // Exclusive
		false, // No-local
		false, // No-Wait
		nil,   // Args
	)
}

func (q *amqpConnector) push(ctx context.Context, job amqpJobWrapper) (err error) {
	if !q.isConnected {
		return ErrNotConnected{}
	}
	body, err := json.Marshal(job)
	if err != nil {
		return err
	}

	q.Lock()
	defer q.Unlock()

	defer func() {
		if err == nil {
			q.dp.Fire(ctx, JobPushedEvent{Job: &job})
		}
	}()

	return q.channel.Publish(
		"",     // Exchange
		q.name, // Routing key
		false,  // Mandatory
		false,  // Immediate
		amqp.Publishing{
			ContentType: "application/json",
			Timestamp:   time.Now(),
			Body:        body,
		},
	)
}

func (q *amqpConnector) once(ctx context.Context, job amqpJobWrapper, hash string) (err error) {
	if q.cache == nil {
		return errors.New("to use once job, connect cache provider")
	}

	cacheKey := onceJobCachePrefix + job.JobName + "-" + hash
	cachedHash, err := q.cache.Get(ctx, cacheKey)
	if err != nil && !errors.Is(err, redis.Nil) {
		return err
	}
	if v, ok := cachedHash.(string); ok && v != "" {
		q.log.Debugw("once job already pushed", "job.name", job.JobName, "job.hash", hash)
		return nil
	}

	delay := job.Options.GetDelay()
	if delay == 0 {
		delay = onceJobDefaultDelay
	} else if delay < onceJobMinDelay {
		delay = onceJobMinDelay
	}

	err = q.cache.Set(
		ctx,
		cacheKey,
		hash,
		&store.Options{
			Expiration: delay,
		},
	)
	if err != nil {
		return err
	}

	return q.delay(ctx, job, delay)
}

func (q *amqpConnector) delay(ctx context.Context, job amqpJobWrapper, delay time.Duration) (err error) {
	if delay == 0 {
		return q.push(ctx, job)
	}
	if !q.isConnected {
		return ErrNotConnected{}
	}
	body, err := json.Marshal(job)
	if err != nil {
		return err
	}

	q.Lock()
	defer q.Unlock()

	queueName := fmt.Sprintf("deferred %s for %s", q.name, delay.String())

	channel := q.channel
	args := amqp.Table{
		"x-expires":                 int64((delay + 3*time.Second) / time.Millisecond),
		"x-dead-letter-routing-key": q.name,
		"x-dead-letter-exchange":    "",
	}
	queue, err := channel.QueueDeclare(
		queueName, // name
		true,      // durable
		true,      // delete when unused
		false,     // exclusive
		false,     // no-wait
		args,      // arguments
	)
	if err != nil {
		return err
	}

	defer func() {
		if err == nil {
			q.dp.Fire(
				ctx,
				JobDelayedEvent{
					Job:   &job,
					Delay: delay,
				},
			)
		}
	}()

	return channel.Publish(
		"",         // exchange
		queue.Name, // routing key
		false,      // mandatory
		false,
		amqp.Publishing{
			Expiration:   fmt.Sprintf("%d", uint(delay/time.Millisecond)),
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			ContentType:  "application/json",
			Body:         body,
		},
	)
}

func (q *amqpConnector) newWorker(name string) *amqpWorker {
	q.Lock()
	defer q.Unlock()
	worker := newAmqpWorker(name, q, q.log, q.dp, q.cache)
	q.workers[name] = worker
	return worker
}

func (q *amqpConnector) connect(addr string) error {
	conn, err := amqp.Dial(addr)
	if err != nil {
		return err
	}

	ch, err := conn.Channel()
	if err != nil {
		return err
	}

	err = ch.Confirm(false)
	if err != nil {
		return err
	}

	_, err = ch.QueueDeclare(
		q.name, // Name
		true,   // Durable
		false,  // Delete when unused
		false,  // Exclusive
		false,  // No-wait
		nil,    // Arguments
	)
	if err != nil {
		return err
	}

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return err
	}

	q.channel = ch
	q.connection = conn
	q.notifyClose = make(chan *amqp.Error)
	q.channel.NotifyClose(q.notifyClose)

	q.isConnected = true
	q.log.Infow("successful connected")

	return nil
}

func (q *amqpConnector) reconnect() {
	for {
		if q.isStopped {
			return
		}
		q.isConnected = false
		q.log.Infow("attempting to connect")

	reconnectLoop:
		for {
			select {
			case <-q.notifyStop:
				q.isStopped = true
				return
			default:
				err := q.connect(q.uri)
				if err == nil {
					break reconnectLoop
				}
				q.log.Warnw("failed to connect. retrying...", zap.Error(err))
				time.Sleep(reconnectDelay)
			}
		}

		select {
		case <-q.notifyStop:
			q.isStopped = true
			return
		case e := <-q.notifyClose:
			if e != nil {
				q.log.Warnw("notify close", zap.Error(e))
			}
		}
	}
}

func (q *amqpConnector) waitConnection() bool {
	q.Lock()
	defer q.Unlock()
	if q.isConnected {
		return q.isConnected
	}
	connected := make(chan bool)
	go func() {
		for range time.NewTicker(100 * time.Millisecond).C {
			if q.isConnected {
				connected <- true
				return
			}
			if q.isStopped {
				connected <- false
				return
			}
		}
	}()
	return <-connected
}

func (q *amqpConnector) initConnection() bool {
	q.initLock.Lock()
	defer q.initLock.Unlock()

	if q.initFlag {
		return q.isConnected
	}
	q.initFlag = true

	go q.reconnect()
	return q.waitConnection()
}

func (q *amqpConnector) inspect() (amqp.Queue, error) {
	if !q.initConnection() {
		return amqp.Queue{}, ErrNotConnected{}
	}
	return q.channel.QueueInspect(q.name)
}
