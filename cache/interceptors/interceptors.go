package interceptors

import (
	"context"
	cacheAdapters "github.com/N-Vokhmyanin/go-framework/cache/adapters"
	"github.com/N-Vokhmyanin/go-framework/cache/ctxcache"
	"github.com/N-Vokhmyanin/go-framework/cron"
	"github.com/N-Vokhmyanin/go-framework/logger"
	"github.com/N-Vokhmyanin/go-framework/queue"
	"google.golang.org/grpc"
)

func ContextCacheUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		memoryCacheInstance := cacheAdapters.NewAdapter(cacheAdapters.NewMemoryCache(), "")
		defer func() { _ = memoryCacheInstance.Clear(ctx) }()
		return handler(ctxcache.ToContext(ctx, memoryCacheInstance), req)
	}
}

func ContextCacheTaskMiddleware() cron.Middleware {
	return func(ctx context.Context, log logger.Logger, task cron.Task) error {
		memoryCacheInstance := cacheAdapters.NewAdapter(cacheAdapters.NewMemoryCache(), "")
		defer func() { _ = memoryCacheInstance.Clear(ctx) }()
		return task.Handle(ctxcache.ToContext(ctx, memoryCacheInstance), log)
	}
}

func ContextCacheJobHandlerMiddleware() queue.Middleware {
	return func(ctx context.Context, log logger.Logger, i queue.JobInteract, handler queue.Handler) error {
		memoryCacheInstance := cacheAdapters.NewAdapter(cacheAdapters.NewMemoryCache(), "")
		defer func() { _ = memoryCacheInstance.Clear(ctx) }()
		return handler.Handle(ctxcache.ToContext(ctx, memoryCacheInstance), log, i)
	}
}
