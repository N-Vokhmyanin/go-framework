package queue

import (
	"fmt"
	"github.com/N-Vokhmyanin/go-framework/cache"
	"github.com/N-Vokhmyanin/go-framework/contracts"
	"github.com/N-Vokhmyanin/go-framework/logger"
	"time"
)

type amqpProvider struct {
	host string
	port string
	user string
	pass string

	stoppingTimeout time.Duration
}

var _ contracts.Provider = (*amqpProvider)(nil)

//goland:noinspection GoUnusedExportedFunction
func NewAmqpProvider() contracts.Provider {
	return &amqpProvider{}
}

func (p *amqpProvider) Config(c contracts.ConfigSet) {
	c.StringVar(&p.host, "AMQP_HOST", "rabbitmq", "rabbitmq host")
	c.StringVar(&p.port, "AMQP_PORT", "5672", "rabbitmq port")
	c.StringVar(&p.user, "AMQP_USER", "user", "rabbitmq user")
	c.StringVar(&p.pass, "AMQP_PASS", "password", "rabbitmq password")

	c.DurationVar(&p.stoppingTimeout, "QUEUE_WORKER_STOPPING_TIMEOUT", time.Minute, "worker stopping timeout")
}

func (p *amqpProvider) Boot(a contracts.Application) {
	a.Singleton(
		func(log logger.Logger, dp contracts.Dispatcher, ch cache.CacheInterface) Manager {
			uri := fmt.Sprintf("amqp://%s:%s@%s:%s/", p.user, p.pass, p.host, p.port)
			return NewAmqpManager(uri, log, dp, ch, p.stoppingTimeout)
		},
	)
}

func (p *amqpProvider) Register(contracts.Application) {
	// nothing to register
}
