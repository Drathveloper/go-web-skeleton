package bootstrap

import (
	"log/slog"
	"os"
	"slices"
	"strings"

	"github.com/Drathveloper/go-web-skeleton/common/wire"
)

func setupLogger() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)
}

func configureLogger(container *wire.Container) {
	confidentialFields := container.Store.GetLoggingConfidentialFields()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: false,
		Level:     container.Store.GetLoggingLevel(),
		ReplaceAttr: func(_ []string, attr slog.Attr) slog.Attr {
			if slices.Contains(confidentialFields, attr.Key) {
				attr.Value = slog.StringValue(strings.Repeat("*", len(attr.Value.String())))
			}
			return attr
		},
	}))
	slog.SetDefault(logger)
}
