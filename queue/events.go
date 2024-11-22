package queue

import "time"

type JobPushedEvent struct {
	Job Job
}

type JobDelayedEvent struct {
	Job   Job
	Delay time.Duration
}

type JobStartedEvent struct {
	Job Job
}

type JobFinishedEvent struct {
	Job Job
}

type JobFailedEvent struct {
	Job Job
	Err error
}
