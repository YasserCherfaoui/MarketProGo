package routes

import (
	"github.com/YasserCherfaoui/MarketProGo/gcs"
	"github.com/YasserCherfaoui/MarketProGo/handlers/product"
	"github.com/YasserCherfaoui/MarketProGo/middlewares"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ProductRoutes(router *gin.RouterGroup, db *gorm.DB, gcsService *gcs.GCService) {
	productRouter := router.Group("/products")
	productHandler := product.NewProductHandler(db, gcsService)

	productRouter.GET("", productHandler.GetAllProducts)
	productRouter.GET("/:id", productHandler.GetProduct)
	productRouter.Use(middlewares.AuthMiddleware())
	{
		productRouter.POST("", productHandler.CreateProduct)
		productRouter.PUT("/:id", productHandler.UpdateProduct)
		productRouter.DELETE("/:id", productHandler.DeleteProduct)
	}

}
