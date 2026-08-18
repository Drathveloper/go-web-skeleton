package routes

import (
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/Drathveloper/go-web-skeleton/common/config"
	"github.com/Drathveloper/go-web-skeleton/common/constants"
	"github.com/Drathveloper/go-web-skeleton/common/http/handler"
	"github.com/Drathveloper/go-web-skeleton/common/http/middleware"
	"github.com/Drathveloper/go-web-skeleton/common/http/static"
	"github.com/Drathveloper/go-web-skeleton/common/http/templates"
	"github.com/Drathveloper/go-web-skeleton/common/wire"
)

const initializeRoutesBaseErrMsg = "initialize routes failed"

func InitializeRoutes(container *wire.Container) (*gin.Engine, error) {
	gin.SetMode(config.Env.GinMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logger(slog.Default()))

	if err := templates.InitializeTemplateRenderer(router); err != nil {
		return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, initializeRoutesBaseErrMsg, err)
	}
	if err := static.InitializeStaticFiles(router); err != nil {
		return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, initializeRoutesBaseErrMsg, err)
	}

	// Unauthenticated infrastructure endpoint: no session, no CSRF, no locale.
	router.GET("/health", handler.HealthHandler())

	webRouter := router.Group("")
	{
		webRouter.Use(middleware.LanguageHandler())

		// scaffold:routes:register
		_ = container
	}
	return router, nil
}
