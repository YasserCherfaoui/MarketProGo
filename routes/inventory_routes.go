package routes

import (
	"github.com/YasserCherfaoui/MarketProGo/handlers/inventory"
	"github.com/YasserCherfaoui/MarketProGo/middlewares"
	"github.com/gin-gonic/gin"
)

func InventoryRoutes(r *gin.RouterGroup, inventoryHandler *inventory.InventoryHandler) {
	// All inventory routes require authentication and admin privileges
	inventoryGroup := r.Group("/inventory")
	inventoryGroup.Use(middlewares.AuthMiddleware())
	// inventoryGroup.Use(middlewares.AdminMiddleware()) // Uncomment when admin middleware is implemented

	// Dashboard route - comprehensive overview for admin
	inventoryGroup.GET("/dashboard", inventoryHandler.GetInventoryDashboard)

	// Warehouse management routes
	warehouseGroup := inventoryGroup.Group("/warehouses")
	{
		warehouseGroup.POST("", inventoryHandler.CreateWarehouse)
		warehouseGroup.GET("", inventoryHandler.GetAllWarehouses)
		warehouseGroup.GET("/:id", inventoryHandler.GetWarehouse)
		warehouseGroup.PUT("/:id", inventoryHandler.UpdateWarehouse)
		warehouseGroup.DELETE("/:id", inventoryHandler.DeleteWarehouse)
	}

	// Product inventory overview route
	inventoryGroup.GET("/products", inventoryHandler.GetProductInventoryOverview)

	// Stock management routes
	stockGroup := inventoryGroup.Group("/stock")
	{
		stockGroup.GET("", inventoryHandler.GetStockLevels)
		stockGroup.POST("/adjust", inventoryHandler.AdjustStock)
		stockGroup.GET("/by-product/:product_variant_id", inventoryHandler.GetMultiWarehouseStock)
		// stockGroup.POST("/bulk-adjust", inventoryHandler.BulkAdjustStock)
		// stockGroup.POST("/transfer", inventoryHandler.TransferStock)
		// stockGroup.POST("/reserve", inventoryHandler.ReserveStock)
		// stockGroup.DELETE("/reserve/:id", inventoryHandler.ReleaseReservation)
	}

	// Batch tracking route
	inventoryGroup.GET("/batches", inventoryHandler.GetInventoryBatches)

	// Stock movement and audit routes
	movementGroup := inventoryGroup.Group("/movements")
	{
		movementGroup.GET("", inventoryHandler.GetStockMovements)
		movementGroup.GET("/:id", inventoryHandler.GetStockMovement)
	}

	// Alerts and notifications routes
	alertsGroup := inventoryGroup.Group("/alerts")
	{
		alertsGroup.GET("", inventoryHandler.GetStockAlerts)
		// alertsGroup.POST("", inventoryHandler.CreateStockAlert)
		// alertsGroup.PUT("/:id", inventoryHandler.UpdateStockAlert)
		// alertsGroup.DELETE("/:id", inventoryHandler.DeleteStockAlert)
	}

	// Reports and analytics routes (keeping commented for future implementation)
	// reportsGroup := inventoryGroup.Group("/reports")
	// {
	// 	reportsGroup.GET("/stock-levels", inventoryHandler.GetStockLevelReport)
	// 	reportsGroup.GET("/low-stock", inventoryHandler.GetLowStockReport)
	// 	reportsGroup.GET("/movements", inventoryHandler.GetMovementReport)
	// 	reportsGroup.GET("/valuation", inventoryHandler.GetInventoryValuation)
	// }
}
