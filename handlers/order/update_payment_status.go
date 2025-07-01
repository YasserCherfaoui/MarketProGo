package order

import (
	"time"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UpdatePaymentStatusRequest struct {
	PaymentStatus    models.PaymentStatus `json:"payment_status" binding:"required"`
	PaymentReference string               `json:"payment_reference"`
	AdminNotes       string               `json:"admin_notes"`
}

// UpdatePaymentStatus - Admin endpoint to update payment status
func (h *OrderHandler) UpdatePaymentStatus(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		response.GenerateBadRequestResponse(c, "order/update_payment", "Order ID is required")
		return
	}

	var req UpdatePaymentStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GenerateBadRequestResponse(c, "order/update_payment", err.Error())
		return
	}

	// Validate payment status
	validPaymentStatuses := []models.PaymentStatus{
		models.PaymentStatusPending,
		models.PaymentStatusPaid,
		models.PaymentStatusFailed,
		models.PaymentStatusRefunded,
	}

	isValidPaymentStatus := false
	for _, status := range validPaymentStatuses {
		if status == req.PaymentStatus {
			isValidPaymentStatus = true
			break
		}
	}

	if !isValidPaymentStatus {
		response.GenerateBadRequestResponse(c, "order/update_payment", "Invalid payment status")
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
			response.GenerateNotFoundResponse(c, "order/update_payment", "Order not found")
		} else {
			response.GenerateInternalServerErrorResponse(c, "order/update_payment", "Failed to get order")
		}
		return
	}

	// Update payment information
	now := time.Now()
	order.PaymentStatus = req.PaymentStatus
	order.AdminNotes = req.AdminNotes

	if req.PaymentReference != "" {
		order.PaymentReference = req.PaymentReference
	}

	// Set payment date if status is paid
	if req.PaymentStatus == models.PaymentStatusPaid && order.PaymentDate == nil {
		order.PaymentDate = &now
	}

	// Handle refund logic
	if req.PaymentStatus == models.PaymentStatusRefunded {
		// If order is not already cancelled/returned, mark it as returned
		if order.Status != models.OrderStatusCancelled && order.Status != models.OrderStatusReturned {
			order.Status = models.OrderStatusReturned
		}

		// Update order items status
		if err := tx.Model(&models.OrderItem{}).
			Where("order_id = ?", order.ID).
			Update("status", "returned").Error; err != nil {
			tx.Rollback()
			response.GenerateInternalServerErrorResponse(c, "order/update_payment", "Failed to update order items")
			return
		}
	}

	if err := tx.Save(&order).Error; err != nil {
		tx.Rollback()
		response.GenerateInternalServerErrorResponse(c, "order/update_payment", "Failed to update payment status")
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "order/update_payment", "Failed to commit transaction")
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
		response.GenerateInternalServerErrorResponse(c, "order/update_payment", "Payment updated but failed to load details")
		return
	}

	response.GenerateSuccessResponse(c, "Payment status updated successfully", completeOrder)
}
