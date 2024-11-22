package adapters

import (
	"context"
	"fmt"
	"github.com/N-Vokhmyanin/go-framework/cache"
	goCache "github.com/eko/gocache/v2/cache"
	"github.com/eko/gocache/v2/store"
)

type adapter struct {
	prefix string
	base   goCache.CacheInterface
}

var _ cache.CacheInterface = (*adapter)(nil)

func NewAdapter(base goCache.CacheInterface, prefix string) cache.CacheInterface {
	return &adapter{
		prefix: prefix,
		base:   base,
	}
}

func (a *adapter) key(key string) string {
	if a.prefix == "" {
		return key
	}
	return fmt.Sprintf("%s__%s", a.prefix, key)
}

func (a *adapter) tags(tags []string) []string {
	newTags := make([]string, len(tags))
	for i, tag := range tags {
		newTags[i] = a.key(tag)
	}

	if a.prefix != "" {
		newTags = append(newTags, a.prefix)
	}

	return newTags
}

func (a *adapter) options(options *store.Options) *store.Options {
	if options == nil {
		return &store.Options{
			Tags: a.tags([]string{}),
		}
	}
	return &store.Options{
		Cost:       options.Cost,
		Expiration: options.Expiration,
		Tags:       a.tags(options.Tags),
	}
}

func (a *adapter) Get(ctx context.Context, key string) (interface{}, error) {
	return a.base.Get(ctx, a.key(key))
}

func (a *adapter) Set(ctx context.Context, key string, object interface{}, options *store.Options) error {
	return a.base.Set(ctx, a.key(key), object, a.options(options))
}

func (a *adapter) Delete(ctx context.Context, key string) error {
	return a.base.Delete(ctx, a.key(key))
}

func (a *adapter) Invalidate(ctx context.Context, options store.InvalidateOptions) error {
	return a.base.Invalidate(ctx, store.InvalidateOptions{
		Tags: a.tags(options.Tags),
	})
}

func (a *adapter) Clear(ctx context.Context) error {
	if a.prefix != "" {
		return a.base.Invalidate(ctx, store.InvalidateOptions{
			Tags: []string{a.prefix},
		})
	}
	return a.base.Clear(ctx)
}

func (a *adapter) GetType() string {
	return a.base.GetType()
}

func (a *adapter) Base() goCache.CacheInterface {
	return a.base
}
