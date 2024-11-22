package adapters

import (
	"context"
	"encoding/json"
	"github.com/N-Vokhmyanin/go-framework/cache"
	"github.com/eko/gocache/v2/store"
	"github.com/go-redis/redis/v8"
)

type jsonAdapter struct {
	cache.CacheInterface
}

func NewJsonCache(base cache.CacheInterface) cache.JsonCacheInterface {
	return &jsonAdapter{
		CacheInterface: base,
	}
}

func (c *jsonAdapter) Unmarshal(ctx context.Context, key string, value any) (bool, error) {
	cached, err := c.Get(ctx, key)
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	cachedString, ok := cached.(string)
	if !ok {
		return false, nil
	}

	err = json.Unmarshal([]byte(cachedString), &value)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (c *jsonAdapter) Marshal(ctx context.Context, key string, value any, options *store.Options) error {
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return err
	}
	value = string(valueBytes)
	return c.Set(ctx, key, value, options)
}
