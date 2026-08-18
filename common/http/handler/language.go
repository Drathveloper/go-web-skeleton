package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Drathveloper/go-web-skeleton/common/http/helper"
	"github.com/Drathveloper/go-web-skeleton/common/i18n"
)

const languageFormField = "lang"

// SetLanguageHandler pins the locale for the session.
//
// Without it, IsLanguageOverridden could never become true: the field was read
// by the language middleware and persisted to Redis, but nothing ever set it,
// so a visitor's choice could not outlive the Accept-Language header of the
// current request. This is the route that makes the flag mean something.
//
// It is a POST because it mutates session state, which also puts it behind the
// CSRF middleware.
func SetLanguageHandler() gin.HandlerFunc {
	return func(c *gin.Context) { //nolint:varnamelen
		session := helper.MustGetSession(c)
		lang := c.PostForm(languageFormField)
		if !i18n.IsAvailableLanguage(lang) {
			// An unknown locale is a bad request, not a silent fallback: the
			// only source of these values is our own switcher.
			c.Status(http.StatusBadRequest)
			return
		}
		session.Language = lang
		session.IsLanguageOverridden = true

		if helper.IsHTMXRequest(c) {
			c.Header("HX-Refresh", "true")
			c.Status(http.StatusNoContent)
			return
		}
		redirectTo := c.Request.Referer()
		if redirectTo == "" {
			redirectTo = "/"
		}
		c.Redirect(http.StatusFound, redirectTo)
	}
}
