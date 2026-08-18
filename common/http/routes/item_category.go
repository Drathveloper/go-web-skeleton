package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/Drathveloper/go-web-skeleton/common/domain"
	"github.com/Drathveloper/go-web-skeleton/common/http/middleware"
	"github.com/Drathveloper/go-web-skeleton/common/wire"
)

func registerItemCategoryRoutes(router gin.IRouter, container *wire.Container) {
	itemCategories := router.Group("/item-category")
	{
		itemCategories.Use(middleware.Authorize(domain.AdminRole))
		itemCategories.GET("", container.ItemCategoryHandler.ListItemCategoriesView())

		itemCategories.GET("/new", container.ItemCategoryHandler.CreateItemCategoryView())
		itemCategories.POST("/new", container.ItemCategoryHandler.CreateItemCategoryProcess())

		itemCategories.GET("/:id/edit", container.ItemCategoryHandler.UpdateItemCategoryView())
		itemCategories.POST("/:id/edit", container.ItemCategoryHandler.UpdateItemCategoryProcess())

		itemCategories.POST("/:id/delete", container.ItemCategoryHandler.DeleteItemCategoryProcess())
	}
}
