package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Drathveloper/go-web-skeleton/common/http/helper"
	"github.com/Drathveloper/go-web-skeleton/common/http/mapper"
)

const homeTemplate = "home/home"

// HomeHandler renders the landing page behind "/".
//
// The sidebar has always linked to "/", so without this route every screen in
// the application carried a dead link back to a 404.
func HomeHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := helper.MustGetSession(c)
		response := mapper.MapDataToViewResponse[struct{}](nil, nil, session)
		c.HTML(http.StatusOK, homeTemplate, response)
	}
}
