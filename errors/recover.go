package errors

import (
	"fmt"
)

type PanicError struct {
	*Stack
	Msg any
}

func NewPanicError(msg any) error {
	return &PanicError{
		Stack: Callers(1),
		Msg:   msg,
	}
}

func (e PanicError) Error() string {
	return fmt.Sprintf("%v", e.Msg)
}
