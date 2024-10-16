package errors

import (
	"fmt"
	"github.com/cockroachdb/errors/errbase"
)

type PreconditionErr struct {
	*Stack
	subject     string
	code        string
	description string
	cause       error
}

func NewPreconditionError(subject, code, description string, args ...any) *PreconditionErr {
	return &PreconditionErr{
		Stack:       Callers(1),
		subject:     subject,
		code:        code,
		description: fmt.Sprintf(description, args...),
	}
}

func (e *PreconditionErr) Error() string {
	return fmt.Sprintf("%s: %s: %s", e.Subject(), e.Code(), e.Description())
}

func (e *PreconditionErr) Code() string {
	if e == nil {
		return ""
	}
	return e.code
}

func (e *PreconditionErr) Subject() string {
	if e == nil {
		return ""
	}
	return e.subject
}

func (e *PreconditionErr) Description() string {
	if e == nil {
		return ""
	}
	return e.description
}

func (e *PreconditionErr) WithCause(cause error) *PreconditionErr {
	if e == nil {
		return e
	}
	e.cause = cause
	return e
}

func (e *PreconditionErr) Cause() error {
	return e.cause
}

// Unwrap provides compatibility for Go 1.13 error chains.
func (e *PreconditionErr) Unwrap() error {
	return e.Cause()
}

func (e *PreconditionErr) Format(s fmt.State, verb rune) { errbase.FormatError(e, s, verb) }
