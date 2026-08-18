package handler

import (
	"context"
	"log/slog"

	"github.com/Drathveloper/go-web-skeleton/common/log"
	"github.com/Drathveloper/go-web-skeleton/pkg/event"
)

// attrLogger is the optional contract an event implements to expose its fields
// as structured log attributes. A type assertion instead of reflection keeps
// the hot path allocation-free and an event that does not opt in is still
// logged, just with its name only.
type attrLogger interface {
	LogAttrs() []slog.Attr
}

// AuditLogger writes one structured log line per security event. It is the
// single subscriber behind every name in security/event/dto.AllEventNames.
type AuditLogger struct{}

func NewAuditLogger() *AuditLogger {
	return &AuditLogger{}
}

// Handle always returns nil: a failed audit line must not be retried by the
// bus, because re-running it cannot recover anything.
func (logger *AuditLogger) Handle(ctx context.Context, evt event.Event) error {
	attrs := []slog.Attr{slog.String("event", evt.GetName())}
	if provider, ok := evt.(attrLogger); ok {
		attrs = append(attrs, provider.LogAttrs()...)
	}
	log.ContextLogger(ctx).LogAttrs(ctx, slog.LevelInfo, "security event", attrs...)
	return nil
}
