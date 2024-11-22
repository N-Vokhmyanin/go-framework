package tracer

import (
	"context"
	"github.com/N-Vokhmyanin/go-framework/contracts"
	"github.com/N-Vokhmyanin/go-framework/tracer/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace/noop"
	"time"
)

type Config struct {
	Enabled    bool
	Endpoint   string
	SampleRate float64
}

type openTelemetryTracer struct {
	traceExporter *otlptrace.Exporter
	traceProvider trace.TracerProvider
}

var _ trace.Tracer = (*openTelemetryTracer)(nil)
var _ contracts.CanInit = (*openTelemetryTracer)(nil)
var _ contracts.CanStop = (*openTelemetryTracer)(nil)

func NewOpenTelemetryTracer(serviceName string, cfg *Config) trace.Tracer {
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(cfg.Endpoint),
		otlptracegrpc.WithReconnectionPeriod(100 * time.Millisecond),
	}

	traceExporter := otlptracegrpc.NewUnstarted(opts...)

	traceResource, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		panic(err)
	}

	var traceProvider trace.TracerProvider
	if cfg.Enabled {
		traceProvider = sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(traceExporter),
			sdktrace.WithResource(traceResource),
			sdktrace.WithSampler(
				sdktrace.ParentBased(
					sdktrace.TraceIDRatioBased(cfg.SampleRate),
				),
			),
		)
	} else {
		traceProvider = noop.NewTracerProvider()
	}

	otel.SetTracerProvider(traceProvider)
	otel.SetTextMapPropagator(propagators)

	return &openTelemetryTracer{
		traceExporter: traceExporter,
		traceProvider: traceProvider,
	}
}

func (t openTelemetryTracer) TracerProvider() trace.TracerProvider {
	return t.traceProvider
}

func (t openTelemetryTracer) InitService() {
	if t.traceExporter != nil {
		err := t.traceExporter.Start(context.Background())
		if err != nil {
			panic(err)
		}
	}
}

func (t openTelemetryTracer) StopService() {
	if t.traceExporter != nil {
		_ = t.traceExporter.Shutdown(context.Background())
	}
}
