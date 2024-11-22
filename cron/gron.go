package cron

import (
	"context"
	"errors"
	"fmt"
	"github.com/N-Vokhmyanin/go-framework/contracts"
	"github.com/N-Vokhmyanin/go-framework/logger"
	"github.com/N-Vokhmyanin/go-framework/logger/grpc/ctxlog"
	"github.com/roylee0704/gron"
	"go.uber.org/zap"
	"sync"
)

//goland:noinspection SpellCheckingInspection
type gronService struct {
	sync.Mutex

	enabled     bool
	log         logger.Logger
	cron        *gron.Cron
	tasks       map[string]Task
	middlewares []Middleware
	cancelTasks map[string]context.CancelFunc
	wg          sync.WaitGroup
}

var _ Service = (*gronService)(nil)
var _ contracts.CanBoot = (*gronService)(nil)
var _ contracts.CanStart = (*gronService)(nil)
var _ contracts.CanStop = (*gronService)(nil)

//goland:noinspection SpellCheckingInspection
func NewGronService(enabled bool, log logger.Logger) Service {
	return &gronService{
		enabled:     enabled,
		log:         log.With(logger.WithComponent, "cron.gron"),
		cron:        gron.New(),
		tasks:       make(map[string]Task),
		cancelTasks: map[string]context.CancelFunc{},
	}
}

func (s *gronService) setCancelTask(taskId string, cancel context.CancelFunc) {
	s.Lock()
	defer s.Unlock()
	s.cancelTasks[taskId] = cancel
	s.wg.Add(1)
}

func (s *gronService) unsetCancelTask(taskId string) {
	s.Lock()
	defer s.Unlock()
	delete(s.cancelTasks, taskId)
	s.wg.Done()
}

func (s *gronService) callCancelTasks() {
	defer s.wg.Wait()
	s.Lock()
	defer s.Unlock()
	for _, cancel := range s.cancelTasks {
		cancel()
	}
	s.cancelTasks = make(map[string]context.CancelFunc)
}

func (s *gronService) createTaskFunc(task Task) func() {
	num := 0
	return func() {
		num += 1
		var err error
		taskId := fmt.Sprintf("%s-%d", task.Name(), num)

		ctx, cancel := context.WithCancel(context.Background())
		s.setCancelTask(taskId, cancel)
		defer func() {
			s.unsetCancelTask(taskId)
			if !errors.Is(err, context.Canceled) {
				cancel()
			}
		}()

		s.log.Infof("cron.%s", task.Name())

		ctx = ctxlog.ToContext(ctx, s.log)
		ctxlog.AddFields(ctx, "cron.name", task.Name())
		log := ctxlog.ExtractWithFallback(ctx, s.log)

		taskWithMiddlewares := newTaskWithMiddlewares(task, s.middlewares...)

		log.Debugw("task started")
		err = taskWithMiddlewares.Handle(ctx, log)
		log = ctxlog.ExtractWithFallback(ctx, log)
		if err != nil {
			log.Errorw("task failed", zap.Error(err))
		}
		log.Debugw("task completed")
	}
}

func (s *gronService) BootService() {
	for _, task := range s.tasks {
		s.cron.AddFunc(task.Schedule(), s.createTaskFunc(task))
	}
}

func (s *gronService) StartService() {
	if s.enabled {
		s.cron.Start()
	}
}

func (s *gronService) StopService() {
	s.cron.Stop()
	s.callCancelTasks()
}

func (s *gronService) Add(tasks ...Task) {
	for _, task := range tasks {
		s.tasks[task.Name()] = task
	}
}

func (s *gronService) Middleware(middlewares ...Middleware) {
	for _, middleware := range middlewares {
		if middleware != nil {
			s.middlewares = append(s.middlewares, middleware)
		}
	}
}

func (s *gronService) Tasks() map[string]Task {
	return s.tasks
}

func (s *gronService) Call(name string) (err error) {
	if task, ok := s.tasks[name]; ok {
		s.createTaskFunc(task)()
		return nil
	}
	return errors.New(`task "` + name + `" not found`)
}
