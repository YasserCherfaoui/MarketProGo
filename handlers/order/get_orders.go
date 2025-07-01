package order

import (
	"strconv"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *OrderHandler) GetOrders(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.GenerateUnauthorizedResponse(c, "order/get_orders", "User not authenticated")
		return
	}
	uid := userID.(uint)

	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	status := c.Query("status")
	paymentStatus := c.Query("payment_status")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	// Build query
	query := h.db.Where("user_id = ?", uid)

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if paymentStatus != "" {
		query = query.Where("payment_status = ?", paymentStatus)
	}

	// Get total count
	var totalCount int64
	query.Model(&models.Order{}).Count(&totalCount)

	// Get orders with preloaded relationships
	var orders []models.Order
	if err := query.
		Preload("User").
		Preload("ShippingAddress").
		Preload("Items.ProductVariant.Product").
		Preload("Items.ProductVariant.Product.Images").
		Preload("Items.ProductVariant.OptionValues").
		Order("order_date DESC").
		Limit(limit).
		Offset(offset).
		Find(&orders).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "order/get_orders", "Failed to get orders")
		return
	}

	// Prepare response
	responseData := map[string]interface{}{
		"orders":      orders,
		"page":        page,
		"limit":       limit,
		"total_count": totalCount,
		"total_pages": (totalCount + int64(limit) - 1) / int64(limit),
	}

	response.GenerateSuccessResponse(c, "Orders retrieved successfully", responseData)
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.GenerateUnauthorizedResponse(c, "order/get_order", "User not authenticated")
		return
	}
	uid := userID.(uint)

	orderID := c.Param("id")
	if orderID == "" {
		response.GenerateBadRequestResponse(c, "order/get_order", "Order ID is required")
		return
	}

	var order models.Order
	if err := h.db.
		Preload("User").
		Preload("ShippingAddress").
		Preload("Items.ProductVariant.Product").
		Preload("Items.ProductVariant.Product.Images").
		Preload("Items.ProductVariant.OptionValues").
		Preload("Items.Product"). // Legacy support
		Where("id = ? AND user_id = ?", orderID, uid).
		First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.GenerateNotFoundResponse(c, "order/get_order", "Order not found")
		} else {
			response.GenerateInternalServerErrorResponse(c, "order/get_order", "Failed to get order")
		}
		return
	}

	response.GenerateSuccessResponse(c, "Order retrieved successfully", order)
}
