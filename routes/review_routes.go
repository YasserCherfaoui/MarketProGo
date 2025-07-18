package routes

import (
	"github.com/YasserCherfaoui/MarketProGo/handlers/review"
	"github.com/YasserCherfaoui/MarketProGo/middlewares"
	"github.com/gin-gonic/gin"
)

// RegisterReviewRoutes sets up all review-related routes
func RegisterReviewRoutes(router *gin.RouterGroup, reviewHandler *review.ReviewHandler) {
	// Public routes (no authentication required)
	reviews := router.Group("/reviews")
	{
		// Get single review by ID
		reviews.GET("/:id", reviewHandler.GetReview)

		// Get reviews for a specific product variant
		reviews.GET("/product/:productVariantId", reviewHandler.GetProductReviews)
	}

	// Authenticated routes (JWT required)
	authenticatedReviews := router.Group("/reviews")
	authenticatedReviews.Use(middlewares.AuthMiddleware())
	{
		// Customer review management
		authenticatedReviews.POST("", reviewHandler.CreateReview)
		authenticatedReviews.PUT("/:id", reviewHandler.UpdateReview)
		authenticatedReviews.DELETE("/:id", reviewHandler.DeleteReview)

		// Review helpfulness
		authenticatedReviews.POST("/:id/helpful", reviewHandler.MarkReviewHelpful)

		// Image upload for reviews
		authenticatedReviews.POST("/upload-images", reviewHandler.UploadReviewImages)

		// Get reviewable products for user
		authenticatedReviews.GET("/reviewable-products", reviewHandler.GetReviewableProducts)

		// Get user's own reviews
		authenticatedReviews.GET("/user/me", reviewHandler.GetUserReviews)
	}

	// Seller routes (seller role required)
	sellerReviews := router.Group("/reviews")
	sellerReviews.Use(middlewares.SellerMiddleware())
	{
		// Seller response management
		sellerReviews.POST("/:id/response", reviewHandler.CreateSellerResponse)
		sellerReviews.PUT("/:id/response", reviewHandler.UpdateSellerResponse)
	}

	// Admin routes (admin role required)
	adminReviews := router.Group("/admin/reviews")
	adminReviews.Use(middlewares.AdminMiddleware())
	{
		// Admin review management
		adminReviews.GET("", reviewHandler.GetAllReviews)
		adminReviews.PUT("/:id/moderate", reviewHandler.ModerateReview)
		adminReviews.DELETE("/:id", reviewHandler.AdminDeleteReview)

		// Moderation statistics
		adminReviews.GET("/stats", reviewHandler.GetModerationStats)
	}

	// Seller dashboard routes
	sellerDashboard := router.Group("/seller/reviews")
	sellerDashboard.Use(middlewares.SellerMiddleware())
	{
		// Get reviews for seller's products
		sellerDashboard.GET("", reviewHandler.GetSellerReviews)
	}
}
