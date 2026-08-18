package handler_test

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
	"github.com/Drathveloper/go-web-skeleton/common/http/handler"
	"github.com/Drathveloper/go-web-skeleton/common/http/helper"
	"github.com/Drathveloper/go-web-skeleton/common/i18n"
)

// Deliberately not parallel: it loads the i18n catalog, which is a package-level
// map that TestHomeHandler also writes. Sequential tests all run before the
// parallel ones resume, so this is the one arrangement with no data race.
//
// SetLanguageHandler is what makes IsLanguageOverridden mean anything: without a
// route that sets it, a visitor's choice of language could never outlive the
// Accept-Language header of the request that made it.
func TestSetLanguageHandler(t *testing.T) {
	require.NoError(t, i18n.InitializeI18n())
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name            string
		lang            string
		referer         string
		htmxHeader      string
		wantLocation    string
		wantLanguage    string
		wantStatus      int
		wantOverridden  bool
		wantRefreshHTMX bool
	}{
		{
			name:           "test set language should pin the language and go back to the current page",
			lang:           "es",
			referer:        "/item",
			wantStatus:     http.StatusFound,
			wantLocation:   "/item",
			wantLanguage:   "es",
			wantOverridden: true,
		},
		{
			name:           "test set language should fall back to home when there is no referer",
			lang:           "en",
			wantStatus:     http.StatusFound,
			wantLocation:   "/",
			wantLanguage:   "en",
			wantOverridden: true,
		},
		{
			name:            "test set language should ask htmx to refresh instead of redirecting",
			lang:            "es",
			htmxHeader:      "true",
			wantStatus:      http.StatusNoContent,
			wantLanguage:    "es",
			wantOverridden:  true,
			wantRefreshHTMX: true,
		},
		{
			name:           "test set language should reject a language with no catalog",
			lang:           "zz",
			wantStatus:     http.StatusBadRequest,
			wantLanguage:   "en",
			wantOverridden: false,
		},
		{
			name:           "test set language should reject an empty language",
			lang:           "",
			wantStatus:     http.StatusBadRequest,
			wantLanguage:   "en",
			wantOverridden: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := &domain.Session{Language: "en"}

			engine := gin.New()
			engine.POST("/language",
				func(c *gin.Context) { c.Set(constants.SessionGinContextKey, session) },
				handler.SetLanguageHandler(),
			)

			body := url.Values{"lang": {tt.lang}}.Encode()
			request := httptest.NewRequestWithContext(
				t.Context(), http.MethodPost, "/language", strings.NewReader(body))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			if tt.referer != "" {
				request.Header.Set("Referer", tt.referer)
			}
			if tt.htmxHeader != "" {
				request.Header.Set(helper.HXRequestHeader, tt.htmxHeader)
			}

			recorder := httptest.NewRecorder()
			engine.ServeHTTP(recorder, request)

			require.Equal(t, tt.wantStatus, recorder.Code)
			require.Equal(t, tt.wantLanguage, session.Language)
			require.Equal(t, tt.wantOverridden, session.IsLanguageOverridden)
			if tt.wantLocation != "" {
				require.Equal(t, tt.wantLocation, recorder.Header().Get("Location"))
			}
			if tt.wantRefreshHTMX {
				require.Equal(t, "true", recorder.Header().Get("Hx-Refresh"))
			}
		})
	}
}
