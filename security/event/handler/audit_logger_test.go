package handler_test

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Drathveloper/go-web-skeleton/common/log"
	"github.com/Drathveloper/go-web-skeleton/security/event/dto"
	"github.com/Drathveloper/go-web-skeleton/security/event/handler"
)

// bareEvent implements only the bus contract, on purpose: the audit logger must
// still log an event that does not expose LogAttrs.
type bareEvent struct{}

func (bareEvent) GetName() string {
	return "bare"
}

func TestAuditLogger_Handle(t *testing.T) {
	t.Parallel()

	t.Run("test handle should log the attributes of an event with log attrs", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		ctx := log.WithLogger(t.Context(), slog.New(slog.NewTextHandler(buf, nil)))

		err := handler.NewAuditLogger().Handle(ctx, dto.LoginEvent{Username: "admin", IsSuccess: true})

		require.NoError(t, err)
		out := buf.String()
		require.Contains(t, out, "security event")
		require.Contains(t, out, "event=login")
		require.Contains(t, out, "username=admin")
		require.Contains(t, out, "is_success=true")
	})

	t.Run("test handle should log an event without log attrs by name only", func(t *testing.T) {
		t.Parallel()

		buf := &bytes.Buffer{}
		ctx := log.WithLogger(t.Context(), slog.New(slog.NewTextHandler(buf, nil)))

		err := handler.NewAuditLogger().Handle(ctx, bareEvent{})

		require.NoError(t, err)
		out := buf.String()
		require.Contains(t, out, "security event")
		require.Contains(t, out, "event=bare")
	})
}
