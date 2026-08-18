package helper_test

import (
	"bytes"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/Drathveloper/go-web-skeleton/common/domain"
	"github.com/Drathveloper/go-web-skeleton/common/http/helper"
	"github.com/Drathveloper/go-web-skeleton/common/http/templates"
	"github.com/Drathveloper/go-web-skeleton/common/log"
)

// errSensitive stands for the kind of error the infrastructure really produces:
// it names a host, a port and a driver. None of it may reach the browser.
var errSensitive = errors.New(`dial tcp 10.0.3.14:6379: connect: connection refused (user="skeleton" db="skeleton")`)

func TestLogError(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	logger := slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequestWithContext(
		log.WithLogger(t.Context(), logger), http.MethodGet, "/test", nil)

	helper.LogError(context, "load session failed", errSensitive)

	require.Contains(t, buf.String(), "load session failed")
	require.Contains(t, buf.String(), "10.0.3.14:6379", "the detail belongs in the log, in full")
}

func TestFlashError(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	logger := slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequestWithContext(
		log.WithLogger(t.Context(), logger), http.MethodGet, "/test", nil)
	session := &domain.Session{Language: "es"}

	helper.FlashError(context, session, http.StatusInternalServerError,
		"load session failed", helper.UnexpectedErrorMessageID, errSensitive)

	require.Len(t, session.AlertMessages, 1)
	require.Equal(t, domain.AlertMessage{
		Code:    http.StatusInternalServerError,
		Title:   "Error",
		Message: "Se ha producido un error inesperado. Inténtalo de nuevo más tarde.",
		Type:    "error",
	}, session.AlertMessages[0])
	require.Contains(t, buf.String(), "10.0.3.14:6379")
}

// The rendered page is the real assertion: the alert on screen must be the
// localized catalog text, the status must be the real one, and nothing from the
// underlying error may appear in the HTML.
func TestRenderErrorPage(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	require.NoError(t, templates.InitializeTemplateRenderer(engine))

	var renderErrors []string
	engine.GET("/test", func(c *gin.Context) {
		session := &domain.Session{Language: "es"}
		helper.RenderErrorPage(c, session, http.StatusInternalServerError,
			"load session failed", helper.UnexpectedErrorMessageID, errSensitive)
		for _, ginErr := range c.Errors {
			renderErrors = append(renderErrors, ginErr.Error())
		}
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", nil)
	engine.ServeHTTP(recorder, request)

	require.Empty(t, renderErrors)
	require.Equal(t, http.StatusInternalServerError, recorder.Code,
		"a failure rendered as 200 is cached by browsers and counted as a success by monitoring")

	body := recorder.Body.String()
	require.Contains(t, body, "Se ha producido un error inesperado.")
	require.Contains(t, body, `<html lang="es">`)
	require.NotContains(t, body, "10.0.3.14", "the error page must not leak infrastructure detail")
	require.NotContains(t, body, "connection refused")
	require.NotContains(t, body, helper.UnexpectedErrorMessageID, "the message identifier is not user-facing text")
}

func TestLocalizedMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		lang      string
		messageID string
		expected  string
	}{
		{
			name:      "test localized message resolves in the session language",
			lang:      "es",
			messageID: helper.InvalidFormDataMessageID,
			expected:  "Datos del formulario no válidos.",
		},
		{
			name:      "test localized message resolves in english",
			lang:      "en",
			messageID: helper.NotFoundErrorMessageID,
			expected:  "The requested resource does not exist.",
		},
		{
			name:      "test localized message falls back to the identifier for an unknown locale",
			lang:      "zz",
			messageID: helper.ErrorTitleMessageID,
			expected:  helper.ErrorTitleMessageID,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual := helper.LocalizedMessage(&domain.Session{Language: tt.lang}, tt.messageID)

			require.Equal(t, tt.expected, actual)
		})
	}
}
