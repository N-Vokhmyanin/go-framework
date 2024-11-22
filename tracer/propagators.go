package tracer

import (
	"go.opentelemetry.io/contrib/propagators/jaeger"
	"go.opentelemetry.io/otel/propagation"
)

var propagators = propagation.NewCompositeTextMapPropagator(
	propagation.TraceContext{},
	propagation.Baggage{},
	jaeger.Jaeger{},
)
