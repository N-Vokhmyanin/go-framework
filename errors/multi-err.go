package errors

import (
	"fmt"
	"github.com/hashicorp/go-multierror"
	"strings"
	"sync"
)

type MultiErr struct {
	sync.Mutex
	me      *multierror.Error
	message string
}

func NewMultiError() *MultiErr {
	return &MultiErr{
		me: &multierror.Error{
			ErrorFormat: multiErrorFormat,
		},
	}
}

func (e *MultiErr) WithMessage(message string) *MultiErr {
	if e == nil {
		return nil
	}
	e.Lock()
	defer e.Unlock()
	e.message = message
	return e
}

func (e *MultiErr) Append(errs ...error) *MultiErr {
	if e == nil {
		return e
	}
	e.Lock()
	defer e.Unlock()
	e.me = multierror.Append(e.me, errs...)
	return e
}

func (e *MultiErr) Error() string {
	if e == nil {
		return ""
	}
	e.Lock()
	defer e.Unlock()
	if e.message != "" {
		return e.message
	}
	return e.me.Error()
}

func (e *MultiErr) AllErrors() []error {
	if e == nil {
		return nil
	}
	e.Lock()
	defer e.Unlock()
	return e.me.Errors
}

func (e *MultiErr) HasErrors() bool {
	return len(e.AllErrors()) > 0
}

func (e *MultiErr) ErrorOrNil() error {
	return e.me.ErrorOrNil()
}

// Errors zap.errorGroup interface
func (e *MultiErr) Errors() []error {
	return e.AllErrors()
}

func (e *MultiErr) As(target interface{}) bool {
	for _, ee := range e.me.Errors {
		if As(ee, target) {
			return true
		}
	}
	return false
}

var multiErrorFormat multierror.ErrorFormatFunc = func(es []error) string {
	if len(es) == 1 {
		return fmt.Sprintf("1 error occurred: %s", es[0].Error())
	}
	points := make([]string, len(es))
	for i, err := range es {
		points[i] = err.Error()
	}
	return fmt.Sprintf("%d errors occurred: %s", len(es), strings.Join(points, "; "))
}
