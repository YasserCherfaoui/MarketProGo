package routes

import (
	"github.com/YasserCherfaoui/MarketProGo/handlers/category"
	"github.com/YasserCherfaoui/MarketProGo/middlewares"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CategoryRoutes(r *gin.RouterGroup, db *gorm.DB) {
	categoryHandler := category.NewCategoryHandler(db)
	categoryRouter := r.Group("/categories")

	categoryRouter.GET("", categoryHandler.GetAllCategories)
	categoryRouter.GET("/:id", categoryHandler.GetCategory)
	categoryRouter.Use(middlewares.AuthMiddleware())
	{
		categoryRouter.POST("", categoryHandler.CreateCategory)
		categoryRouter.DELETE("/:id", categoryHandler.DeleteCategory)
	}

}
