package health

import (
	"github.com/N-Vokhmyanin/go-framework/contracts"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	grpcHealthV1 "google.golang.org/grpc/health/grpc_health_v1"
)

type healthClient struct {
	address     string
	dialOptions []grpc.DialOption
	grpcHealthV1.HealthClient
}

var _ contracts.CanStart = (*healthClient)(nil)
var _ grpcHealthV1.HealthClient = (*healthClient)(nil)

func NewHealthClient(addr string) grpcHealthV1.HealthClient {
	return &healthClient{
		address: addr,
		dialOptions: []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		},
	}
}

func (c *healthClient) StartService() {
	cc, err := grpc.NewClient(c.address, c.dialOptions...)
	if err != nil {
		panic(err)
	}
	c.HealthClient = grpcHealthV1.NewHealthClient(cc)
}
