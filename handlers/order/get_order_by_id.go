package order

import (
	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetOrderByID - Admin endpoint to get any order by ID
func (h *OrderHandler) GetOrderByID(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		response.GenerateBadRequestResponse(c, "order/get_order_by_id", "Order ID is required")
		return
	}

	var order models.Order
	if err := h.db.
		Preload("User").
		Preload("Company").
		Preload("ShippingAddress").
		Preload("Items.ProductVariant.Product").
		Preload("Items.ProductVariant.Product.Images").
		Preload("Items.ProductVariant.OptionValues").
		Preload("Items.Product"). // Legacy support
		Preload("Items.InventoryItem").
		First(&order, orderID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.GenerateNotFoundResponse(c, "order/get_order_by_id", "Order not found")
		} else {
			response.GenerateInternalServerErrorResponse(c, "order/get_order_by_id", "Failed to get order")
		}
		return
	}

	response.GenerateSuccessResponse(c, "Order retrieved successfully", order)
}
