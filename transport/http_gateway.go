package transport

import (
	"context"
	"github.com/N-Vokhmyanin/go-framework/contracts"
	health "github.com/N-Vokhmyanin/go-framework/health/contracts"
	"github.com/N-Vokhmyanin/go-framework/logger"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	grpcHealthV1 "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/protobuf/encoding/protojson"
	"net"
	"net/http"
	"strings"
)

type httpGateway struct {
	addr string
	grpc string
	log  logger.Logger

	options *httpOptions
	grpcMux *runtime.ServeMux

	serving bool
}

var _ HttpGateway = (*httpGateway)(nil)
var _ health.Service = (*httpGateway)(nil)
var _ contracts.CanBoot = (*httpGateway)(nil)
var _ contracts.CanStart = (*httpGateway)(nil)
var _ contracts.CanStop = (*httpGateway)(nil)

func NewHttpServer(addr, grpc string, log logger.Logger) HttpGateway {
	gw := &httpGateway{
		addr:    addr,
		grpc:    grpc,
		log:     log.With(logger.WithComponent, "transport.http"),
		options: &httpOptions{},
	}
	var serverMuxOptions = []runtime.ServeMuxOption{
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.HTTPBodyMarshaler{
			Marshaler: &runtime.JSONPb{
				MarshalOptions: protojson.MarshalOptions{
					UseProtoNames:   true,
					EmitUnpopulated: true,
				},
				UnmarshalOptions: protojson.UnmarshalOptions{
					DiscardUnknown: true,
				},
			},
		}),
		runtime.WithIncomingHeaderMatcher(func(key string) (string, bool) {
			for _, header := range gw.options.matchHeaders {
				if strings.EqualFold(key, header) {
					return key, true
				}
			}
			return runtime.DefaultHeaderMatcher(key)
		}),
	}
	gw.WithOptions(WithServerMuxOptions(serverMuxOptions...))
	return gw
}

func (s *httpGateway) WithOptions(options ...HttpOption) {
	for _, opt := range options {
		opt(s.options)
	}
}

func (s *httpGateway) HealthStatus(context.Context) grpcHealthV1.HealthCheckResponse_ServingStatus {
	return health.HealthStatusFromBool(s.serving)
}

func (s *httpGateway) BootService() {
	s.grpcMux = runtime.NewServeMux(s.options.serverMuxOptions...)
}

func (s *httpGateway) StartService() {
	ctx := context.Background()

	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		s.log.Fatalw("listen addr failed", "addr", s.addr, zap.Error(err))
	}

	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	conn, err := grpc.NewClient(s.grpc /*grpc.WithBlock(),*/, dialOpts...)
	if err != nil {
		s.log.Fatalw("dial context failed", "target", s.grpc, zap.Error(err))
	}

	for _, registerHandler := range s.options.registerHandlers {
		registerHandler(s.grpcMux, conn)
	}
	for _, h := range s.options.httpHandlers {
		if err = s.grpcMux.HandlePath(h.method, h.pattern, h.HandlerFunc(s.grpcMux)); err != nil {
			s.log.Fatalw("register http handler failed", "pattern", h.pattern, zap.Error(err))
		}
	}

	var handler http.Handler = s.grpcMux
	for _, middleware := range s.options.httpMiddlewares {
		handler = middleware(handler)
	}

	s.log.Infow("starting server...", "addr", listener.Addr().String())
	go func() {
		if conn.WaitForStateChange(ctx, connectivity.Idle) {
			s.serving = true
		}
		if err = http.Serve(listener, handler); err != nil {
			s.log.Fatal(err)
		}
	}()
}

func (s *httpGateway) StopService() {
	s.serving = false
}
