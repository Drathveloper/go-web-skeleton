package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Drathveloper/go-web-skeleton/common/constants"
	commondomain "github.com/Drathveloper/go-web-skeleton/common/domain"
	"github.com/Drathveloper/go-web-skeleton/common/http/helper"
	commonmapper "github.com/Drathveloper/go-web-skeleton/common/http/mapper"
	"github.com/Drathveloper/go-web-skeleton/common/log"
)

// ErrSessionManagement is the only session failure the user ever sees. The error
// returned by the service carries infrastructure detail (Redis addresses, driver
// messages) and goes to the request logger instead.
var ErrSessionManagement = errors.New("your session could not be processed, please try again")

var SessionIDGeneratorFunc = helper.GenerateSessionID //nolint:gochecknoglobals

const sessionManagementErrTitle = "session management error"

type SessionService interface {
	LoadSession(ctx context.Context, sessionID string) (*commondomain.Session, error)
	CreateAnonymousSession(ctx context.Context, sessionID string) (*commondomain.Session, error)
	UpdateSession(ctx context.Context, session *commondomain.Session) error
}

type SessionConfigProvider interface {
	GetSessionTTL() time.Duration
	IsSecureCookie() bool
	GetCookieDomain() string
}

func SessionHandler(service SessionService, configProvider SessionConfigProvider) gin.HandlerFunc {
	return func(c *gin.Context) { //nolint:varnamelen
		var session *commondomain.Session
		sessionID, err := c.Cookie(constants.SessionIDKey)
		if err != nil {
			sessionID = SessionIDGeneratorFunc()
			session, err = service.CreateAnonymousSession(c.Request.Context(), sessionID)
			if err != nil {
				abortWithSessionError(c, err)
				return
			}
			c.SetSameSite(http.SameSiteStrictMode)
			c.SetCookie(
				constants.SessionIDKey,
				sessionID,
				int(configProvider.GetSessionTTL().Seconds()),
				"/",
				configProvider.GetCookieDomain(),
				configProvider.IsSecureCookie(),
				true)
		} else {
			session, err = service.LoadSession(c.Request.Context(), sessionID)
			if err != nil {
				abortWithSessionError(c, err)
				return
			}
		}
		c.Set(constants.SessionGinContextKey, session)
		c.Writer.Header().Set("Vary", "Cookie")
		c.Writer.Header().Set("Cache-Control", `no-cache="Set-Cookie"`)
		c.Next()
		// Re-read rather than reusing the captured pointer. Login rotates the
		// session (new ID, anonymous entry deleted) and logout destroys it;
		// persisting the pre-handler value here would re-save the anonymous
		// session that login just removed, and resurrect the one logout just
		// killed. A handler that cleared the session leaves nothing to write.
		current, sessionErr := helper.GetSession(c)
		if sessionErr != nil {
			return
		}
		if err = service.UpdateSession(c.Request.Context(), current); err != nil {
			abortWithSessionError(c, err)
			return
		}
	}
}

func abortWithSessionError(c *gin.Context, err error) { //nolint:varnamelen
	log.ContextLogger(c.Request.Context()).Error(
		sessionManagementErrTitle, slog.String("error", err.Error()))
	response := commonmapper.MapDomainErrorToViewResponse(
		"", http.StatusInternalServerError, "errors.title", "errors.internal")
	c.HTML(http.StatusInternalServerError, "error", response)
	c.Abort()
}
