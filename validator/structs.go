package validator

import (
	"github.com/N-Vokhmyanin/go-framework/errors"
	"google.golang.org/grpc/status"
)

type validationWrapperError struct {
	error
}

func (e *validationWrapperError) Unwrap() error {
	return e.error
}

func (e *validationWrapperError) Cause() error {
	return e.error
}

func (e *validationWrapperError) GRPCStatus() *status.Status {
	return errors.ErrorToGRPCStatus(e.error)
}
