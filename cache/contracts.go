package cache

import (
	"context"
	"github.com/bsm/redislock"
	"github.com/eko/gocache/v2/cache"
	"github.com/eko/gocache/v2/store"
	"time"
)

type CacheInterface interface {
	Get(ctx context.Context, key string) (any, error)
	Set(ctx context.Context, key string, value any, options *store.Options) error
	Delete(ctx context.Context, key string) error
	Invalidate(ctx context.Context, options store.InvalidateOptions) error
	Clear(ctx context.Context) error
	GetType() string
	Base() cache.CacheInterface
}

type JsonCacheInterface interface {
	CacheInterface
	Unmarshal(ctx context.Context, key string, value any) (bool, error)
	Marshal(ctx context.Context, key string, value any, options *store.Options) error
}

type Locker interface {
	Obtain(ctx context.Context, key string, ttl time.Duration, opt *redislock.Options) (*redislock.Lock, error)
}
