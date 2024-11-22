package adapters

import (
	"context"
	goCache "github.com/eko/gocache/v2/cache"
	"github.com/eko/gocache/v2/store"
)

var NoopStoreInstance = &noopStore{}
var NoopStoreAdapter = NewAdapter(NoopStoreInstance, "")

type noopStore struct{}

var _ goCache.CacheInterface = (*noopStore)(nil)

func (noopStore) Get(context.Context, interface{}) (interface{}, error) {
	return nil, nil
}

func (noopStore) Set(context.Context, interface{}, interface{}, *store.Options) error {
	return nil
}

func (noopStore) Delete(context.Context, interface{}) error {
	return nil
}

func (noopStore) Invalidate(context.Context, store.InvalidateOptions) error {
	return nil
}

func (noopStore) Clear(context.Context) error {
	return nil
}

func (noopStore) GetType() string {
	return "noop"
}
