package errors

import (
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/runtime/protoiface"
	"strings"
)

func ErrorToFieldViolations(err error, fieldPrefix string) []*errdetails.BadRequest_FieldViolation {
	var violations []*errdetails.BadRequest_FieldViolation

	if err == nil {
		return violations
	}

	var vErr FieldValidationError
	if As(err, &vErr) {
		var cause error
		if vErrWithCause, ok := vErr.(interface{ Cause() error }); ok {
			cause = vErrWithCause.Cause()
		}
		causeViolations := ErrorToFieldViolations(cause, vErr.Field()+".")
		if len(causeViolations) > 0 {
			violations = append(violations, causeViolations...)
		} else {
			violations = append(violations, &errdetails.BadRequest_FieldViolation{
				Field:       fieldPrefix + vErr.Field(),
				Description: vErr.Reason(),
			})
		}
	}

	var vMultiErr MultiError
	if As(err, &vMultiErr) {
		vErrors := vMultiErr.AllErrors()
		if len(vErrors) > 0 {
			for _, subErr := range vErrors {
				violations = append(violations, ErrorToFieldViolations(subErr, fieldPrefix)...)
			}
		} else if fieldPrefix != "" {
			violations = append(violations, &errdetails.BadRequest_FieldViolation{
				Field:       fieldPrefix,
				Description: vMultiErr.Error(),
			})
		}
	}

	return violations
}

func FieldViolationsToMessage(violations []*errdetails.BadRequest_FieldViolation) string {
	messages := make([]string, len(violations))
	for i, violation := range violations {
		if violation.Field == "" {
			messages[i] = violation.Description
		} else {
			messages[i] = violation.Field + ": " + violation.Description
		}
	}
	return strings.Join(messages, "; ")
}

func ErrorToPreconditionViolations(err error) []*errdetails.PreconditionFailure_Violation {
	var violations []*errdetails.PreconditionFailure_Violation

	if err == nil {
		return violations
	}

	var vErr PreconditionError
	if As(err, &vErr) {
		var cause error
		if vErrWithCause, ok := vErr.(interface{ Cause() error }); ok {
			cause = vErrWithCause.Cause()
		}
		causeViolations := ErrorToPreconditionViolations(cause)
		if len(causeViolations) > 0 {
			violations = append(violations, causeViolations...)
		} else {
			violations = append(violations, &errdetails.PreconditionFailure_Violation{
				Type:        vErr.Code(),
				Subject:     vErr.Subject(),
				Description: vErr.Description(),
			})
		}
	}

	var vMultiErr MultiError
	if As(err, &vMultiErr) {
		vErrors := vMultiErr.AllErrors()
		if len(vErrors) > 0 {
			for _, subErr := range vErrors {
				violations = append(violations, ErrorToPreconditionViolations(subErr)...)
			}
		}
	}

	if vErrWithCause, ok := vErr.(interface{ Cause() error }); ok {
		violations = append(violations, ErrorToPreconditionViolations(vErrWithCause.Cause())...)
	}

	return violations
}

func PreconditionViolationsToMessage(violations []*errdetails.PreconditionFailure_Violation) string {
	messages := make([]string, len(violations))
	for i, violation := range violations {
		var parts []string
		if violation.Subject != "" {
			parts = append(parts, violation.Subject)
		}
		if violation.Type != "" {
			parts = append(parts, violation.Type)
		}
		if violation.Description != "" {
			parts = append(parts, violation.Description)
		}
		messages[i] = strings.Join(parts, ": ")
	}
	return strings.Join(messages, "; ")
}

func ErrorToGRPCStatus(err error) *status.Status {

	var errWithGrpcStatus interface {
		GRPCStatus() *status.Status
	}

	if As(err, &errWithGrpcStatus) {
		return errWithGrpcStatus.GRPCStatus()
	}

	var st *status.Status
	var details []protoiface.MessageV1

	var domainErr DomainError
	if As(err, &domainErr) {
		details = append(details, &errdetails.ErrorInfo{
			Reason:   domainErr.Code(),
			Domain:   domainErr.Domain(),
			Metadata: domainErr.Meta(),
		})
		if st == nil {
			st = status.New(domainCodeToGRPCStatusCode(domainErr.Code()), domainErr.Error())
		}
	}

	preconditionViolations := ErrorToPreconditionViolations(err)
	if len(preconditionViolations) > 0 {
		details = append(details, &errdetails.PreconditionFailure{
			Violations: preconditionViolations,
		})

		if st == nil {
			st = status.New(codes.FailedPrecondition, PreconditionViolationsToMessage(preconditionViolations))
		}
	}

	fieldViolations := ErrorToFieldViolations(err, "")
	if len(fieldViolations) > 0 {
		details = append(details, &errdetails.BadRequest{
			FieldViolations: fieldViolations,
		})

		if st == nil {
			st = status.New(codes.InvalidArgument, FieldViolationsToMessage(fieldViolations))
		}
	}

	if st == nil {
		st, _ = status.FromError(err)
	}

	st, _ = st.WithDetails(details...)

	return st
}

func domainCodeToGRPCStatusCode(code string) codes.Code {
	switch code {
	case DomainCodeValidationErr:
		return codes.InvalidArgument
	case DomainCodeNotFoundErr:
		return codes.NotFound
	case DomainCodePermissionDeniedErr:
		return codes.PermissionDenied
	case DomainCodeUnauthenticatedErr:
		return codes.Unauthenticated
	default:
		return codes.InvalidArgument
	}
}
