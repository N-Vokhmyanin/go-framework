package adapters

import (
	"context"
	goCache "github.com/eko/gocache/v2/cache"
	"github.com/eko/gocache/v2/store"
	goMemCache "github.com/patrickmn/go-cache"
	"strings"
	"time"
)

const memoryCacheStoreTime = 5 * time.Minute

type memoryCache struct {
	cache goCache.CacheInterface
}

var _ goCache.CacheInterface = (*memoryCache)(nil)

func NewMemoryCache() goCache.CacheInterface {
	goCacheInstance := goMemCache.New(memoryCacheStoreTime, memoryCacheStoreTime)
	return &memoryCache{
		cache: store.NewGoCache(goCacheInstance, nil),
	}
}

func (c memoryCache) Get(ctx context.Context, key interface{}) (res interface{}, err error) {
	res, err = c.cache.Get(ctx, key)
	if isNotFound(err) {
		return nil, nil
	}
	return res, err
}

func (c memoryCache) Set(ctx context.Context, key, object interface{}, options *store.Options) error {
	return c.cache.Set(ctx, key, object, options)
}

func (c memoryCache) Delete(ctx context.Context, key interface{}) error {
	return c.cache.Delete(ctx, key)
}

func (c memoryCache) Invalidate(ctx context.Context, options store.InvalidateOptions) error {
	return c.cache.Invalidate(ctx, options)
}

func (c memoryCache) Clear(ctx context.Context) error {
	return c.cache.Clear(ctx)
}

func (c memoryCache) GetType() string {
	return c.cache.GetType()
}

func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "not found")
}
