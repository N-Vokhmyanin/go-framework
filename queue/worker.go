package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/N-Vokhmyanin/go-framework/cache"
	"github.com/N-Vokhmyanin/go-framework/contracts"
	"github.com/N-Vokhmyanin/go-framework/errors"
	"github.com/N-Vokhmyanin/go-framework/logger"
	"github.com/N-Vokhmyanin/go-framework/logger/grpc/ctxlog"
	"github.com/N-Vokhmyanin/go-framework/tracer/trace"
	"github.com/N-Vokhmyanin/go-framework/utils/maps"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"sync"
	"time"
)

const DefaultReleaseDelay = time.Second * 10

type amqpWorker struct {
	sync.Mutex
	conn  *amqpConnector
	log   logger.Logger
	dp    contracts.Dispatcher
	work  bool
	cache cache.CacheInterface

	cancelFunc context.CancelFunc
}

func newAmqpWorker(
	name string,
	c *amqpConnector,
	log logger.Logger,
	dp contracts.Dispatcher,
	ch cache.CacheInterface,
) *amqpWorker {
	return &amqpWorker{
		conn:  c,
		log:   log.With("queue.worker", name),
		dp:    dp,
		cache: ch,
	}
}

func (w *amqpWorker) start() {
	w.log.Infow("worker started")
	var err error
	var isClosed = true
	var delivery <-chan amqp.Delivery
	for {
		if isClosed {
			delivery, err = w.conn.stream()
			if err != nil {
				w.log.Warnw("queue unavailable, waiting reconnect")
				time.Sleep(10 * time.Second)
				continue
			}
			isClosed = false
		}
		msg, ok := <-delivery
		if w.conn.isStopped {
			return
		}
		if !ok {
			isClosed = true
			continue
		}
		_ = w.process(msg)
	}
}

func (w *amqpWorker) stop() {
	wait := make(chan bool)
	go func() {
		w.wait()
		wait <- true
	}()

	select {
	case <-wait:
		return
	case <-time.After(w.conn.stoppingTimeout):
		if w.cancelFunc != nil {
			w.cancelFunc()
			w.cancelFunc = nil
		}
		<-wait
	}
}

func (w *amqpWorker) process(msg amqp.Delivery) (err error) {
	w.work = true
	defer func() {
		if err = msg.Ack(false); err != nil {
			w.log.Errorw("ack message failed", zap.Error(err))
		}
		w.work = false
	}()

	var jobWrapper amqpJobWrapper
	if err = json.Unmarshal(msg.Body, &jobWrapper); err != nil {
		w.log.Errorw("incorrect message body", "msg.body", string(msg.Body), zap.Error(err))
		return err
	}

	log := w.log.With("job.name", jobWrapper.JobName)

	if jobWrapper.Options.Hash != "" {
		log = log.With("job.hash", jobWrapper.Options.Hash)
	}

	var handler Handler
	if handler = w.conn.handlers[jobWrapper.JobName]; handler == nil {
		w.log.Errorw("handler not registered", "job.name", jobWrapper.JobName)
		return ErrUnknownJob{Name: jobWrapper.JobName}
	}

	jobWrapper.Attempts++
	log = log.With("job.attempts", jobWrapper.Attempts)

	log.Debugw("job started")
	defer func() { log.Debugw("job finished") }()

	var ctx context.Context
	var span trace.Span
	ctx, w.cancelFunc = context.WithCancel(context.Background())

	w.log.Infof("job.%s", jobWrapper.JobName)

	span.SetAttributes(
		attribute.String("job.name", jobWrapper.JobName),
		attribute.String("job.body", jobWrapper.JobBody),
		attribute.Int("job.attempts", int(jobWrapper.Attempts)),
	)
	defer func() {
		trace.End(span, err)
		if w.cancelFunc != nil {
			w.cancelFunc()
		}
	}()

	ctx = ctxlog.ToContext(ctx, log)

	w.dp.Fire(ctx, JobStartedEvent{Job: &jobWrapper})
	defer func() { w.dp.Fire(ctx, JobFinishedEvent{Job: &jobWrapper}) }()

	var jobInteracts = newAmqpJobInteract(&jobWrapper)
	err = w.handleRecover(
		ctx,
		log,
		jobInteracts,
		newHandlerWithMiddlewares(handler, w.conn.manager.GetMiddlewares()),
	)
	if err == nil && jobInteracts.failedErr != nil {
		err = jobInteracts.failedErr
	}

	log = ctxlog.ExtractWithFallback(ctx, log)

	jobWrapper.Result = maps.Merge(jobWrapper.Result, jobInteracts.resultData)

	isCanceled := errors.Is(err, context.Canceled)
	isRequeue := isCanceled || jobInteracts.releaseFlag || (err != nil && !jobInteracts.deleteFlag)
	if err != nil && !isCanceled {
		maxAttemptsReached := false
		if maxAttempts := jobWrapper.Options.MaxAttempts; maxAttempts > 0 {
			// check max attempts if job has option
			if maxAttemptsReached = jobWrapper.Attempts >= maxAttempts; maxAttemptsReached {
				// job reached max attempts
				isRequeue = false
			}
		}
		log.Errorw(
			"handle job error",
			"max_attempts_reached", maxAttemptsReached,
			"requeue", isRequeue,
			zap.Error(err),
		)
		if !isRequeue {
			w.dp.Fire(
				ctx,
				JobFailedEvent{
					Job: &jobWrapper,
					Err: err,
				},
			)
		}
	}

	if isRequeue {
		// needs push job back to queue
		delay := DefaultReleaseDelay
		if jobInteracts.releaseDelay != nil {
			delay = *jobInteracts.releaseDelay
		}
		if err = w.conn.delay(ctx, jobWrapper, delay); err != nil {
			log.Errorw("requeue job failed", zap.Error(err))
			return err
		}
		return nil
	}

	if err != nil {
		// check job has fails jobs
		for _, failJob := range jobWrapper.Options.Fails {
			failJob.Result = jobWrapper.Result
			if err = w.conn.manager.Push(ctx, &failJob); err != nil {
				log.Errorw("fail job push failed", zap.Error(err))
				return err
			}
		}
	} else {
		// check job has after jobs
		for _, afterJob := range jobWrapper.Options.After {
			afterJob.Result = jobWrapper.Result
			if err = w.conn.manager.Push(ctx, &afterJob); err != nil {
				log.Errorw("after job push failed", zap.Error(err))
				return err
			}
		}
	}

	// push always jobs
	for _, alwaysJob := range jobWrapper.Options.Always {
		alwaysJob.Result = jobWrapper.Result
		if pushErr := w.conn.manager.Push(ctx, &alwaysJob); pushErr != nil {
			log.Errorw("always job push failed", zap.Error(pushErr))
		}
	}

	return err
}

func (w *amqpWorker) handleRecover(ctx context.Context, log logger.Logger, i JobInteract, h Handler) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if recoverErr, ok := r.(error); ok {
				err = errors.WrapWith(recoverErr, "panic recovered")
			} else {
				err = fmt.Errorf("panic recovered: %+v", r)
			}
		}
	}()
	return h.Handle(ctx, log, i)
}

func (w *amqpWorker) wait() {
	for {
		if !w.work {
			return
		}
		w.log.Infow("waiting job done")
		time.Sleep(1 * time.Second)
	}
}
