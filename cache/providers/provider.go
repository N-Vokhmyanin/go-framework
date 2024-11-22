package providers

import (
	"github.com/N-Vokhmyanin/go-framework/cache"
	cacheAdapters "github.com/N-Vokhmyanin/go-framework/cache/adapters"
	cacheInterceptors "github.com/N-Vokhmyanin/go-framework/cache/interceptors"
	"github.com/N-Vokhmyanin/go-framework/contracts"
	"github.com/N-Vokhmyanin/go-framework/cron"
	"github.com/N-Vokhmyanin/go-framework/queue"
	"github.com/N-Vokhmyanin/go-framework/transport"
	"github.com/bsm/redislock"
	goCache "github.com/eko/gocache/v2/cache"
	"github.com/eko/gocache/v2/store"
	"github.com/go-redis/redis/v8"
)

type provider struct {
	redisAddr string
	redisDb   int

	withGrpcInterceptor       bool
	withJobHandlerMiddleware  bool
	withCronHandlerMiddleware bool
}

var _ contracts.Provider = (*provider)(nil)

//goland:noinspection GoUnusedExportedFunction
func NewProvider() contracts.Provider {
	return &provider{}
}

func (p *provider) WithGrpcInterceptor() *provider {
	p.withGrpcInterceptor = true
	return p
}

func (p *provider) WithJobHandlerMiddleware() *provider {
	p.withJobHandlerMiddleware = true
	return p
}

func (p *provider) WithCronHandlerMiddleware() *provider {
	p.withCronHandlerMiddleware = true
	return p
}

func (p *provider) WithAllInterceptors() *provider {
	p.withGrpcInterceptor = true
	p.withJobHandlerMiddleware = true
	p.withCronHandlerMiddleware = true
	return p
}

func (p *provider) Config(c contracts.ConfigSet) {
	c.StringVar(&p.redisAddr, "REDIS_ADDR", "localhost:6379", "redis address with port")
	c.IntVar(&p.redisDb, "REDIS_DB", 0, "redis db number")
}

func (p *provider) Boot(a contracts.Application) {
	a.Singleton(func() *redis.Client {
		redisOptions := &redis.Options{
			Addr: p.redisAddr,
			DB:   p.redisDb,
		}
		return redis.NewClient(redisOptions)
	})

	a.Singleton(func(redisClient *redis.Client) goCache.CacheInterface {
		return goCache.New(store.NewRedis(redisClient, nil))
	})

	a.Singleton(func(base goCache.CacheInterface) cache.CacheInterface {
		return cacheAdapters.NewAdapter(base, a.Name())
	})

	a.Singleton(cacheAdapters.NewJsonCache)

	a.Singleton(func(redisClient *redis.Client) cache.Locker {
		return redislock.New(redisClient)
	})
}

func (p *provider) Register(a contracts.Application) {
	if p.withGrpcInterceptor {
		a.Make(func(grpcServer transport.GrpcServer) {
			if grpcServer == nil {
				return
			}
			grpcServer.WithOptions(
				transport.WithUnaryInterceptors(cacheInterceptors.ContextCacheUnaryServerInterceptor()),
			)
		})
	}
	if p.withJobHandlerMiddleware {
		a.Make(func(queueMgr queue.Manager) {
			if queueMgr == nil {
				return
			}
			queueMgr.Middleware(
				cacheInterceptors.ContextCacheJobHandlerMiddleware(),
			)
		})
	}
	if p.withCronHandlerMiddleware {
		a.Make(func(cronSvc cron.Service) {
			if cronSvc == nil {
				return
			}
			cronSvc.Middleware(
				cacheInterceptors.ContextCacheTaskMiddleware(),
			)
		})
	}
}
