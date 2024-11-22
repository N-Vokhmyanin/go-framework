package tracer

import (
	"context"
	"github.com/N-Vokhmyanin/go-framework/cron"
	"github.com/N-Vokhmyanin/go-framework/logger"
	"github.com/N-Vokhmyanin/go-framework/queue"
	"github.com/N-Vokhmyanin/go-framework/tracer/trace"
	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware/v2"
	"google.golang.org/grpc"
)

func ContextTracerUnaryServerInterceptor(tr trace.TracerProvider) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		tracer := tr.Tracer("grpc" + info.FullMethod)
		return handler(trace.ToContext(ctx, tracer), req)
	}
}

func ContextTracerStreamServerInterceptor(tr trace.TracerProvider) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		tracer := tr.Tracer("grpc" + info.FullMethod)
		wrapped := grpcMiddleware.WrapServerStream(ss)
		wrapped.WrappedContext = trace.ToContext(ss.Context(), tracer)
		return handler(srv, wrapped)
	}
}

func ContextTracerCronTaskMiddleware(tr trace.TracerProvider) cron.Middleware {
	return func(ctx context.Context, log logger.Logger, task cron.Task) error {
		tracer := tr.Tracer("task/" + task.Name())
		return task.Handle(trace.ToContext(ctx, tracer), log)
	}
}

func ContextTracerQueueHandlerMiddleware(tr trace.TracerProvider) queue.Middleware {
	return func(ctx context.Context, log logger.Logger, i queue.JobInteract, handler queue.Handler) error {
		tracer := tr.Tracer("queue/" + handler.Name())
		return handler.Handle(trace.ToContext(ctx, tracer), log, i)
	}
}
