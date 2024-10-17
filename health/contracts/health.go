package contracts

import (
	"context"
	"github.com/N-Vokhmyanin/go-framework/contracts"
	grpcHealthV1 "google.golang.org/grpc/health/grpc_health_v1"
)

type Service interface {
	HealthStatus(ctx context.Context) grpcHealthV1.HealthCheckResponse_ServingStatus
}

type Transport interface {
	contracts.CanBoot
	contracts.CanStart
	contracts.CanStop
}

func HealthStatusFromBool(val bool) grpcHealthV1.HealthCheckResponse_ServingStatus {
	if val {
		return grpcHealthV1.HealthCheckResponse_SERVING
	} else {
		return grpcHealthV1.HealthCheckResponse_NOT_SERVING
	}
}
