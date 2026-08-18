package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/Drathveloper/go-web-skeleton/common/domain"
	"github.com/Drathveloper/go-web-skeleton/common/http/middleware"
	"github.com/Drathveloper/go-web-skeleton/common/wire"
)

func registerItemRoutes(router gin.IRouter, container *wire.Container) {
	items := router.Group("/item")
	{
		items.Use(middleware.Authorize(domain.AdminRole))
		items.GET("", container.ItemHandler.ListItemsView())

		items.GET("/new", container.ItemHandler.CreateItemView())
		items.POST("/new", container.ItemHandler.CreateItemProcess())

		items.GET("/:id/edit", container.ItemHandler.UpdateItemView())
		items.POST("/:id/edit", container.ItemHandler.UpdateItemProcess())

		items.POST("/:id/delete", container.ItemHandler.DeleteItemProcess())
	}
}
