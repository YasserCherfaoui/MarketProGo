package routes

import (
	"github.com/YasserCherfaoui/MarketProGo/handlers/product"
	"github.com/YasserCherfaoui/MarketProGo/middlewares"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ProductRoutes(router *gin.RouterGroup, db *gorm.DB) {
	productRouter := router.Group("/products")
	productHandler := product.NewProductHandler(db)

	productRouter.GET("", productHandler.GetAllProducts)
	productRouter.Use(middlewares.AuthMiddleware())
	{
		productRouter.POST("", productHandler.CreateProduct)
		productRouter.DELETE("/:id", productHandler.DeleteProduct)
	}

}
