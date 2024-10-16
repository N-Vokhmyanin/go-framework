package errors

import (
	"fmt"
	"github.com/cockroachdb/errors/errbase"
)

const (
	DomainCodeNotFoundErr         = "not_found"
	DomainCodeUnauthenticatedErr  = "unauthenticated"
	DomainCodePermissionDeniedErr = "permission_denied"
	DomainCodeValidationErr       = "validation"
)

type DomainErr struct {
	*Stack
	code    string
	domain  string
	message string
	meta    map[string]string
	cause   error
}

func newDomainErrorInner(code, message string, args ...any) *DomainErr {
	return &DomainErr{
		code:    code,
		message: fmt.Sprintf(message, args...),
	}
}

func NewDomainError(code, message string, args ...any) *DomainErr {
	return newDomainErrorInner(code, message, args...).WithStack(Callers(1))
}

func NewUnauthenticatedError(message string, args ...any) *DomainErr {
	return newDomainErrorInner(DomainCodeUnauthenticatedErr, message, args...).WithStack(Callers(1))
}

func NewPermissionDeniedError(message string, args ...any) *DomainErr {
	return newDomainErrorInner(DomainCodePermissionDeniedErr, message, args...).WithStack(Callers(1))
}

func NewNotFoundError(message string, args ...any) *DomainErr {
	return newDomainErrorInner(DomainCodeNotFoundErr, message, args...).WithStack(Callers(1))
}

func NewValidationError(message string, args ...any) *DomainErr {
	return newDomainErrorInner(DomainCodeValidationErr, message, args...).WithStack(Callers(1))
}

func (e *DomainErr) Error() string {
	return e.Message()
}

func (e *DomainErr) Code() string {
	if e == nil {
		return ""
	}
	return e.code
}

func (e *DomainErr) Domain() string {
	if e == nil {
		return ""
	}
	return e.domain
}

func (e *DomainErr) Message() string {
	if e == nil {
		return ""
	}
	return e.message
}

func (e *DomainErr) Meta() map[string]string {
	if e == nil {
		return nil
	}
	return e.meta
}

func (e *DomainErr) WithMeta(keyValue ...string) *DomainErr {
	if e == nil {
		return e
	}
	lenIn := len(keyValue)
	if lenIn == 0 {
		return e
	}
	if e.meta == nil {
		e.meta = make(map[string]string)
	}
	for i := 0; i < lenIn/2; i++ {
		keyIndex := i * 2
		valIndex := keyIndex + 1
		if keyIndex < lenIn && valIndex < lenIn {
			e.meta[keyValue[keyIndex]] = keyValue[valIndex]
		}
	}
	return e
}

func (e *DomainErr) WithDomain(domain string) *DomainErr {
	if e == nil {
		return e
	}
	e.domain = domain
	return e
}

func (e *DomainErr) WithCause(cause error) *DomainErr {
	if e == nil {
		return e
	}
	e.cause = cause
	return e
}

func (e *DomainErr) WithStack(stack *Stack) *DomainErr {
	if e == nil {
		return e
	}
	e.Stack = stack
	return e
}

func (e *DomainErr) Cause() error {
	return e.cause
}

// Unwrap provides compatibility for Go 1.13 error chains.
func (e *DomainErr) Unwrap() error {
	return e.Cause()
}

func (e *DomainErr) Format(s fmt.State, verb rune) { errbase.FormatError(e, s, verb) }
