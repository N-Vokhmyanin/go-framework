package trace

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type ctxTracer struct{}

var ctxTracerKey = &ctxTracer{}

func Extract(ctx context.Context) trace.Tracer {
	tr, ok := ctx.Value(ctxTracerKey).(trace.Tracer)
	if !ok {
		return otel.GetTracerProvider().Tracer("")
	}
	return tr
}

func ToContext(ctx context.Context, tracer trace.Tracer) context.Context {
	return context.WithValue(ctx, ctxTracerKey, tracer)
}
