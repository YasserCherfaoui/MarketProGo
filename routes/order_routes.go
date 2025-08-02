package routes

import (
	"github.com/YasserCherfaoui/MarketProGo/handlers/order"
	"github.com/YasserCherfaoui/MarketProGo/middlewares"
	"github.com/gin-gonic/gin"
)

func OrderRoutes(router *gin.RouterGroup, orderHandler *order.OrderHandler) {
	// Customer order routes (require authentication)
	orderRouter := router.Group("/orders")
	orderRouter.Use(middlewares.AuthMiddleware())
	{
		orderRouter.POST("/place", orderHandler.PlaceOrder)
		orderRouter.GET("", orderHandler.GetOrders)
		orderRouter.GET("/:id", orderHandler.GetOrder)
		orderRouter.PUT("/:id/cancel", orderHandler.CancelOrder)
	}

	// Admin order routes (require admin authentication)
	adminOrderRouter := router.Group("/admin/orders")
	adminOrderRouter.Use(middlewares.AuthMiddleware()) // TODO: Add admin middleware when available
	{
		// Order management
		adminOrderRouter.GET("", orderHandler.GetAllOrders)
		adminOrderRouter.GET("/stats", orderHandler.GetOrderStats)
		adminOrderRouter.GET("/:id", orderHandler.GetOrderByID)

		// Order status management
		adminOrderRouter.PUT("/:id/status", orderHandler.UpdateOrderStatus)
		adminOrderRouter.PUT("/:id/payment", orderHandler.UpdatePaymentStatus)
	}

	// Admin invoice routes
	adminInvoiceRouter := router.Group("/admin/invoices")
	adminInvoiceRouter.Use(middlewares.AuthMiddleware()) // TODO: Add admin middleware when available
	{
		adminInvoiceRouter.POST("", orderHandler.CreateInvoice)
		adminInvoiceRouter.GET("", orderHandler.GetInvoices)
		adminInvoiceRouter.GET("/:id", orderHandler.GetInvoice)
		adminInvoiceRouter.PUT("/:id", orderHandler.UpdateInvoice)
	}
}
