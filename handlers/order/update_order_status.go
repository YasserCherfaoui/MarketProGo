package order

import (
	"time"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UpdateOrderStatusRequest struct {
	Status         models.OrderStatus `json:"status" binding:"required"`
	AdminNotes     string             `json:"admin_notes"`
	TrackingNumber string             `json:"tracking_number"`
}

// UpdateOrderStatus - Admin endpoint to update order status
func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		response.GenerateBadRequestResponse(c, "order/update_status", "Order ID is required")
		return
	}

	var req UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GenerateBadRequestResponse(c, "order/update_status", err.Error())
		return
	}

	// Validate status
	validStatuses := []models.OrderStatus{
		models.OrderStatusPending,
		models.OrderStatusProcessing,
		models.OrderStatusShipped,
		models.OrderStatusDelivered,
		models.OrderStatusCancelled,
		models.OrderStatusReturned,
	}

	isValidStatus := false
	for _, status := range validStatuses {
		if status == req.Status {
			isValidStatus = true
			break
		}
	}

	if !isValidStatus {
		response.GenerateBadRequestResponse(c, "order/update_status", "Invalid order status")
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
	if err := tx.First(&order, orderID).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			response.GenerateNotFoundResponse(c, "order/update_status", "Order not found")
		} else {
			response.GenerateInternalServerErrorResponse(c, "order/update_status", "Failed to get order")
		}
		return
	}

	// Validate status transition
	if !isValidStatusTransition(order.Status, req.Status) {
		tx.Rollback()
		response.GenerateBadRequestResponse(c, "order/update_status", "Invalid status transition")
		return
	}

	// Update order
	now := time.Now()
	order.Status = req.Status
	order.AdminNotes = req.AdminNotes

	// Set specific date fields based on status
	switch req.Status {
	case models.OrderStatusShipped:
		if order.ShippedDate == nil {
			order.ShippedDate = &now
		}
		if req.TrackingNumber != "" {
			order.TrackingNumber = req.TrackingNumber
		}
	case models.OrderStatusDelivered:
		if order.DeliveredDate == nil {
			order.DeliveredDate = &now
		}
		// Auto-update payment status if not already paid
		if order.PaymentStatus == models.PaymentStatusPending {
			order.PaymentStatus = models.PaymentStatusPaid
			order.PaymentDate = &now
		}
	}

	if err := tx.Save(&order).Error; err != nil {
		tx.Rollback()
		response.GenerateInternalServerErrorResponse(c, "order/update_status", "Failed to update order status")
		return
	}

	// Update order items status if order is cancelled or returned
	if req.Status == models.OrderStatusCancelled || req.Status == models.OrderStatusReturned {
		itemStatus := "cancelled"
		if req.Status == models.OrderStatusReturned {
			itemStatus = "returned"
		}

		if err := tx.Model(&models.OrderItem{}).
			Where("order_id = ?", order.ID).
			Update("status", itemStatus).Error; err != nil {
			tx.Rollback()
			response.GenerateInternalServerErrorResponse(c, "order/update_status", "Failed to update order items")
			return
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "order/update_status", "Failed to commit transaction")
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
		response.GenerateInternalServerErrorResponse(c, "order/update_status", "Order updated but failed to load details")
		return
	}

	response.GenerateSuccessResponse(c, "Order status updated successfully", completeOrder)
}

// isValidStatusTransition validates if the status transition is allowed
func isValidStatusTransition(currentStatus, newStatus models.OrderStatus) bool {
	// Define valid transitions
	validTransitions := map[models.OrderStatus][]models.OrderStatus{
		models.OrderStatusPending: {
			models.OrderStatusProcessing,
			models.OrderStatusCancelled,
		},
		models.OrderStatusProcessing: {
			models.OrderStatusShipped,
			models.OrderStatusCancelled,
		},
		models.OrderStatusShipped: {
			models.OrderStatusDelivered,
			models.OrderStatusReturned,
		},
		models.OrderStatusDelivered: {
			models.OrderStatusReturned,
		},
		models.OrderStatusCancelled: {}, // No transitions allowed from cancelled
		models.OrderStatusReturned:  {}, // No transitions allowed from returned
	}

	allowedTransitions, exists := validTransitions[currentStatus]
	if !exists {
		return false
	}

	// Allow staying in the same status (for updating notes/tracking)
	if currentStatus == newStatus {
		return true
	}

	for _, allowedStatus := range allowedTransitions {
		if allowedStatus == newStatus {
			return true
		}
	}

	return false
}
