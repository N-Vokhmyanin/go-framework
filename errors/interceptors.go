package errors

import (
	"context"
	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware/v2"
	"google.golang.org/grpc"
)

type ctxErrorMarker struct{}

var ctxErrorMarkerKey = &ctxErrorMarker{}

type ctxErrorWrapper struct {
	err error
}

func (oe *ctxErrorWrapper) SetError(err error) {
	if oe != nil {
		oe.err = err
	}
}

func (oe *ctxErrorWrapper) GetError() error {
	if oe != nil {
		return oe.err
	}
	return nil
}

func setErrorWrapperToContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxErrorMarkerKey, &ctxErrorWrapper{})
}

func extractErrorWrapperFromContext(ctx context.Context) *ctxErrorWrapper {
	t, ok := ctx.Value(ctxErrorMarkerKey).(*ctxErrorWrapper)
	if !ok {
		return nil
	}
	return t
}

func UnaryPrependServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		return handler(setErrorWrapperToContext(ctx), req)
	}
}

func UnaryAppendServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		res, err := handler(ctx, req)
		if err != nil {
			SetErrorToContext(ctx, err)
		}
		return res, err
	}
}

func StreamPrependServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		wrapped := grpcMiddleware.WrapServerStream(stream)
		wrapped.WrappedContext = setErrorWrapperToContext(stream.Context())
		return handler(srv, wrapped)
	}
}

func StreamAppendServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		err := handler(srv, stream)
		if err != nil {
			SetErrorToContext(stream.Context(), err)
		}
		return err
	}
}

func SetErrorToContext(ctx context.Context, err error) {
	extractErrorWrapperFromContext(ctx).SetError(err)
}

func ExtractErrorFromContext(ctx context.Context) error {
	return extractErrorWrapperFromContext(ctx).GetError()
}
