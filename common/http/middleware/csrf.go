package middleware

import (
	"crypto/subtle"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Drathveloper/go-web-skeleton/common/constants"
	"github.com/Drathveloper/go-web-skeleton/common/http/helper"
	"github.com/Drathveloper/go-web-skeleton/common/http/mapper"
	"github.com/Drathveloper/go-web-skeleton/common/log"
)

var ErrCSRFValidationFailed = errors.New("CSRF validation failed")

func CSRFHandler() gin.HandlerFunc {
	return func(c *gin.Context) { //nolint:varnamelen
		session := helper.MustGetSession(c)
		if c.Request.Method == http.MethodGet {
			if session.CSRFToken == "" {
				session.CSRFToken = helper.GenerateSessionID()
			}
			c.Next()
			return
		}
		if c.Request.Method == http.MethodPost ||
			c.Request.Method == http.MethodPut ||
			c.Request.Method == http.MethodDelete ||
			c.Request.Method == http.MethodPatch {
			if !isValidCSRFToken(session.CSRFToken, c.Request.FormValue(constants.CSRFTokenKey)) {
				log.ContextLogger(c.Request.Context()).Warn(
					"csrf validation failed",
					slog.String("error", ErrCSRFValidationFailed.Error()),
					slog.String("path", c.Request.URL.Path))
				response := mapper.MapDomainErrorToViewResponse(
					session.Language, http.StatusForbidden, "errors.title", "session.invalid_csrf")
				c.HTML(http.StatusForbidden, "error", response)
				c.Abort()
				return
			}
		}
	}
}

// isValidCSRFToken compares the token minted for the session with the one submitted
// in the form, in constant time. An empty session token never validates: otherwise
// the first unsafe request of a session, made before any GET has minted a token,
// would match an equally empty form value and skip the check altogether.
func isValidCSRFToken(sessionToken, requestToken string) bool {
	if sessionToken == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(sessionToken), []byte(requestToken)) == 1
}
