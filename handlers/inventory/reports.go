package inventory

import (
	"time"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

type InventoryDashboardReport struct {
	Summary          InventorySummary    `json:"summary"`
	WarehouseSummary []WarehouseSummary  `json:"warehouse_summary"`
	LowStockItems    []LowStockItem      `json:"low_stock_items"`
	ExpiringItems    []ExpiringItem      `json:"expiring_items"`
	TopProducts      []TopProductByValue `json:"top_products_by_value"`
	RecentMovements  []RecentMovement    `json:"recent_movements"`
	StockAlerts      []StockAlert        `json:"stock_alerts"`
}

type InventorySummary struct {
	TotalWarehouses    int64   `json:"total_warehouses"`
	ActiveWarehouses   int64   `json:"active_warehouses"`
	TotalProducts      int64   `json:"total_products"`
	TotalQuantity      int64   `json:"total_quantity"`
	ReservedQuantity   int64   `json:"reserved_quantity"`
	AvailableQuantity  int64   `json:"available_quantity"`
	TotalValue         float64 `json:"total_value"`
	LowStockCount      int64   `json:"low_stock_count"`
	OutOfStockCount    int64   `json:"out_of_stock_count"`
	ExpiringItemsCount int64   `json:"expiring_items_count"`
}

type WarehouseSummary struct {
	WarehouseID   uint    `json:"warehouse_id"`
	WarehouseName string  `json:"warehouse_name"`
	WarehouseCode string  `json:"warehouse_code"`
	ProductCount  int64   `json:"product_count"`
	TotalQuantity int64   `json:"total_quantity"`
	TotalValue    float64 `json:"total_value"`
	LowStockCount int64   `json:"low_stock_count"`
	IsActive      bool    `json:"is_active"`
}

type LowStockItem struct {
	ProductVariantID  uint   `json:"product_variant_id"`
	ProductName       string `json:"product_name"`
	VariantName       string `json:"variant_name"`
	SKU               string `json:"sku"`
	WarehouseName     string `json:"warehouse_name"`
	CurrentQuantity   int    `json:"current_quantity"`
	ReservedQuantity  int    `json:"reserved_quantity"`
	AvailableQuantity int    `json:"available_quantity"`
	ReorderLevel      int    `json:"reorder_level"`
}

type ExpiringItem struct {
	ProductVariantID uint       `json:"product_variant_id"`
	ProductName      string     `json:"product_name"`
	VariantName      string     `json:"variant_name"`
	SKU              string     `json:"sku"`
	WarehouseName    string     `json:"warehouse_name"`
	BatchNumber      string     `json:"batch_number"`
	Quantity         int        `json:"quantity"`
	ExpiryDate       *time.Time `json:"expiry_date"`
	DaysToExpiry     int        `json:"days_to_expiry"`
}

type TopProductByValue struct {
	ProductVariantID uint    `json:"product_variant_id"`
	ProductName      string  `json:"product_name"`
	VariantName      string  `json:"variant_name"`
	SKU              string  `json:"sku"`
	TotalQuantity    int64   `json:"total_quantity"`
	TotalValue       float64 `json:"total_value"`
	CostPrice        float64 `json:"cost_price"`
}

type RecentMovement struct {
	MovementID    uint      `json:"movement_id"`
	ProductName   string    `json:"product_name"`
	VariantName   string    `json:"variant_name"`
	SKU           string    `json:"sku"`
	WarehouseName string    `json:"warehouse_name"`
	MovementType  string    `json:"movement_type"`
	Quantity      int       `json:"quantity"`
	Reason        string    `json:"reason"`
	CreatedAt     time.Time `json:"created_at"`
	UserName      string    `json:"user_name,omitempty"`
}

type StockAlert struct {
	AlertType         string     `json:"alert_type"` // low_stock, out_of_stock, expiring_soon, expired
	ProductVariantID  uint       `json:"product_variant_id"`
	ProductName       string     `json:"product_name"`
	VariantName       string     `json:"variant_name"`
	SKU               string     `json:"sku"`
	WarehouseName     string     `json:"warehouse_name"`
	CurrentQuantity   int        `json:"current_quantity"`
	ThresholdQuantity int        `json:"threshold_quantity,omitempty"`
	ExpiryDate        *time.Time `json:"expiry_date,omitempty"`
	DaysToExpiry      int        `json:"days_to_expiry,omitempty"`
	Severity          string     `json:"severity"` // critical, warning, info
}

// GetInventoryDashboard - Admin endpoint to get comprehensive inventory dashboard data
func (h *InventoryHandler) GetInventoryDashboard(c *gin.Context) {
	// Get date range parameters
	daysBack := 30 // Default to last 30 days for movements
	if d := c.Query("days_back"); d != "" {
		if parsed, err := time.ParseDuration(d + "h"); err == nil {
			daysBack = int(parsed.Hours() / 24)
		}
	}

	report := InventoryDashboardReport{}

	// 1. Calculate overall summary
	report.Summary = h.calculateInventorySummary()

	// 2. Get warehouse summaries
	report.WarehouseSummary = h.getWarehouseSummaries()

	// 3. Get low stock items (quantity <= 10)
	report.LowStockItems = h.getLowStockItems(50) // Limit to top 50

	// 4. Get expiring items (expiring within 30 days)
	report.ExpiringItems = h.getExpiringItems(30, 50) // Next 30 days, limit 50

	// 5. Get top products by inventory value
	report.TopProducts = h.getTopProductsByValue(20) // Top 20

	// 6. Get recent movements
	report.RecentMovements = h.getRecentMovements(daysBack, 100) // Last N days, limit 100

	// 7. Generate stock alerts
	report.StockAlerts = h.generateStockAlerts()

	response.GenerateSuccessResponse(c, "Inventory dashboard data retrieved successfully", report)
}

// Helper methods for dashboard calculations
func (h *InventoryHandler) calculateInventorySummary() InventorySummary {
	var summary InventorySummary

	// Warehouse counts
	h.db.Model(&models.Warehouse{}).Count(&summary.TotalWarehouses)
	h.db.Model(&models.Warehouse{}).Where("is_active = ?", true).Count(&summary.ActiveWarehouses)

	// Product counts
	h.db.Model(&models.InventoryItem{}).Distinct("product_variant_id").Count(&summary.TotalProducts)

	// Quantity summaries
	h.db.Model(&models.InventoryItem{}).Select("COALESCE(SUM(quantity), 0)").Row().Scan(&summary.TotalQuantity)
	h.db.Model(&models.InventoryItem{}).Select("COALESCE(SUM(reserved), 0)").Row().Scan(&summary.ReservedQuantity)
	summary.AvailableQuantity = summary.TotalQuantity - summary.ReservedQuantity

	// Value calculation
	var totalValue interface{}
	h.db.Table("inventory_items").
		Select("COALESCE(SUM(inventory_items.quantity * product_variants.cost_price), 0)").
		Joins("JOIN product_variants ON product_variants.id = inventory_items.product_variant_id").
		Row().Scan(&totalValue)

	if totalValue != nil {
		if val, ok := totalValue.(float64); ok {
			summary.TotalValue = val
		}
	}

	// Stock status counts
	h.db.Model(&models.InventoryItem{}).Where("quantity <= ?", 10).Count(&summary.LowStockCount)
	h.db.Model(&models.InventoryItem{}).Where("quantity = ?", 0).Count(&summary.OutOfStockCount)

	// Expiring items (within 30 days)
	thirtyDaysFromNow := time.Now().AddDate(0, 0, 30)
	h.db.Model(&models.InventoryItem{}).
		Where("expiry_date IS NOT NULL AND expiry_date <= ?", thirtyDaysFromNow).
		Count(&summary.ExpiringItemsCount)

	return summary
}

func (h *InventoryHandler) getWarehouseSummaries() []WarehouseSummary {
	var warehouses []models.Warehouse
	h.db.Find(&warehouses)

	var summaries []WarehouseSummary
	for _, warehouse := range warehouses {
		summary := WarehouseSummary{
			WarehouseID:   warehouse.ID,
			WarehouseName: warehouse.Name,
			WarehouseCode: warehouse.Code,
			IsActive:      warehouse.IsActive,
		}

		// Product count
		h.db.Model(&models.InventoryItem{}).
			Where("warehouse_id = ?", warehouse.ID).
			Distinct("product_variant_id").
			Count(&summary.ProductCount)

		// Total quantity
		h.db.Model(&models.InventoryItem{}).
			Where("warehouse_id = ?", warehouse.ID).
			Select("COALESCE(SUM(quantity), 0)").
			Row().Scan(&summary.TotalQuantity)

		// Total value
		var totalValue interface{}
		h.db.Table("inventory_items").
			Select("COALESCE(SUM(inventory_items.quantity * product_variants.cost_price), 0)").
			Joins("JOIN product_variants ON product_variants.id = inventory_items.product_variant_id").
			Where("inventory_items.warehouse_id = ?", warehouse.ID).
			Row().Scan(&totalValue)

		if totalValue != nil {
			if val, ok := totalValue.(float64); ok {
				summary.TotalValue = val
			}
		}

		// Low stock count
		h.db.Model(&models.InventoryItem{}).
			Where("warehouse_id = ? AND quantity <= ?", warehouse.ID, 10).
			Count(&summary.LowStockCount)

		summaries = append(summaries, summary)
	}

	return summaries
}

func (h *InventoryHandler) getLowStockItems(limit int) []LowStockItem {
	var items []LowStockItem

	query := `
		SELECT 
			i.product_variant_id,
			p.name as product_name,
			pv.name as variant_name,
			pv.sku,
			w.name as warehouse_name,
			i.quantity as current_quantity,
			i.reserved as reserved_quantity,
			(i.quantity - i.reserved) as available_quantity,
			10 as reorder_level
		FROM inventory_items i
		JOIN product_variants pv ON pv.id = i.product_variant_id
		JOIN products p ON p.id = pv.product_id
		JOIN warehouses w ON w.id = i.warehouse_id
		WHERE i.quantity <= 10
		ORDER BY i.quantity ASC
		LIMIT ?
	`

	h.db.Raw(query, limit).Scan(&items)
	return items
}

func (h *InventoryHandler) getExpiringItems(daysAhead, limit int) []ExpiringItem {
	var items []ExpiringItem
	futureDate := time.Now().AddDate(0, 0, daysAhead)

	query := `
		SELECT 
			i.product_variant_id,
			p.name as product_name,
			pv.name as variant_name,
			pv.sku,
			w.name as warehouse_name,
			i.batch_number,
			i.quantity,
			i.expiry_date,
			EXTRACT(DAY FROM (i.expiry_date - NOW())) as days_to_expiry
		FROM inventory_items i
		JOIN product_variants pv ON pv.id = i.product_variant_id
		JOIN products p ON p.id = pv.product_id
		JOIN warehouses w ON w.id = i.warehouse_id
		WHERE i.expiry_date IS NOT NULL 
		  AND i.expiry_date <= ?
		  AND i.quantity > 0
		ORDER BY i.expiry_date ASC
		LIMIT ?
	`

	h.db.Raw(query, futureDate, limit).Scan(&items)
	return items
}

func (h *InventoryHandler) getTopProductsByValue(limit int) []TopProductByValue {
	var items []TopProductByValue

	query := `
		SELECT 
			pv.id as product_variant_id,
			p.name as product_name,
			pv.name as variant_name,
			pv.sku,
			SUM(i.quantity) as total_quantity,
			SUM(i.quantity * pv.cost_price) as total_value,
			pv.cost_price
		FROM inventory_items i
		JOIN product_variants pv ON pv.id = i.product_variant_id
		JOIN products p ON p.id = pv.product_id
		GROUP BY pv.id, p.name, pv.name, pv.sku, pv.cost_price
		ORDER BY total_value DESC
		LIMIT ?
	`

	h.db.Raw(query, limit).Scan(&items)
	return items
}

func (h *InventoryHandler) getRecentMovements(daysBack, limit int) []RecentMovement {
	var movements []RecentMovement
	fromDate := time.Now().AddDate(0, 0, -daysBack)

	query := `
		SELECT 
			sm.id as movement_id,
			p.name as product_name,
			pv.name as variant_name,
			pv.sku,
			w.name as warehouse_name,
			sm.movement_type,
			sm.quantity,
			sm.reason,
			sm.created_at,
			COALESCE(u.email, '') as user_name
		FROM stock_movements sm
		JOIN inventory_items i ON i.id = sm.inventory_item_id
		JOIN product_variants pv ON pv.id = i.product_variant_id
		JOIN products p ON p.id = pv.product_id
		JOIN warehouses w ON w.id = i.warehouse_id
		LEFT JOIN users u ON u.id = sm.user_id
		WHERE sm.created_at >= ?
		ORDER BY sm.created_at DESC
		LIMIT ?
	`

	h.db.Raw(query, fromDate, limit).Scan(&movements)
	return movements
}

func (h *InventoryHandler) generateStockAlerts() []StockAlert {
	var alerts []StockAlert

	// Low stock alerts
	lowStockItems := h.getLowStockItems(100)
	for _, item := range lowStockItems {
		alert := StockAlert{
			AlertType:         "low_stock",
			ProductVariantID:  item.ProductVariantID,
			ProductName:       item.ProductName,
			VariantName:       item.VariantName,
			SKU:               item.SKU,
			WarehouseName:     item.WarehouseName,
			CurrentQuantity:   item.CurrentQuantity,
			ThresholdQuantity: 10,
		}

		if item.CurrentQuantity == 0 {
			alert.AlertType = "out_of_stock"
			alert.Severity = "critical"
		} else if item.CurrentQuantity <= 5 {
			alert.Severity = "critical"
		} else {
			alert.Severity = "warning"
		}

		alerts = append(alerts, alert)
	}

	// Expiring item alerts
	expiringItems := h.getExpiringItems(30, 100)
	for _, item := range expiringItems {
		alert := StockAlert{
			AlertType:        "expiring_soon",
			ProductVariantID: item.ProductVariantID,
			ProductName:      item.ProductName,
			VariantName:      item.VariantName,
			SKU:              item.SKU,
			WarehouseName:    item.WarehouseName,
			CurrentQuantity:  item.Quantity,
			ExpiryDate:       item.ExpiryDate,
			DaysToExpiry:     item.DaysToExpiry,
		}

		if item.DaysToExpiry <= 0 {
			alert.AlertType = "expired"
			alert.Severity = "critical"
		} else if item.DaysToExpiry <= 7 {
			alert.Severity = "critical"
		} else {
			alert.Severity = "warning"
		}

		alerts = append(alerts, alert)
	}

	return alerts
}
