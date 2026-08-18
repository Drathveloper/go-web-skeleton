package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Drathveloper/go-web-skeleton/common/http/dto"
)

const okStatus = "ok"

func HealthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, dto.HealthResponse{Status: okStatus})
	}
}
