package routes

import (
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/Drathveloper/go-web-skeleton/common/constants"
	commondomain "github.com/Drathveloper/go-web-skeleton/common/domain"
	"github.com/Drathveloper/go-web-skeleton/common/http/handler"
	"github.com/Drathveloper/go-web-skeleton/common/http/middleware"
	"github.com/Drathveloper/go-web-skeleton/common/http/static"
	"github.com/Drathveloper/go-web-skeleton/common/http/templates"
	"github.com/Drathveloper/go-web-skeleton/common/wire"
)

const initializeRoutesBaseErrMsg = "initialize routes failed"

func InitializeRoutes(container *wire.Container) (*gin.Engine, error) {
	gin.SetMode(container.Env.GinMode)
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
		// Order matters: LanguageHandler, CSRFHandler and every handler call
		// helper.MustGetSession, which panics unless SessionHandler ran first.
		webRouter.Use(middleware.SessionHandler(container.SessionService, container.Store))
		webRouter.Use(middleware.CSRFHandler())
		webRouter.Use(middleware.FlushSessionHandler())
		webRouter.Use(middleware.LanguageHandler())

		webRouter.GET("/", handler.HomeHandler())
		webRouter.POST("/language", handler.SetLanguageHandler())

		registerAuthenticationRoutes(webRouter, container)

		registerItemCategoryRoutes(webRouter, container)
		registerItemRoutes(webRouter, container)
		// scaffold:routes:register
	}
	return router, nil
}

// registerAuthenticationRoutes wires the auth screens.
//
// login and logout are deliberately public; everything under /auth/user is
// not. In the source this whole group had no Authorize at all, so an
// unauthenticated caller could POST /auth/user/new and grant itself the admin
// role — the only route group in the application missing the check.
func registerAuthenticationRoutes(router gin.IRouter, container *wire.Container) {
	auth := router.Group("/auth")
	{
		auth.GET("/login", container.AuthenticationHandler.LoginView())
		auth.POST("/login", container.AuthenticationHandler.LoginProcess())
		auth.GET("/logout", container.AuthenticationHandler.LogoutProcess())

		users := auth.Group("/user", middleware.Authorize(commondomain.AdminRole))
		{
			users.GET("", container.UserManagementHandler.ListUsersView())
			users.GET("/new", container.UserManagementHandler.CreateUserView())
			users.POST("/new", container.UserManagementHandler.CreateUserProcess())
			users.GET("/:id/edit", container.UserManagementHandler.UpdateUserView())
			users.POST("/:id/edit", container.UserManagementHandler.UpdateUserProcess())
		}
	}
}
