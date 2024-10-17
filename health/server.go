package health

import (
	"context"
	"github.com/N-Vokhmyanin/go-framework/contracts"
	health "github.com/N-Vokhmyanin/go-framework/health/contracts"
	grpcHealthV1 "google.golang.org/grpc/health/grpc_health_v1"
)

type healthServer struct {
	container contracts.Container
}

var _ grpcHealthV1.HealthServer = (*healthServer)(nil)

func NewHealthServer(container contracts.Container) grpcHealthV1.HealthServer {
	return &healthServer{
		container: container,
	}
}

func (s healthServer) Check(ctx context.Context, _ *grpcHealthV1.HealthCheckRequest) (res *grpcHealthV1.HealthCheckResponse, err error) {
	res = &grpcHealthV1.HealthCheckResponse{
		Status: grpcHealthV1.HealthCheckResponse_SERVING,
	}
	instances := s.container.Instances()
	for _, instance := range instances {
		if service, ok := instance.(health.Service); ok {
			res.Status = service.HealthStatus(ctx)
			if res.Status != grpcHealthV1.HealthCheckResponse_SERVING {
				return res, nil
			}
		}
	}
	return res, nil
}

func (s healthServer) Watch(req *grpcHealthV1.HealthCheckRequest, server grpcHealthV1.Health_WatchServer) (err error) {
	res, err := s.Check(server.Context(), req)
	if err != nil {
		return err
	}
	return server.Send(res)
}
