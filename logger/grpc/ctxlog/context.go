package ctxlog

import (
	"context"

	"github.com/N-Vokhmyanin/go-framework/logger"
	grpcCtxTags "github.com/N-Vokhmyanin/go-framework/logger/grpc/ctxtags"
	"go.uber.org/zap"
)

type ctxMarker struct{}

type ctxLogger struct {
	logger logger.Logger
	fields []interface{}
}

var (
	ctxMarkerKey = &ctxMarker{}
)

// AddFields adds fields to the logger.
//
//goland:noinspection GoUnusedExportedFunction
func AddFields(ctx context.Context, fields ...interface{}) {
	l, ok := ctx.Value(ctxMarkerKey).(*ctxLogger)
	if !ok || l == nil {
		return
	}
	l.fields = append(l.fields, fields...)
}

func uniqueFields(fields []any) []any {
	var otherFields []any
	mapFields := map[string]any{}
	cnt := len(fields)
	skip := false
	for idx, f := range fields {
		if skip {
			skip = false
			continue
		}

		if zf, ok := f.(zap.Field); ok {
			mapFields[zf.Key] = zf
			continue
		}
		if err, ok := f.(error); ok {
			zf := zap.Error(err)
			mapFields[zf.Key] = zf
			continue
		}

		left := cnt - (idx + 1)
		if key, ok := f.(string); ok && left > 0 {
			mapFields[key] = fields[idx+1]
			skip = true
			continue
		}

		otherFields = append(otherFields, f)
	}

	var result []any
	for key, val := range mapFields {
		if zf, ok := val.(zap.Field); ok {
			result = append(result, zf)
		} else {
			result = append(result, key, val)
		}
	}
	result = append(result, otherFields...)

	return result
}

// Extract takes the call-scoped Logger from grpc_kit middleware.
//
// It always returns a Logger that has all the grpc_ctxtags updated.
func Extract(ctx context.Context) logger.Logger {
	l, ok := ctx.Value(ctxMarkerKey).(*ctxLogger)
	if !ok || l == nil {
		return logger.GetNopLogger()
	}
	// Add grpc_ctxtags tags metadata until now.
	fields := TagsToFields(ctx)
	fields = append(fields, MetadataToFields(ctx)...)
	fields = append(fields, l.fields...)
	fields = uniqueFields(fields)
	return l.logger.With(fields...)
}

// ExtractWithFallback Extract takes the call-scoped Logger with fallback Logger
//
//goland:noinspection GoUnusedExportedFunction
func ExtractWithFallback(ctx context.Context, log logger.Logger) logger.Logger {
	l := Extract(ctx)
	np := logger.GetNopLogger()
	if l != np {
		return l
	}
	if log == nil {
		return np
	}
	return log
}

// TagsToFields transforms the Tags on the supplied context into kit fields.
func TagsToFields(ctx context.Context) []interface{} {
	var fields []interface{}
	tags := grpcCtxTags.Extract(ctx)
	for k, v := range tags.Values() {
		fields = append(fields, k, v)
	}
	return fields
}

func MetadataToFields(ctx context.Context) []interface{} {
	var fields []interface{}

	userAgent := GetUserAgentFromContext(ctx)
	if userAgent != "" {
		fields = append(fields, "user_agent", userAgent)
	}

	return fields
}

func ExtractFields(ctx context.Context) (fields []interface{}) {
	l, ok := ctx.Value(ctxMarkerKey).(*ctxLogger)
	if ok && l != nil {
		fields = append(fields, l.fields...)
	}
	fields = append(fields, TagsToFields(ctx)...)
	fields = append(fields, MetadataToFields(ctx)...)
	return fields
}

// ToContext adds the kit.Logger to the context for extraction later.
// Returning the new context that has been created.
func ToContext(ctx context.Context, logger logger.Logger) context.Context {
	l := &ctxLogger{
		logger: logger,
	}
	return context.WithValue(ctx, ctxMarkerKey, l)
}
