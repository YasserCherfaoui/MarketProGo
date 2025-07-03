package inventory

import (
	"strconv"

	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

// GetStockAlerts - Browse and manage stock alerts
func (h *InventoryHandler) GetStockAlerts(c *gin.Context) {
	// Query parameters
	alertType := c.Query("alert_type") // low_stock, out_of_stock, expiring_soon, expired
	severity := c.Query("severity")    // critical, warning, info
	warehouseID := c.Query("warehouse_id")

	// Pagination parameters
	page := 1
	pageSize := 50
	if p := c.Query("page"); p != "" {
		if parsedPage, err := strconv.Atoi(p); err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if parsedPageSize, err := strconv.Atoi(ps); err == nil && parsedPageSize > 0 {
			pageSize = parsedPageSize
		}
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// Generate all stock alerts using the existing method
	allAlerts := h.generateStockAlerts()

	// Apply filters
	var filteredAlerts []StockAlert
	for _, alert := range allAlerts {
		include := true

		// Filter by alert type
		if alertType != "" && alert.AlertType != alertType {
			include = false
		}

		// Filter by severity
		if severity != "" && alert.Severity != severity {
			include = false
		}

		// Filter by warehouse (we need to get warehouse ID by name comparison)
		if warehouseID != "" {
			if warehouseIDInt, err := strconv.Atoi(warehouseID); err == nil {
				// We need to check the warehouse ID from the database
				// This is a simplified approach - in production you might want to join properly
				var warehouseName string
				h.db.Table("warehouses").Where("id = ?", warehouseIDInt).Select("name").Scan(&warehouseName)
				if alert.WarehouseName != warehouseName {
					include = false
				}
			}
		}

		if include {
			filteredAlerts = append(filteredAlerts, alert)
		}
	}

	// Apply pagination
	total := int64(len(filteredAlerts))
	start := (page - 1) * pageSize
	end := start + pageSize

	if start > len(filteredAlerts) {
		start = len(filteredAlerts)
	}
	if end > len(filteredAlerts) {
		end = len(filteredAlerts)
	}

	paginatedAlerts := filteredAlerts[start:end]

	// Prepare paginated response
	resp := PaginatedResponse{
		Data:     paginatedAlerts,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}

	response.GenerateSuccessResponse(c, "Stock alerts retrieved successfully", resp)
}
