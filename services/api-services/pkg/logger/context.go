package logger

import (
	"context"

	"go.uber.org/zap"
)

type contextKey struct{}

// WithContext returns a new context with the logger attached
func WithContext(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, logger)
}

// FromContext returns the logger from the context or the global logger if none is found
func FromContext(ctx context.Context) *zap.Logger {
	if logger, ok := ctx.Value(contextKey{}).(*zap.Logger); ok {
		return logger
	}
	return Get()
}
