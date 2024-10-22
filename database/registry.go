package database

import (
	"context"
	health "github.com/N-Vokhmyanin/go-framework/health/contracts"
	grpcHealthV1 "google.golang.org/grpc/health/grpc_health_v1"
	"sync"
)

type connectionRegistry struct {
	sync.Mutex
	pool map[string]Connection
}

var _ health.Service = (*connectionRegistry)(nil)

func NewConnectionRegistry() ConnectionRegistry {
	return &connectionRegistry{
		pool: map[string]Connection{},
	}
}

func (r *connectionRegistry) Default() Connection {
	return r.Get("")
}

func (r *connectionRegistry) Get(name string) Connection {
	r.Lock()
	defer r.Unlock()

	conn, ok := r.pool[name]
	if !ok {
		return nil
	}

	return conn
}

func (r *connectionRegistry) Register(name string, conn Connection) {
	r.Lock()
	defer r.Unlock()

	r.pool[name] = conn
}

func (r *connectionRegistry) HealthStatus(context.Context) grpcHealthV1.HealthCheckResponse_ServingStatus {
	r.Lock()
	defer r.Unlock()

	for _, connection := range r.pool {
		if !connection.IsConnected() {
			return grpcHealthV1.HealthCheckResponse_NOT_SERVING
		}
	}

	return grpcHealthV1.HealthCheckResponse_SERVING
}
