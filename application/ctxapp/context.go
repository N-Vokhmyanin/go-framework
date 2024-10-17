package ctxapp

import (
	"context"
	"github.com/N-Vokhmyanin/go-framework/contracts"
)

type ctxMarker struct{}

var (
	ctxMarkerKey = &ctxMarker{}
)

func Extract(ctx context.Context) contracts.Container {
	app, ok := ctx.Value(ctxMarkerKey).(contracts.Container)
	if !ok {
		return nil
	}
	return app
}

func Inject(ctx context.Context, app contracts.Container) context.Context {
	return context.WithValue(ctx, ctxMarkerKey, app)
}
