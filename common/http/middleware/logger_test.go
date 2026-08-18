package middleware_test

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/Drathveloper/go-web-skeleton/common/http/middleware"
	"github.com/Drathveloper/go-web-skeleton/common/log"
)

func TestLogger(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		method          string
		path            string
		wantLogContains []string
		handlerStatus   int
	}{
		{
			name:          "test logger middleware should log incoming request and outgoing response",
			method:        http.MethodGet,
			path:          "/test",
			handlerStatus: http.StatusOK,
			wantLogContains: []string{
				"incoming request",
				"outgoing response",
				"method=GET",
				"path=/test",
				"status=200",
				"request_id=",
			},
		},
		{
			name:          "test logger middleware should log POST request with error status",
			method:        http.MethodPost,
			path:          "/error",
			handlerStatus: http.StatusInternalServerError,
			wantLogContains: []string{
				"incoming request",
				"outgoing response",
				"method=POST",
				"path=/error",
				"status=500",
				"request_id=",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			logger := slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			}))

			engine := gin.New()
			engine.Handle(tt.method, tt.path,
				middleware.Logger(logger),
				func(c *gin.Context) {
					ctxLogger := log.ContextLogger(c.Request.Context())
					require.NotNil(t, ctxLogger, "logger should be available in request context")
					require.NotSame(t, slog.Default(), ctxLogger,
						"context logger should not be the default logger")

					c.Status(tt.handlerStatus)
				},
			)

			request := httptest.NewRequestWithContext(t.Context(), tt.method, tt.path, nil)
			recorder := serve(t, engine, request)

			require.Equal(t, tt.handlerStatus, recorder.Code)

			requestID := recorder.Header().Get("X-Request-Id")
			require.NotEmpty(t, requestID, "X-Request-Id header should be set")

			logOutput := buf.String()
			for _, want := range tt.wantLogContains {
				require.Contains(t, logOutput, want,
					"log output should contain %q, got: %s", want, logOutput)
			}
			require.Contains(t, logOutput, requestID,
				"log output should contain the request ID set in the header")
		})
	}
}

// Two requests must not share a request ID: correlating logs is the only reason
// the header exists.
func TestLogger_GivesEachRequestItsOwnID(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	logger := slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	engine := gin.New()
	engine.GET("/test", middleware.Logger(logger), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	first := serve(t, engine, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", nil))
	second := serve(t, engine, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", nil))

	require.NotEmpty(t, first.Header().Get("X-Request-Id"))
	require.NotEqual(t, first.Header().Get("X-Request-Id"), second.Header().Get("X-Request-Id"))
}
