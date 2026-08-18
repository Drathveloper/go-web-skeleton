package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/Drathveloper/go-web-skeleton/common/constants"
	"github.com/Drathveloper/go-web-skeleton/common/domain"
	"github.com/Drathveloper/go-web-skeleton/common/http/helper"
	"github.com/Drathveloper/go-web-skeleton/common/http/middleware"
)

func TestCSRFHandler_MintsATokenOnGet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		existingToken string
	}{
		{
			name:          "test csrf handler should mint a token when the session has none",
			existingToken: "",
		},
		{
			name:          "test csrf handler should keep the token the session already has",
			existingToken: "someToken",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			session := &domain.Session{CSRFToken: tt.existingToken}
			var seenToken string

			engine := newEngine(t)
			engine.GET("/test",
				setSession(session),
				middleware.CSRFHandler(),
				func(c *gin.Context) {
					seenToken = helper.MustGetSession(c).CSRFToken
					c.Status(http.StatusOK)
				},
			)

			recorder := serve(t, engine,
				httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", nil))

			require.Equal(t, http.StatusOK, recorder.Code)
			require.NotEmpty(t, seenToken, "a form rendered without a token can never be submitted")
			if tt.existingToken != "" {
				require.Equal(t, tt.existingToken, seenToken, "the token must not rotate on every GET")
			}
		})
	}
}

func TestCSRFHandler_ValidatesUnsafeMethods(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		method           string
		sessionCSRFToken string
		formToken        string
		wantStatus       int
	}{
		{
			name:             "test csrf handler should accept a POST carrying the session token",
			method:           http.MethodPost,
			sessionCSRFToken: "someToken",
			formToken:        "someToken",
			wantStatus:       http.StatusOK,
		},
		{
			name:             "test csrf handler should reject a POST carrying a different token",
			method:           http.MethodPost,
			sessionCSRFToken: "someToken",
			formToken:        "invalidToken",
			wantStatus:       http.StatusForbidden,
		},
		{
			name:             "test csrf handler should reject a POST carrying no token",
			method:           http.MethodPost,
			sessionCSRFToken: "someToken",
			formToken:        "",
			wantStatus:       http.StatusForbidden,
		},
		// The check has to fail closed. Comparing an empty session token with an
		// equally empty form value would otherwise match, and the first unsafe
		// request of a session — made before any GET has minted a token — would
		// skip CSRF validation altogether.
		{
			name:             "test csrf handler should reject a POST when the session has no token yet",
			method:           http.MethodPost,
			sessionCSRFToken: "",
			formToken:        "",
			wantStatus:       http.StatusForbidden,
		},
		{
			name:             "test csrf handler should validate PUT requests",
			method:           http.MethodPut,
			sessionCSRFToken: "someToken",
			formToken:        "invalidToken",
			wantStatus:       http.StatusForbidden,
		},
		{
			name:             "test csrf handler should validate PATCH requests",
			method:           http.MethodPatch,
			sessionCSRFToken: "someToken",
			formToken:        "invalidToken",
			wantStatus:       http.StatusForbidden,
		},
		{
			name:             "test csrf handler should validate DELETE requests",
			method:           http.MethodDelete,
			sessionCSRFToken: "someToken",
			formToken:        "invalidToken",
			wantStatus:       http.StatusForbidden,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handlerCalled := false
			session := &domain.Session{CSRFToken: tt.sessionCSRFToken, Language: "es"}

			engine := newEngine(t)
			engine.Handle(tt.method, "/test",
				setSession(session),
				middleware.CSRFHandler(),
				func(c *gin.Context) {
					handlerCalled = true
					c.Status(http.StatusOK)
				},
			)

			body := url.Values{constants.CSRFTokenKey: {tt.formToken}}.Encode()
			request := httptest.NewRequestWithContext(t.Context(), tt.method, "/test", strings.NewReader(body))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			recorder := serve(t, engine, request)

			require.Equal(t, tt.wantStatus, recorder.Code)
			require.Equal(t, tt.wantStatus == http.StatusOK, handlerCalled,
				"a rejected request must never reach the handler")
			if tt.wantStatus == http.StatusForbidden {
				require.Contains(t, recorder.Body.String(),
					"El token de seguridad no es válido o ha caducado. Recarga la página.",
					"the rejection must be explained in the language of the session")
			}
		})
	}
}
