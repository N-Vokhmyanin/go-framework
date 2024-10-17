package health

import (
	"fmt"
	"github.com/N-Vokhmyanin/go-framework/contracts"
	health "github.com/N-Vokhmyanin/go-framework/health/contracts"
	"github.com/N-Vokhmyanin/go-framework/logger"
	grpcHealthV1 "google.golang.org/grpc/health/grpc_health_v1"
)

type provider struct {
	portGrpc uint
	portHttp uint
}

var _ contracts.Provider = (*provider)(nil)

func NewProvider() contracts.Provider {
	return &provider{}
}

func (p *provider) Config(c contracts.ConfigSet) {
	c.UintVar(&p.portGrpc, "HEALTH_GRPC_PORT", 10860, "grpc listener port")
	c.UintVar(&p.portHttp, "HEALTH_HTTP_PORT", 10861, "http listener port")
}

func (p *provider) Boot(a contracts.Application) {
	grpcAddr := fmt.Sprintf(":%d", p.portGrpc)
	httpAddr := fmt.Sprintf(":%d", p.portHttp)

	a.Singleton(func() grpcHealthV1.HealthServer {
		return NewHealthServer(a)
	})

	a.Singleton(func() grpcHealthV1.HealthClient {
		return NewHealthClient(grpcAddr)
	})

	a.Singleton(func(
		log logger.Logger,
		server grpcHealthV1.HealthServer,
		client grpcHealthV1.HealthClient,
	) health.Transport {
		return NewHealthTransport(log, grpcAddr, httpAddr, server, client)
	})
}

func (p *provider) Register(contracts.Application) {
	// no register
}
