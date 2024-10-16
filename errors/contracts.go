package errors

import (
	"github.com/cockroachdb/errors"
)

//goland:noinspection GoUnusedGlobalVariable
var (
	New           = errors.New
	Errorf        = errors.Errorf
	WrapWith      = errors.Wrapf
	WrapWithDepth = errors.WrapWithDepthf
	Wrap          = errors.WithStack
	WrapDepth     = errors.WithStackDepth
	Unwrap        = errors.Unwrap
	Cause         = errors.Cause
	As            = errors.As
	Is            = errors.Is
)

type DomainError interface {
	error
	Code() string
	Domain() string
	Meta() map[string]string
}

type PreconditionError interface {
	error
	Code() string
	Subject() string
	Description() string
}

type FieldValidationError interface {
	error
	Field() string
	Reason() string
}

type MultiError interface {
	error
	AllErrors() []error
}

func AsErr[T any](err error) (out T) {
	if err == nil {
		return
	}
	if As(err, &out) {
		return out
	}
	return
}

func AsErrOk[T any](err error) (out T, ok bool) {
	if err == nil {
		return
	}
	if As(err, &out) {
		return out, true
	}
	return out, false
}

func IsErr[T any](err error) bool {
	if err == nil {
		return false
	}
	var out T
	return As(err, &out)
}
