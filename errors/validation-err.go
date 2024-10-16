package errors

import (
	"fmt"
	"github.com/cockroachdb/errors/errbase"
)

type FieldValidationErr struct {
	*Stack
	reason string
	field  string
	cause  error
}

func NewFieldValidationError(field, format string, args ...any) *FieldValidationErr {
	return &FieldValidationErr{
		Stack:  Callers(1),
		field:  field,
		reason: fmt.Sprintf(format, args...),
	}
}

func (e *FieldValidationErr) Error() string {
	return e.Reason()
}

func (e *FieldValidationErr) Reason() string {
	if e == nil {
		return ""
	}
	return e.reason
}

func (e *FieldValidationErr) Field() string {
	if e == nil {
		return ""
	}
	return e.field
}

func (e *FieldValidationErr) WithCause(cause error) *FieldValidationErr {
	if e == nil {
		return nil
	}
	e.cause = cause
	return e
}

func (e *FieldValidationErr) Cause() error {
	if e == nil {
		return nil
	}
	return e.cause
}

// Unwrap provides compatibility for Go 1.13 error chains.
func (e *FieldValidationErr) Unwrap() error {
	return e.Cause()
}

func (e *FieldValidationErr) Format(s fmt.State, verb rune) { errbase.FormatError(e, s, verb) }
