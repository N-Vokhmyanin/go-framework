package trace

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace/noop"
	"net/http"
)

func Start(ctx context.Context, name string, opts ...SpanStartOption) (context.Context, Span) {
	if ctx == nil {
		return ctx, noop.Span{}
	}
	return Extract(ctx).Start(ctx, name, opts...)
}

func End(span Span, err error, opts ...SpanEndOption) {
	if span == nil {
		return
	}
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
	}
	span.End(opts...)
}

func Http(ctx context.Context, request *http.Request) {
	headers := propagation.HeaderCarrier(request.Header)
	otel.GetTextMapPropagator().Inject(ctx, headers)
	request.Header = http.Header(headers)
}
