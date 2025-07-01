package order

import (
	"strconv"
	"strings"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

// GetAllOrders - Admin endpoint to get all orders with filtering and search
func (h *OrderHandler) GetAllOrders(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")
	paymentStatus := c.Query("payment_status")
	search := c.Query("search") // Search by order number, customer name, or email
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	sortBy := c.DefaultQuery("sort_by", "order_date")
	sortOrder := c.DefaultQuery("sort_order", "desc")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Build query
	query := h.db.Model(&models.Order{})

	// Apply filters
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if paymentStatus != "" {
		query = query.Where("payment_status = ?", paymentStatus)
	}
	if startDate != "" {
		query = query.Where("order_date >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("order_date <= ?", endDate)
	}

	// Apply search
	if search != "" {
		searchTerm := "%" + strings.ToLower(search) + "%"
		query = query.Joins("LEFT JOIN users ON orders.user_id = users.id").
			Where("LOWER(orders.order_number) LIKE ? OR LOWER(users.first_name) LIKE ? OR LOWER(users.last_name) LIKE ? OR LOWER(users.email) LIKE ?",
				searchTerm, searchTerm, searchTerm, searchTerm)
	}

	// Get total count
	var totalCount int64
	query.Count(&totalCount)

	// Apply sorting
	orderClause := sortBy
	if sortOrder == "desc" {
		orderClause += " DESC"
	} else {
		orderClause += " ASC"
	}

	// Get orders with preloaded relationships
	var orders []models.Order
	if err := query.
		Preload("User").
		Preload("ShippingAddress").
		Preload("Items.ProductVariant.Product.Images").
		Preload("Items.ProductVariant.OptionValues").
		Order(orderClause).
		Limit(limit).
		Offset(offset).
		Find(&orders).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "order/get_all_orders", "Failed to get orders "+err.Error())
		return
	}

	// Calculate statistics
	var stats struct {
		TotalRevenue     float64 `json:"total_revenue"`
		PendingOrders    int64   `json:"pending_orders"`
		ProcessingOrders int64   `json:"processing_orders"`
		ShippedOrders    int64   `json:"shipped_orders"`
		DeliveredOrders  int64   `json:"delivered_orders"`
		CancelledOrders  int64   `json:"cancelled_orders"`
	}

	h.db.Model(&models.Order{}).Select("COALESCE(SUM(final_amount), 0)").Where("status != ?", models.OrderStatusCancelled).Scan(&stats.TotalRevenue)
	h.db.Model(&models.Order{}).Where("status = ?", models.OrderStatusPending).Count(&stats.PendingOrders)
	h.db.Model(&models.Order{}).Where("status = ?", models.OrderStatusProcessing).Count(&stats.ProcessingOrders)
	h.db.Model(&models.Order{}).Where("status = ?", models.OrderStatusShipped).Count(&stats.ShippedOrders)
	h.db.Model(&models.Order{}).Where("status = ?", models.OrderStatusDelivered).Count(&stats.DeliveredOrders)
	h.db.Model(&models.Order{}).Where("status = ?", models.OrderStatusCancelled).Count(&stats.CancelledOrders)

	// Prepare response
	responseData := map[string]interface{}{
		"orders":      orders,
		"page":        page,
		"limit":       limit,
		"total_count": totalCount,
		"total_pages": (totalCount + int64(limit) - 1) / int64(limit),
		"statistics":  stats,
	}

	response.GenerateSuccessResponse(c, "Orders retrieved successfully", responseData)
}
