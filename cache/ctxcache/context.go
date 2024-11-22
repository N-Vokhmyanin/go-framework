package ctxcache

import (
	"context"
	"time"

	"github.com/N-Vokhmyanin/go-framework/cache"
	"github.com/N-Vokhmyanin/go-framework/cache/adapters"
	"github.com/eko/gocache/v2/store"
)

type ctxCacheStore struct{}

var ctxCacheStoreKey = &ctxCacheStore{}
var ctxDefaultTtl = time.Minute * 30

//goland:noinspection GoUnusedExportedFunction
func Extract(ctx context.Context) cache.CacheInterface {
	s, ok := ctx.Value(ctxCacheStoreKey).(cache.CacheInterface)
	if !ok {
		return adapters.NoopStoreAdapter
	}
	return s
}

func ToContext(ctx context.Context, cache cache.CacheInterface) context.Context {
	return context.WithValue(ctx, ctxCacheStoreKey, cache)
}

func Set(ctx context.Context, key string, value any) {
	SetWithTTL(ctx, key, value, ctxDefaultTtl)
}

func SetWithTTL(ctx context.Context, key string, value any, ttl time.Duration) {
	ch := Extract(ctx)
	if ch == adapters.NoopStoreAdapter {
		return
	}
	_ = ch.Set(ctx, key, value, &store.Options{
		Expiration: ttl,
	})
}

func Get[OUT any](ctx context.Context, key string) OUT {
	out, _ := GetOK[OUT](ctx, key)
	return out
}

func GetOK[OUT any](ctx context.Context, key string) (out OUT, ok bool) {
	ch := Extract(ctx)
	if ch == adapters.NoopStoreAdapter {
		return out, false
	}
	v, err := ch.Get(ctx, key)
	if err != nil {
		return out, false
	}
	out, ok = v.(OUT)
	return out, ok
}

func Delete(ctx context.Context, key string) {
	ch := Extract(ctx)
	if ch == adapters.NoopStoreAdapter {
		return
	}
	_ = ch.Delete(ctx, key)
}
