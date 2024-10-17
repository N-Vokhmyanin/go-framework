package transport

import (
	"context"
	"github.com/N-Vokhmyanin/go-framework/errors"
	"github.com/N-Vokhmyanin/go-framework/logger/grpc/ctxlog"
)

func RecoverFrom(ctx context.Context, p any) error {
	ctxlog.AddFields(ctx, "is_panic", true)

	return errors.NewPanicError(p)
}
