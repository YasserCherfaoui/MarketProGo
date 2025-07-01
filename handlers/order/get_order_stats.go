package order

import (
	"time"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

type OrderStats struct {
	// Overall statistics
	TotalOrders       int64   `json:"total_orders"`
	TotalRevenue      float64 `json:"total_revenue"`
	AverageOrderValue float64 `json:"average_order_value"`

	// Order status breakdown
	PendingOrders    int64 `json:"pending_orders"`
	ProcessingOrders int64 `json:"processing_orders"`
	ShippedOrders    int64 `json:"shipped_orders"`
	DeliveredOrders  int64 `json:"delivered_orders"`
	CancelledOrders  int64 `json:"cancelled_orders"`
	ReturnedOrders   int64 `json:"returned_orders"`

	// Payment status breakdown
	PendingPayments int64 `json:"pending_payments"`
	PaidOrders      int64 `json:"paid_orders"`
	FailedPayments  int64 `json:"failed_payments"`
	RefundedOrders  int64 `json:"refunded_orders"`

	// Time-based statistics
	TodayOrders  int64   `json:"today_orders"`
	TodayRevenue float64 `json:"today_revenue"`
	WeekOrders   int64   `json:"week_orders"`
	WeekRevenue  float64 `json:"week_revenue"`
	MonthOrders  int64   `json:"month_orders"`
	MonthRevenue float64 `json:"month_revenue"`

	// Recent activity
	RecentOrders []models.Order `json:"recent_orders"`
}

// GetOrderStats - Admin endpoint to get order statistics for dashboard
func (h *OrderHandler) GetOrderStats(c *gin.Context) {
	var stats OrderStats

	// Calculate date ranges
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	startOfWeek := startOfDay.AddDate(0, 0, -int(now.Weekday()))
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	// Overall statistics
	h.db.Model(&models.Order{}).Count(&stats.TotalOrders)

	h.db.Model(&models.Order{}).
		Select("COALESCE(SUM(final_amount), 0)").
		Where("status != ?", models.OrderStatusCancelled).
		Scan(&stats.TotalRevenue)

	if stats.TotalOrders > 0 {
		var completedOrders int64
		h.db.Model(&models.Order{}).
			Where("status != ?", models.OrderStatusCancelled).
			Count(&completedOrders)
		if completedOrders > 0 {
			stats.AverageOrderValue = stats.TotalRevenue / float64(completedOrders)
		}
	}

	// Order status breakdown
	h.db.Model(&models.Order{}).Where("status = ?", models.OrderStatusPending).Count(&stats.PendingOrders)
	h.db.Model(&models.Order{}).Where("status = ?", models.OrderStatusProcessing).Count(&stats.ProcessingOrders)
	h.db.Model(&models.Order{}).Where("status = ?", models.OrderStatusShipped).Count(&stats.ShippedOrders)
	h.db.Model(&models.Order{}).Where("status = ?", models.OrderStatusDelivered).Count(&stats.DeliveredOrders)
	h.db.Model(&models.Order{}).Where("status = ?", models.OrderStatusCancelled).Count(&stats.CancelledOrders)
	h.db.Model(&models.Order{}).Where("status = ?", models.OrderStatusReturned).Count(&stats.ReturnedOrders)

	// Payment status breakdown
	h.db.Model(&models.Order{}).Where("payment_status = ?", models.PaymentStatusPending).Count(&stats.PendingPayments)
	h.db.Model(&models.Order{}).Where("payment_status = ?", models.PaymentStatusPaid).Count(&stats.PaidOrders)
	h.db.Model(&models.Order{}).Where("payment_status = ?", models.PaymentStatusFailed).Count(&stats.FailedPayments)
	h.db.Model(&models.Order{}).Where("payment_status = ?", models.PaymentStatusRefunded).Count(&stats.RefundedOrders)

	// Today's statistics
	h.db.Model(&models.Order{}).
		Where("order_date >= ?", startOfDay).
		Count(&stats.TodayOrders)

	h.db.Model(&models.Order{}).
		Select("COALESCE(SUM(final_amount), 0)").
		Where("order_date >= ? AND status != ?", startOfDay, models.OrderStatusCancelled).
		Scan(&stats.TodayRevenue)

	// This week's statistics
	h.db.Model(&models.Order{}).
		Where("order_date >= ?", startOfWeek).
		Count(&stats.WeekOrders)

	h.db.Model(&models.Order{}).
		Select("COALESCE(SUM(final_amount), 0)").
		Where("order_date >= ? AND status != ?", startOfWeek, models.OrderStatusCancelled).
		Scan(&stats.WeekRevenue)

	// This month's statistics
	h.db.Model(&models.Order{}).
		Where("order_date >= ?", startOfMonth).
		Count(&stats.MonthOrders)

	h.db.Model(&models.Order{}).
		Select("COALESCE(SUM(final_amount), 0)").
		Where("order_date >= ? AND status != ?", startOfMonth, models.OrderStatusCancelled).
		Scan(&stats.MonthRevenue)

	// Recent orders (last 10)
	h.db.Preload("User").
		Preload("ShippingAddress").
		Order("created_at DESC").
		Limit(10).
		Find(&stats.RecentOrders)

	response.GenerateSuccessResponse(c, "Order statistics retrieved successfully", stats)
}
