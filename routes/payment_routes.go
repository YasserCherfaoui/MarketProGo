package routes

import (
	"github.com/YasserCherfaoui/MarketProGo/handlers/payment"
	"github.com/YasserCherfaoui/MarketProGo/middlewares"
	"github.com/gin-gonic/gin"
)

// SetupPaymentRoutes sets up payment-related routes
func SetupPaymentRoutes(router *gin.Engine, paymentHandler *payment.PaymentHandler) {
	// Payment routes group
	paymentRoutes := router.Group("/api/v1/payments")
	{
		// Customer routes (require authentication)
		customerRoutes := paymentRoutes.Group("")
		customerRoutes.Use(middlewares.AuthMiddleware())
		{
			// Create a new payment
			customerRoutes.POST("", paymentHandler.InitiatePayment)

			// Get payment details
			customerRoutes.GET("/:id", paymentHandler.GetPayment)

			// Get payment status
			customerRoutes.GET("/:id/status", paymentHandler.GetPaymentStatus)

			// List user's payments
			customerRoutes.GET("", paymentHandler.ListPayments)

			// Cancel a payment
			customerRoutes.POST("/:id/cancel", paymentHandler.CancelPayment)
		}

		// Admin routes (require admin authentication)
		adminRoutes := paymentRoutes.Group("/admin")
		adminRoutes.Use(middlewares.AuthMiddleware())
		adminRoutes.Use(middlewares.AdminMiddleware())
		{
			// Process refund (admin only)
			adminRoutes.POST("/:id/refund", paymentHandler.RefundPayment)
		}

		// Webhook route (no authentication required, but signature validation)
		paymentRoutes.POST("/webhook", paymentHandler.HandleWebhook)
	}
}
