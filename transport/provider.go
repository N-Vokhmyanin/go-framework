package transport

import (
	"fmt"
	"github.com/N-Vokhmyanin/go-framework/contracts"
	"github.com/N-Vokhmyanin/go-framework/logger"
)

type transportProvider struct {
	portHttp uint
	portGrpc uint
}

//goland:noinspection GoUnusedExportedFunction
func NewTransportProvider() contracts.Provider {
	return &transportProvider{}
}

func (p *transportProvider) Config(c contracts.ConfigSet) {
	c.UintVar(&p.portHttp, "SERVER_HTTP_PORT", 8080, "http listener port")
	c.UintVar(&p.portGrpc, "SERVER_GRPC_PORT", 8090, "grpc listener port")
}

func (p *transportProvider) Boot(a contracts.Application) {
	grpcAddr := fmt.Sprintf(":%d", p.portGrpc)
	httpAddr := fmt.Sprintf(":%d", p.portHttp)

	a.Singleton(func(log logger.Logger) GrpcServer {
		return NewGrpcServer(grpcAddr, log)
	})

	a.Singleton(func(log logger.Logger) HttpGateway {
		return NewHttpServer(httpAddr, grpcAddr, log)
	})
}

func (p *transportProvider) Register(a contracts.Application) {
	a.Make(func(log logger.Logger, grpc GrpcServer) {
		a.Command(NewRpcListCommand(grpc))
	})
}
