package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/Drathveloper/go-web-skeleton/common/domain"
	"github.com/Drathveloper/go-web-skeleton/common/http/middleware"
)

func TestAuthorize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		language     string
		wantMessage  string
		userRoles    []domain.Role
		allowedRoles []domain.Role
		wantStatus   int
	}{
		{
			name:         "test authorize should let through a role in the allowed list",
			userRoles:    []domain.Role{domain.AdminRole},
			allowedRoles: []domain.Role{domain.AdminRole},
			language:     "en",
			wantStatus:   http.StatusOK,
		},
		{
			name:         "test authorize should let through a session holding one of several allowed roles",
			userRoles:    []domain.Role{domain.UserRole},
			allowedRoles: []domain.Role{domain.AdminRole, domain.UserRole},
			language:     "en",
			wantStatus:   http.StatusOK,
		},
		{
			name:         "test authorize should reject a role that is not allowed",
			userRoles:    []domain.Role{domain.UserRole},
			allowedRoles: []domain.Role{domain.AdminRole},
			language:     "en",
			wantStatus:   http.StatusForbidden,
			wantMessage:  "You do not have permission to perform this action.",
		},
		{
			name:         "test authorize should reject an anonymous session with no roles",
			userRoles:    nil,
			allowedRoles: []domain.Role{domain.AdminRole},
			language:     "es",
			wantStatus:   http.StatusForbidden,
			wantMessage:  "No tienes permiso para realizar esta acción.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handlerCalled := false
			engine := newEngine(t)
			engine.GET("/test",
				setSession(&domain.Session{Roles: tt.userRoles, Language: tt.language}),
				middleware.Authorize(tt.allowedRoles...),
				func(c *gin.Context) {
					handlerCalled = true
					c.Status(http.StatusOK)
				},
			)

			recorder := serve(t, engine,
				httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", nil))

			require.Equal(t, tt.wantStatus, recorder.Code)
			require.Equal(t, tt.wantStatus == http.StatusOK, handlerCalled,
				"a denied request must never reach the handler")
			if tt.wantMessage != "" {
				require.Contains(t, recorder.Body.String(), tt.wantMessage)
			}
		})
	}
}
