package queue

import (
	"time"
)

type amqpJobWrapper struct {
	JobQueue string         `json:"queue"`
	JobName  string         `json:"name"`
	JobBody  string         `json:"body"`
	Attempts uint           `json:"attempts"`
	Options  amqpJobOptions `json:"options"`
	Result   amqpJobResult  `json:"result,omitempty"`
}

type amqpJobOptions struct {
	Hash        string           `json:"hash,omitempty"`
	MaxAttempts uint             `json:"max_attempts"`
	DelayTime   time.Duration    `json:"delay_time,omitempty"`
	After       []amqpJobWrapper `json:"after,omitempty"`
	Fails       []amqpJobWrapper `json:"fails,omitempty"`
	Always      []amqpJobWrapper `json:"always,omitempty"`
}

func (opts amqpJobOptions) GetDelay() time.Duration {
	if opts.DelayTime > 0 {
		return opts.DelayTime
	}
	return 0
}

type amqpJobResult map[string]string

func (r amqpJobResult) Get(key string) string {
	if r == nil {
		return ""
	}
	return r[key]
}

var _ Job = (*amqpJobWrapper)(nil)

func wrap(job Job) (wrapper amqpJobWrapper, err error) {
	body, err := job.Body()
	if err != nil {
		return wrapper, err
	}

	wrapper = amqpJobWrapper{
		JobQueue: job.Queue(),
		JobName:  job.Name(),
		JobBody:  string(body),
	}

	var jobOpts jobOptions
	if jobWithOpts, ok := job.(JobWithOptions); ok {
		for _, opt := range jobWithOpts.Options() {
			opt(&jobOpts)
		}
	}

	wrapper.Options.MaxAttempts = jobOpts.MaxAttempts
	wrapper.Options.DelayTime = jobOpts.DelayTime
	wrapper.Options.Hash = jobOpts.Hash

	if wrapper.Options.After, err = wrapSlice(jobOpts.After); err != nil {
		return wrapper, err
	}
	if wrapper.Options.Fails, err = wrapSlice(jobOpts.Fails); err != nil {
		return wrapper, err
	}
	if wrapper.Options.Always, err = wrapSlice(jobOpts.Always); err != nil {
		return wrapper, err
	}

	return wrapper, nil
}

func wrapSlice(jobs []Job) (wrappers []amqpJobWrapper, err error) {
	for _, job := range jobs {
		wrapper, wrapErr := wrap(job)
		if wrapErr != nil {
			return nil, wrapErr
		}
		wrappers = append(wrappers, wrapper)
	}
	return wrappers, nil
}

func (w *amqpJobWrapper) Name() string {
	return w.JobName
}

func (w *amqpJobWrapper) Queue() string {
	return w.JobQueue
}

func (w *amqpJobWrapper) Body() ([]byte, error) {
	return []byte(w.JobBody), nil
}
