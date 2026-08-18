package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/Drathveloper/go-web-skeleton/common/constants"
	"github.com/Drathveloper/go-web-skeleton/common/domain"
	"github.com/Drathveloper/go-web-skeleton/common/http/handler"
	"github.com/Drathveloper/go-web-skeleton/common/http/templates"
	"github.com/Drathveloper/go-web-skeleton/common/i18n"
)

// The home page is the one screen the sidebar has always linked to, so it is
// exercised end to end: real session, real renderer, real route.
func TestHomeHandler(t *testing.T) {
	t.Parallel()
	require.NoError(t, i18n.InitializeI18n())
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	require.NoError(t, templates.InitializeTemplateRenderer(engine))
	engine.Use(func(c *gin.Context) {
		c.Set(constants.SessionGinContextKey, &domain.Session{
			ID: "session-id", Username: "admin", Language: "es", CSRFToken: "csrf", UserID: 1,
		})
		c.Next()
	})
	engine.GET("/", handler.HomeHandler())

	recorder := httptest.NewRecorder()
	request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	engine.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	body := recorder.Body.String()
	require.Contains(t, body, `<html lang="es">`)
	require.Contains(t, body, "<title>Inicio</title>")
	require.Contains(t, body, "Gestionar usuarios")
	require.Contains(t, body, `href="/auth/user"`)
	require.NotContains(t, body, "home.cards")
}
