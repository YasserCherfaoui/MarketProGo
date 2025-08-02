package routes

import (
	"github.com/YasserCherfaoui/MarketProGo/handlers/wishlist"
	"github.com/YasserCherfaoui/MarketProGo/middlewares"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func WishlistRoutes(router *gin.RouterGroup, db *gorm.DB) {
	wishlistHandler := wishlist.NewWishlistHandler(db)

	wishlistGroup := router.Group("/wishlist")
	wishlistGroup.Use(middlewares.AuthMiddleware())
	{
		// Get user's wishlist
		wishlistGroup.GET("", wishlistHandler.GetWishlist)

		// Add item to wishlist
		wishlistGroup.POST("/items", wishlistHandler.AddItem)

		// Update wishlist item
		wishlistGroup.PUT("/items/:id", wishlistHandler.UpdateItem)

		// Remove item from wishlist
		wishlistGroup.DELETE("/items/:id", wishlistHandler.RemoveItem)
	}
}
