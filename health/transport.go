package health

import (
	"github.com/N-Vokhmyanin/go-framework/contracts"
	health "github.com/N-Vokhmyanin/go-framework/health/contracts"
	"github.com/N-Vokhmyanin/go-framework/logger"
	"github.com/N-Vokhmyanin/go-framework/transport"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	grpcHealthV1 "google.golang.org/grpc/health/grpc_health_v1"
)

type healthTransport struct {
	grpcServer  transport.GrpcServer
	httpGateway transport.HttpGateway
}

var _ health.Transport = (*healthTransport)(nil)

func NewHealthTransport(
	log logger.Logger,
	grpcAddr, httpAddr string,
	server grpcHealthV1.HealthServer,
	client grpcHealthV1.HealthClient,
) health.Transport {
	log = log.With("server", "health")

	grpcServer := transport.NewGrpcServer(grpcAddr, log)
	grpcServer.WithOptions(
		transport.WithRegisterServers(func(s *grpc.Server) {
			grpcHealthV1.RegisterHealthServer(s, server)
		}),
	)

	httpGateway := transport.NewHttpServer(httpAddr, grpcAddr, log)
	httpGateway.WithOptions(
		transport.WithServerMuxOptions(
			runtime.WithHealthzEndpoint(client),
		),
	)

	return &healthTransport{
		grpcServer:  grpcServer,
		httpGateway: httpGateway,
	}
}

func (t healthTransport) BootService() {
	if grpcBoot, ok := t.grpcServer.(contracts.CanBoot); ok {
		grpcBoot.BootService()
	}
	if httpBoot, ok := t.httpGateway.(contracts.CanBoot); ok {
		httpBoot.BootService()
	}
}

func (t healthTransport) StartService() {
	if grpcStart, ok := t.grpcServer.(contracts.CanStart); ok {
		grpcStart.StartService()
	}
	if httpStart, ok := t.httpGateway.(contracts.CanStart); ok {
		httpStart.StartService()
	}
}

func (t healthTransport) StopService() {
	if grpcStop, ok := t.grpcServer.(contracts.CanStop); ok {
		grpcStop.StopService()
	}
	if httpStop, ok := t.httpGateway.(contracts.CanStop); ok {
		httpStop.StopService()
	}
}
