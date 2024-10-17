package validator

import "github.com/N-Vokhmyanin/go-framework/errors"

// ValidationError deprecated, use errors.FieldValidationError
type ValidationError interface {
	errors.FieldValidationError
}

// ValidationMultiError deprecated, use errors.MultiError
type ValidationMultiError interface {
	errors.MultiError
}

type singleValidator interface {
	Validate() error
}

type allValidator interface {
	ValidateAll() error
}
