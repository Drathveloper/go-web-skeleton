package middleware

import (
	"errors"
	"log/slog"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"

	"github.com/Drathveloper/go-web-skeleton/common/domain"
	"github.com/Drathveloper/go-web-skeleton/common/http/helper"
	commonmapper "github.com/Drathveloper/go-web-skeleton/common/http/mapper"
	"github.com/Drathveloper/go-web-skeleton/common/log"
)

var ErrRoleNotAllowedToAccessResource = errors.New("role not allowed to access to requested resource")

func Authorize(allowedRoles ...domain.Role) gin.HandlerFunc {
	return func(c *gin.Context) { //nolint:varnamelen
		session := helper.MustGetSession(c)
		if !containsRole(allowedRoles, session.Roles) {
			log.ContextLogger(c.Request.Context()).Warn(
				"authorization denied",
				slog.String("error", ErrRoleNotAllowedToAccessResource.Error()),
				slog.String("path", c.Request.URL.Path))
			response := commonmapper.MapDomainErrorToViewResponse(
				session.Language, http.StatusForbidden, "errors.title", "errors.forbidden")
			c.HTML(http.StatusForbidden, "error", response)
			c.Abort()
			return
		}
		c.Next()
	}
}

func containsRole(allowedRoles []domain.Role, sessionRoles []domain.Role) bool {
	for _, sessionRole := range sessionRoles {
		if slices.Contains(allowedRoles, sessionRole) {
			return true
		}
	}
	return false
}
