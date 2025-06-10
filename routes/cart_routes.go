package routes

import (
	"github.com/YasserCherfaoui/MarketProGo/handlers/cart"
	"github.com/YasserCherfaoui/MarketProGo/middlewares"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CartRoutes(router *gin.RouterGroup, db *gorm.DB) {
	cartHandler := cart.NewCartHandler(db)

	cartRouter := router.Group("/cart")
	cartRouter.Use(middlewares.AuthMiddleware())
	{
		cartRouter.GET("", cartHandler.GetCart)
		cartRouter.POST("/items", cartHandler.AddItem)
		cartRouter.PUT("/items/:id", cartHandler.UpdateItem)
		cartRouter.DELETE("/items/:id", cartHandler.DeleteItem)
	}
}
