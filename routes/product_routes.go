package routes

import (
	"github.com/YasserCherfaoui/MarketProGo/aw"
	"github.com/YasserCherfaoui/MarketProGo/gcs"
	"github.com/YasserCherfaoui/MarketProGo/handlers/product"
	"github.com/YasserCherfaoui/MarketProGo/middlewares"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ProductRoutes(router *gin.RouterGroup, db *gorm.DB, gcsService *gcs.GCService, appwriteService *aw.AppwriteService) {
	productRouter := router.Group("/products")
	productHandler := product.NewProductHandler(db, gcsService, appwriteService)

	productRouter.GET("", productHandler.GetAllProducts)
	productRouter.GET("/:id", productHandler.GetProduct)

	// Product variants endpoint - requires authentication for stock management
	productVariantRouter := router.Group("/product-variants")
	productVariantRouter.Use(middlewares.AuthMiddleware())
	{
		productVariantRouter.GET("", productHandler.GetProductVariants)
	}

	productRouter.Use(middlewares.AuthMiddleware())
	{
		productRouter.POST("", productHandler.CreateProduct)
		productRouter.PUT("/:id", productHandler.UpdateProduct)
		productRouter.DELETE("/:id", productHandler.DeleteProduct)
	}

}
