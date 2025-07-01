package order

import (
	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *OrderHandler) CancelOrder(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.GenerateUnauthorizedResponse(c, "order/cancel_order", "User not authenticated")
		return
	}
	uid := userID.(uint)

	orderID := c.Param("id")
	if orderID == "" {
		response.GenerateBadRequestResponse(c, "order/cancel_order", "Order ID is required")
		return
	}

	// Start transaction
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get the order
	var order models.Order
	if err := tx.Where("id = ? AND user_id = ?", orderID, uid).First(&order).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			response.GenerateNotFoundResponse(c, "order/cancel_order", "Order not found")
		} else {
			response.GenerateInternalServerErrorResponse(c, "order/cancel_order", "Failed to get order")
		}
		return
	}

	// Check if order can be cancelled
	if order.Status != models.OrderStatusPending {
		tx.Rollback()
		response.GenerateBadRequestResponse(c, "order/cancel_order", "Order cannot be cancelled. Only pending orders can be cancelled")
		return
	}

	// Update order status
	order.Status = models.OrderStatusCancelled
	if err := tx.Save(&order).Error; err != nil {
		tx.Rollback()
		response.GenerateInternalServerErrorResponse(c, "order/cancel_order", "Failed to cancel order")
		return
	}

	// Update order items status
	if err := tx.Model(&models.OrderItem{}).
		Where("order_id = ?", order.ID).
		Update("status", "cancelled").Error; err != nil {
		tx.Rollback()
		response.GenerateInternalServerErrorResponse(c, "order/cancel_order", "Failed to update order items")
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "order/cancel_order", "Failed to commit transaction")
		return
	}

	// Load the complete order with relationships for response
	var completeOrder models.Order
	if err := h.db.
		Preload("User").
		Preload("ShippingAddress").
		Preload("Items.ProductVariant.Product").
		Preload("Items.ProductVariant.Product.Images").
		Preload("Items.ProductVariant.OptionValues").
		Preload("Items.Product"). // Legacy support
		First(&completeOrder, order.ID).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "order/cancel_order", "Order cancelled but failed to load details")
		return
	}

	response.GenerateSuccessResponse(c, "Order cancelled successfully", completeOrder)
}
