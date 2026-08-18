package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Drathveloper/go-web-skeleton/common/http/helper"
)

func FlushSessionHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if c.Request.Method == http.MethodGet {
			session, err := helper.GetSession(c)
			if err != nil {
				return
			}
			session.FlushAlertMessages()
			return
		}
	}
}
