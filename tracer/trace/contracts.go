package trace

import (
	"go.opentelemetry.io/otel/trace"
)

type (
	Span            = trace.Span
	SpanStartOption = trace.SpanStartOption
	SpanEndOption   = trace.SpanEndOption
)

type TracerProvider = trace.TracerProvider

type Tracer interface {
	TracerProvider() TracerProvider
}
