package queue

import "time"

type jobOptions struct {
	Hash        string
	MaxAttempts uint
	DelayTime   time.Duration
	After       []Job
	Fails       []Job
	Always      []Job
}

type JobOptionFunc func(o *jobOptions)

//goland:noinspection GoUnusedExportedFunction
func OptMaxAttempts(attempts uint) JobOptionFunc {
	return func(o *jobOptions) {
		o.MaxAttempts = attempts
	}
}

//goland:noinspection GoUnusedExportedFunction
func OptDelay(delay uint) JobOptionFunc {
	return OptDelayTime(time.Duration(delay) * time.Second)
}

//goland:noinspection GoUnusedExportedFunction
func OptDelayTime(delay time.Duration) JobOptionFunc {
	return func(o *jobOptions) {
		o.DelayTime = delay
	}
}

//goland:noinspection GoUnusedExportedFunction
func OptAfter(jobs ...Job) JobOptionFunc {
	return func(o *jobOptions) {
		o.After = append(o.After, jobs...)
	}
}

//goland:noinspection GoUnusedExportedFunction
func OptFails(jobs ...Job) JobOptionFunc {
	return func(o *jobOptions) {
		o.Fails = append(o.Fails, jobs...)
	}
}

//goland:noinspection GoUnusedExportedFunction
func OptAlways(jobs ...Job) JobOptionFunc {
	return func(o *jobOptions) {
		o.Always = append(o.Always, jobs...)
	}
}

//goland:noinspection GoUnusedExportedFunction
func OptOnce(hash string) JobOptionFunc {
	return func(o *jobOptions) {
		o.Hash = hash
	}
}

type jobWithOptions struct {
	job  Job
	opts []JobOptionFunc
}

var _ Job = (*jobWithOptions)(nil)
var _ JobWithOptions = (*jobWithOptions)(nil)

//goland:noinspection GoUnusedExportedFunction
func WithOptions(job Job, opts ...JobOptionFunc) Job {
	j := &jobWithOptions{
		job: job,
	}
	if onceJob, ok := job.(OnceJob); ok {
		j.opts = append(j.opts, OptOnce(onceJob.Hash()))
	}
	if jobOpts, ok := job.(JobWithOptions); ok {
		j.opts = append(j.opts, jobOpts.Options()...)
	}
	j.opts = append(j.opts, opts...)
	return j
}

func (j *jobWithOptions) Name() string {
	return j.job.Name()
}

func (j *jobWithOptions) Queue() string {
	return j.job.Queue()
}

func (j *jobWithOptions) Body() ([]byte, error) {
	return j.job.Body()
}

func (j *jobWithOptions) Options() []JobOptionFunc {
	return j.opts
}

//goland:noinspection GoUnusedExportedFunction
func WithChain(jobs ...Job) Job {
	var first = jobs[0]
	var opts []JobOptionFunc
	if len(jobs) > 1 {
		opts = append(opts, OptAfter(WithChain(jobs[1:]...)))
	}
	return WithOptions(first, opts...)
}
