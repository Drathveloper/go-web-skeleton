package log_test

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Drathveloper/go-web-skeleton/common/constants"
	"github.com/Drathveloper/go-web-skeleton/common/log"
)

// Every handler logs through ContextLogger. It has to answer with something
// usable even when the logger middleware did not run — a nil return would turn
// an error path into a nil pointer dereference.
func TestContextLogger(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	requestLogger := slog.New(slog.NewTextHandler(buf, nil))

	tests := []struct {
		ctx      func(t *testing.T) context.Context
		wantSame *slog.Logger
		name     string
	}{
		{
			name: "test context logger should return the logger stored in the context",
			ctx: func(t *testing.T) context.Context {
				t.Helper()

				return log.WithLogger(t.Context(), requestLogger)
			},
			wantSame: requestLogger,
		},
		{
			name: "test context logger should fall back to the default logger when there is none",
			ctx: func(t *testing.T) context.Context {
				t.Helper()

				return t.Context()
			},
			wantSame: slog.Default(),
		},
		{
			name: "test context logger should fall back to the default logger when the value is not a logger",
			ctx: func(t *testing.T) context.Context {
				t.Helper()

				// A foreign value stored under our own key: the type assertion
				// in ContextLogger is what has to catch it.
				return context.WithValue(t.Context(), constants.LoggerContextKey, "not a logger")
			},
			wantSame: slog.Default(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Same(t, tt.wantSame, log.ContextLogger(tt.ctx(t)))
		})
	}
}
