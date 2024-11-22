package tracer

import (
	"github.com/N-Vokhmyanin/go-framework/contracts"
	"github.com/N-Vokhmyanin/go-framework/cron"
	"github.com/N-Vokhmyanin/go-framework/queue"
	"github.com/N-Vokhmyanin/go-framework/tracer/trace"
	"github.com/N-Vokhmyanin/go-framework/transport"
	_ "google.golang.org/grpc"
)

type provider struct {
	cfg *Config
}

var _ contracts.Provider = (*provider)(nil)

func NewProvider() contracts.Provider {
	return &provider{
		cfg: &Config{},
	}
}

func (p *provider) Config(c contracts.ConfigSet) {
	c.BoolVar(&p.cfg.Enabled, "TRACER_ENABLED", false, "traces enabled")
	c.StringVar(&p.cfg.Endpoint, "TRACER_ENDPOINT", "opentelemetry:4317", "opentelemetry grpc port")
	c.Float64Var(&p.cfg.SampleRate, "TRACER_SAMPLE_RATE", 1.0, "traces sample rate")
}

func (p *provider) Boot(a contracts.Application) {
	a.Singleton(func() *Config {
		return p.cfg
	})

	a.Singleton(func(cfg *Config) trace.Tracer {
		return NewOpenTelemetryTracer(a.Name(), cfg)
	})
}

func (p *provider) Register(a contracts.Application) {
	a.Make(func(
		tracer trace.Tracer,
		httpGateway transport.HttpGateway,
		grpcServer transport.GrpcServer,
		queueManager queue.Manager,
		cronService cron.Service,
	) {
		tracerProvider := tracer.TracerProvider()

		if httpGateway != nil {
			httpGateway.WithOptions(
				transport.WithMatchHeaders(propagators.Fields()...),
				transport.WithServerMuxOptions(OutgoingHeaderMatcher),
			)
		}

		if grpcServer != nil {
			grpcServer.WithOptions(
				transport.WithPrependUnaryInterceptors(
					ContextTracerUnaryServerInterceptor(tracerProvider),
					TraceIdUnaryServerInterceptor(),
				),
				transport.WithPrependStreamInterceptors(
					ContextTracerStreamServerInterceptor(tracerProvider),
					TraceIdStreamServerInterceptor(),
				),
			)
		}

		if queueManager != nil {
			queueManager.Middleware(
				ContextTracerQueueHandlerMiddleware(tracerProvider),
			)
		}

		if cronService != nil {
			cronService.Middleware(
				ContextTracerCronTaskMiddleware(tracerProvider),
			)
		}
	})
}
