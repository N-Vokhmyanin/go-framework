package queue

import (
	"encoding/json"
	"time"
)

type amqpJobInteract struct {
	jobWrapper *amqpJobWrapper
	resultData amqpJobResult

	releaseFlag  bool
	deleteFlag   bool
	releaseDelay *time.Duration
	failedErr    error
}

var _ JobInteract = (*amqpJobInteract)(nil)

func newAmqpJobInteract(jobWrapper *amqpJobWrapper) *amqpJobInteract {
	return &amqpJobInteract{jobWrapper: jobWrapper}
}

// Unmarshal job body into struct
func (i *amqpJobInteract) Unmarshal(j interface{}) error {
	return json.Unmarshal([]byte(i.jobWrapper.JobBody), &j)
}

// WithResult done job with result
func (i *amqpJobInteract) WithResult(r map[string]string) error {
	i.resultData = r
	return nil
}

// GetResult get result of previous job
func (i *amqpJobInteract) GetResult(key string) string {
	return i.jobWrapper.Result.Get(key)
}

// Release the job back into the queue.
func (i *amqpJobInteract) Release(delay uint) error {
	delayTime := time.Duration(delay) * time.Second
	i.releaseFlag = true
	i.releaseDelay = &delayTime
	return nil
}

func (i *amqpJobInteract) Fail(err error) error {
	i.deleteFlag = true
	i.failedErr = err
	return err
}

func (i *amqpJobInteract) Delete() error {
	i.deleteFlag = true
	return nil
}

func (i *amqpJobInteract) Attempts() uint {
	return i.jobWrapper.Attempts
}

func (i *amqpJobInteract) Body() []byte {
	return []byte(i.jobWrapper.JobBody)
}

func (i *amqpJobInteract) IsFailed() bool {
	return i.failedErr != nil
}
