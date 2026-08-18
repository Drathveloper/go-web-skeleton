package middleware_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/Drathveloper/go-web-skeleton/common/http/templates"
	"github.com/Drathveloper/go-web-skeleton/common/i18n"
)

// The i18n catalog is a package-level map: loading it once here keeps the tests
// below from writing it concurrently, and lets them assert the real localized
// text of the error page instead of a stub template.
func TestMain(m *testing.M) {
	if err := i18n.InitializeI18n(); err != nil {
		fmt.Fprintf(os.Stderr, "initialize i18n failed: %v\n", err)
		os.Exit(1)
	}
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

// newEngine builds a router with the real template renderer. The middleware that
// deny a request answer with the "error" page, so a stub template would hide
// whether what reaches the user is the localized message or an error string.
func newEngine(t *testing.T) *gin.Engine {
	t.Helper()

	engine := gin.New()
	require.NoError(t, templates.InitializeTemplateRenderer(engine))

	return engine
}

func serve(t *testing.T, engine *gin.Engine, request *http.Request) *httptest.ResponseRecorder {
	t.Helper()

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)

	return recorder
}
