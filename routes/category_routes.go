package routes

import (
	"github.com/YasserCherfaoui/MarketProGo/aw"
	"github.com/YasserCherfaoui/MarketProGo/gcs"
	"github.com/YasserCherfaoui/MarketProGo/handlers/category"
	"github.com/YasserCherfaoui/MarketProGo/middlewares"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CategoryRoutes(r *gin.RouterGroup, db *gorm.DB, gcs *gcs.GCService, appwriteService *aw.AppwriteService) {
	categoryHandler := category.NewCategoryHandler(db, gcs, appwriteService)
	categoryRouter := r.Group("/categories")

	categoryRouter.GET("", categoryHandler.GetAllCategories)
	categoryRouter.GET("/:id", categoryHandler.GetCategory)
	categoryRouter.Use(middlewares.AuthMiddleware())
	{
		categoryRouter.POST("", categoryHandler.CreateCategory)
		categoryRouter.PUT("/:id", categoryHandler.UpdateCategory)
		categoryRouter.DELETE("/:id", categoryHandler.DeleteCategory)
	}

}
