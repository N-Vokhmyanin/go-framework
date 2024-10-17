package transport

import (
	"context"
	"github.com/N-Vokhmyanin/go-framework/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type GrpcStatus interface {
	GRPCStatus() *status.Status
}

type GrpcStatusMappingFunc func(context.Context, error) *status.Status

func (m GrpcStatusMappingFunc) statusError(ctx context.Context, err error) error {
	return m(ctx, err).Err()
}

func (m GrpcStatusMappingFunc) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		if resp, err = handler(ctx, req); err != nil {
			return nil, m.statusError(ctx, err)
		}
		return resp, err
	}
}

func (m GrpcStatusMappingFunc) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		if err = handler(srv, ss); err != nil {
			return m.statusError(ss.Context(), err)
		}
		return err
	}
}

var DefaultGrpcStatusMapping GrpcStatusMappingFunc = func(_ context.Context, err error) *status.Status {
	return errors.ErrorToGRPCStatus(err)
}
