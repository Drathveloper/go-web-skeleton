package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/Drathveloper/go-web-skeleton/common/http/helper"
	"github.com/Drathveloper/go-web-skeleton/common/i18n"
)

func LanguageHandler() gin.HandlerFunc {
	return func(c *gin.Context) { //nolint:varnamelen
		session := helper.MustGetSession(c)
		if session.Language != "" {
			c.Next()
			return
		}
		lang := c.GetHeader("Accept-Language")
		parsedLang := i18n.ParseAcceptLanguage(lang)
		session.Language = parsedLang
		c.Next()
		if !session.IsLanguageOverridden {
			session.Language = ""
		}
	}
}
