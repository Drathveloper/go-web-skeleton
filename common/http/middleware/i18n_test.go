package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/Drathveloper/go-web-skeleton/common/constants"
	"github.com/Drathveloper/go-web-skeleton/common/domain"
	"github.com/Drathveloper/go-web-skeleton/common/http/middleware"
)

// The language of a request has two lives: the one the handler renders with, and
// the one that survives into the session. Asserting only the second — as the
// version this was ported from did — cannot tell "resolved to English" from
// "never resolved anything", because both leave the session empty.
func TestLanguageHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		session          *domain.Session
		name             string
		acceptLanguage   string
		wantDuringVisit  string
		wantAfterRequest string
	}{
		{
			name:             "test language handler should keep the language already in the session",
			session:          &domain.Session{Language: "es"},
			acceptLanguage:   "en-US,en;q=0.9",
			wantDuringVisit:  "es",
			wantAfterRequest: "es",
		},
		{
			name:             "test language handler should resolve an available accept language header",
			session:          &domain.Session{},
			acceptLanguage:   "es-ES,es;q=0.9,en;q=0.8",
			wantDuringVisit:  "es",
			wantAfterRequest: "",
		},
		{
			name:             "test language handler should fall back to the default for an empty header",
			session:          &domain.Session{},
			acceptLanguage:   "",
			wantDuringVisit:  "en",
			wantAfterRequest: "",
		},
		{
			name:             "test language handler should fall back to the default for an unavailable language",
			session:          &domain.Session{},
			acceptLanguage:   "fr-FR,de;q=0.8",
			wantDuringVisit:  "en",
			wantAfterRequest: "",
		},
		{
			name:             "test language handler should fall back to the default for a malformed header",
			session:          &domain.Session{},
			acceptLanguage:   "not a language tag;;;",
			wantDuringVisit:  "en",
			wantAfterRequest: "",
		},
		{
			name:             "test language handler should keep an overridden language",
			session:          &domain.Session{Language: "es", IsLanguageOverridden: true},
			acceptLanguage:   "en-US,en;q=0.9",
			wantDuringVisit:  "es",
			wantAfterRequest: "es",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var duringVisit string
			engine := newEngine(t)
			engine.GET("/test",
				setSession(tt.session),
				middleware.LanguageHandler(),
				func(c *gin.Context) {
					duringVisit = tt.session.Language
					c.Status(http.StatusOK)
				},
			)

			request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", nil)
			request.Header.Set("Accept-Language", tt.acceptLanguage)
			recorder := serve(t, engine, request)

			require.Equal(t, http.StatusOK, recorder.Code)
			require.Equal(t, tt.wantDuringVisit, duringVisit, "language the handler rendered with")
			require.Equal(t, tt.wantAfterRequest, tt.session.Language, "language left in the session")
		})
	}
}

// The regression the IsLanguageOverridden flag exists for: a language the user
// picked has to outlive the request, or the next one resolves the header again
// and the choice is gone.
func TestLanguageHandler_KeepsAChoiceMadeDuringTheRequest(t *testing.T) {
	t.Parallel()

	session := &domain.Session{}
	engine := newEngine(t)
	engine.GET("/test",
		setSession(session),
		middleware.LanguageHandler(),
		func(c *gin.Context) {
			session.Language = "es"
			session.IsLanguageOverridden = true
			c.Status(http.StatusOK)
		},
	)

	request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", nil)
	request.Header.Set("Accept-Language", "en-US,en;q=0.9")
	recorder := serve(t, engine, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, "es", session.Language)
}

func setSession(session *domain.Session) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(constants.SessionGinContextKey, session)
	}
}
