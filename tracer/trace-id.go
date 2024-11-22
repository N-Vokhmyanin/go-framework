package tracer

import (
	"context"
	"fmt"
	"github.com/N-Vokhmyanin/go-framework/logger/grpc/ctxtags"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"strings"
)

const HeaderTraceId = "X-Trace-ID"

var OutgoingHeaderMatcher = runtime.WithOutgoingHeaderMatcher(func(key string) (string, bool) {
	if key == strings.ToLower(HeaderTraceId) {
		return key, true
	}
	return fmt.Sprintf("%s%s", runtime.MetadataHeaderPrefix, key), true
})

func TraceIdUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		if spanCtx := trace.SpanContextFromContext(ctx); spanCtx.HasTraceID() {
			traceId := spanCtx.TraceID().String()
			ctxtags.Extract(ctx).Set("trace.id", traceId)
			_ = grpc.SetHeader(ctx, metadata.Pairs(HeaderTraceId, traceId))
		}
		return handler(ctx, req)
	}
}

func TraceIdStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if spanCtx := trace.SpanContextFromContext(ss.Context()); spanCtx.HasTraceID() {
			traceId := spanCtx.TraceID().String()
			ctxtags.Extract(ss.Context()).Set("trace.id", traceId)
			_ = ss.SetHeader(metadata.Pairs(HeaderTraceId, traceId))
		}
		return handler(srv, ss)
	}
}
