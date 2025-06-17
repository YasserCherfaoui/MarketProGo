package routes

import (
	"github.com/YasserCherfaoui/MarketProGo/aw"
	"github.com/YasserCherfaoui/MarketProGo/gcs"
	"github.com/YasserCherfaoui/MarketProGo/handlers/brand"
	"github.com/YasserCherfaoui/MarketProGo/middlewares"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func BrandRoutes(r *gin.RouterGroup, db *gorm.DB, gcs *gcs.GCService, appwriteService *aw.AppwriteService) {
	brandHandler := brand.NewBrandHandler(db, gcs, appwriteService)
	brandRouter := r.Group("/brands")

	brandRouter.GET("", brandHandler.GetAllBrands)
	brandRouter.GET("/:id", brandHandler.GetBrand)
	brandRouter.Use(middlewares.AuthMiddleware())
	{
		brandRouter.POST("", brandHandler.CreateBrand)
		brandRouter.PUT("/:id", brandHandler.UpdateBrand)
		brandRouter.DELETE("/:id", brandHandler.DeleteBrand)
	}
}
