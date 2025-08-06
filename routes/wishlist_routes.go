package routes

import (
	"github.com/YasserCherfaoui/MarketProGo/handlers/wishlist"
	"github.com/YasserCherfaoui/MarketProGo/middlewares"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func WishlistRoutes(router *gin.RouterGroup, db *gorm.DB) {
	wishlistHandler := wishlist.NewWishlistHandler(db)

	// Customer wishlist routes (require authentication)
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

	// Admin wishlist routes (require admin authentication)
	adminWishlistGroup := router.Group("/admin/wishlists")
	adminWishlistGroup.Use(middlewares.AuthMiddleware())
	adminWishlistGroup.Use(middlewares.AdminMiddleware())
	{
		// Get all wishlists with pagination and filtering
		adminWishlistGroup.GET("", wishlistHandler.GetAllWishlists)

		// Get wishlist statistics
		adminWishlistGroup.GET("/stats", wishlistHandler.GetWishlistStats)

		// Get specific wishlist by ID
		adminWishlistGroup.GET("/:id", wishlistHandler.GetWishlistByID)

		// Get specific user's wishlist
		adminWishlistGroup.GET("/user/:user_id", wishlistHandler.GetUserWishlist)

		// Delete wishlist item (admin override)
		adminWishlistGroup.DELETE("/items/:id", wishlistHandler.DeleteWishlistItem)
	}
}
