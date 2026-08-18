package log

import (
	"context"
	"log/slog"

	"github.com/Drathveloper/go-web-skeleton/common/constants"
)

func ContextLogger(ctx context.Context) *slog.Logger {
	value := ctx.Value(constants.LoggerContextKey)
	if value == nil {
		return slog.Default()
	}
	logger, ok := value.(*slog.Logger)
	if !ok {
		return slog.Default()
	}
	return logger
}

func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, constants.LoggerContextKey, logger)
}
