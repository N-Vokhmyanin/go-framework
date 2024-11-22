package trace

import (
	"go.opentelemetry.io/otel/trace"
)

const HeaderTraceId = "X-Trace-ID"

type (
	Span            = trace.Span
	SpanStartOption = trace.SpanStartOption
	SpanEndOption   = trace.SpanEndOption
)

type TracerProvider = trace.TracerProvider

type Tracer interface {
	TracerProvider() TracerProvider
}
