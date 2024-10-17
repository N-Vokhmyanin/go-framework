package transport

import (
	"context"
	"github.com/N-Vokhmyanin/go-framework/contracts"
	"github.com/N-Vokhmyanin/go-framework/errors"
	health "github.com/N-Vokhmyanin/go-framework/health/contracts"
	"github.com/N-Vokhmyanin/go-framework/logger"
	grpcCtxTags "github.com/N-Vokhmyanin/go-framework/transport/ctxtags"
	"github.com/N-Vokhmyanin/go-framework/validator"
	grpcRecovery "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	grpcHealthV1 "google.golang.org/grpc/health/grpc_health_v1"
	"net"
)

type grpcServer struct {
	addr string
	log  logger.Logger

	options *grpcOptions
	server  *grpc.Server

	serving bool
}

var _ GrpcServer = (*grpcServer)(nil)
var _ health.Service = (*grpcServer)(nil)
var _ contracts.CanBoot = (*grpcServer)(nil)
var _ contracts.CanStart = (*grpcServer)(nil)
var _ contracts.CanStop = (*grpcServer)(nil)

func NewGrpcServer(addr string, log logger.Logger) GrpcServer {
	srv := &grpcServer{
		addr:    addr,
		log:     log.With(logger.WithComponent, "transport.grpc"),
		options: newGrpcOptions(),
	}
	srv.WithOptions(
		WithUnaryInterceptors(
			validator.UnaryServerInterceptor(),
		),
		WithStreamInterceptors(
			validator.StreamServerInterceptor(),
		),
		WithGrpcRecoverOptions(
			grpcRecovery.WithRecoveryHandlerContext(RecoverFrom),
		),
	)
	return srv
}

func (s *grpcServer) GetServiceInfo() map[string]grpc.ServiceInfo {
	return s.server.GetServiceInfo()
}

func (s *grpcServer) WithOptions(options ...GrpcOption) {
	for _, opt := range options {
		opt(s.options)
	}
}

func (s *grpcServer) HealthStatus(context.Context) grpcHealthV1.HealthCheckResponse_ServingStatus {
	return health.HealthStatusFromBool(s.serving)
}

func (s *grpcServer) BootService() {
	var unaryInterceptors = []grpc.UnaryServerInterceptor{
		// ctx_tags обязательно идет первым!
		grpcCtxTags.UnaryServerInterceptor(),
		errors.UnaryPrependServerInterceptor(),
	}
	unaryInterceptors = append(unaryInterceptors, s.options.prependUnaryInterceptors...)
	unaryInterceptors = append(unaryInterceptors, s.options.statusMapping.UnaryServerInterceptor())
	unaryInterceptors = append(unaryInterceptors, s.options.appendUnaryInterceptors...)
	unaryInterceptors = append(
		unaryInterceptors,
		errors.UnaryAppendServerInterceptor(),
		grpcRecovery.UnaryServerInterceptor(s.options.recoverOptions...),
	)

	var streamInterceptors = []grpc.StreamServerInterceptor{
		// ctx_tags обязательно идет первым!
		grpcCtxTags.StreamServerInterceptor(),
		errors.StreamPrependServerInterceptor(),
	}
	streamInterceptors = append(streamInterceptors, s.options.prependStreamInterceptors...)
	streamInterceptors = append(streamInterceptors, s.options.statusMapping.StreamServerInterceptor())
	streamInterceptors = append(streamInterceptors, s.options.appendStreamInterceptors...)
	streamInterceptors = append(
		streamInterceptors,
		errors.StreamAppendServerInterceptor(),
		grpcRecovery.StreamServerInterceptor(s.options.recoverOptions...),
	)

	serverOptions := append(
		s.options.serverOptions,
		grpc.ChainUnaryInterceptor(unaryInterceptors...),
		grpc.ChainStreamInterceptor(streamInterceptors...),
	)
	s.server = grpc.NewServer(serverOptions...)
	for _, registerServer := range s.options.registerServers {
		registerServer(s.server)
	}
}

func (s *grpcServer) StartService() {
	if listener, err := net.Listen("tcp", s.addr); err != nil {
		s.log.Fatalw("listen addr failed", "addr", s.addr, zap.Error(err))
	} else {
		s.log.Infow("starting server...", "addr", listener.Addr().String())
		go func() {
			s.serving = true
			if err = s.server.Serve(listener); err != nil {
				s.log.Fatal(err)
			}
		}()
	}
}

func (s *grpcServer) StopService() {
	s.serving = false
	if s.server != nil {
		s.server.GracefulStop()
	}
}
