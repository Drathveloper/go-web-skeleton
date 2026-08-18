package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/Drathveloper/go-web-skeleton/common/constants"
	"github.com/Drathveloper/go-web-skeleton/common/domain"
	"github.com/Drathveloper/go-web-skeleton/common/http/helper"
	"github.com/Drathveloper/go-web-skeleton/common/http/middleware"
)

// Flash messages are cleared after the page that displayed them, and only then.
// A POST usually ends in a redirect, and its alert is meant for the GET that
// follows: clearing it there would drop the message on the floor.
func TestFlushSessionHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		method      string
		wantFlushed bool
	}{
		{
			name:        "test flush session handler should clear the alerts after a GET",
			method:      http.MethodGet,
			wantFlushed: true,
		},
		{
			name:        "test flush session handler should keep the alerts after a POST",
			method:      http.MethodPost,
			wantFlushed: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			session := &domain.Session{
				AlertMessages: domain.AlertMessages{
					domain.NewErrorAlertMessage(http.StatusForbidden, "someTitle", "someMsg"),
				},
			}

			var seenByHandler domain.AlertMessages
			engine := newEngine(t)
			engine.Handle(tt.method, "/test",
				setSession(session),
				middleware.FlushSessionHandler(),
				func(c *gin.Context) {
					seenByHandler = helper.MustGetSession(c).AlertMessages
					c.Status(http.StatusOK)
				},
			)

			request := httptest.NewRequestWithContext(t.Context(), tt.method, "/test", strings.NewReader(""))
			recorder := serve(t, engine, request)

			require.Equal(t, http.StatusOK, recorder.Code)
			require.Len(t, seenByHandler, 1, "the handler must still be able to render the alert")
			if tt.wantFlushed {
				require.Empty(t, session.AlertMessages, "an alert shown once must not be shown again")
			} else {
				require.Len(t, session.AlertMessages, 1, "the alert still has to survive the redirect")
			}
		})
	}
}

// The middleware runs on routes that may abort before a session exists. It must
// not take the process down when it finds none.
func TestFlushSessionHandler_WithoutASession(t *testing.T) {
	t.Parallel()

	engine := newEngine(t)
	engine.GET("/test",
		func(c *gin.Context) { c.Set(constants.SessionGinContextKey, nil) },
		middleware.FlushSessionHandler(),
		func(c *gin.Context) { c.Status(http.StatusOK) },
	)

	require.NotPanics(t, func() {
		recorder := serve(t, engine, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", nil))

		require.Equal(t, http.StatusOK, recorder.Code)
	})
}
