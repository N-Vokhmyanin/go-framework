package queue

import "fmt"

type ErrNotConnected struct {
}

func (ErrNotConnected) Error() string {
	return "not connected to the queue"
}

type ErrAlreadyClosed struct {
}

func (ErrAlreadyClosed) Error() string {
	return "already closed: not connected to the queue"
}

type ErrUnknownJob struct {
	Name string
}

func (e ErrUnknownJob) Error() string {
	return fmt.Sprintf("unknown job: %s", e.Name)
}

type ErrUnknownQueue struct {
	Name string
}

func (e ErrUnknownQueue) Error() string {
	return fmt.Sprintf("unknown queue: %s", e.Name)
}
