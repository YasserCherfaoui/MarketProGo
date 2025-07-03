package inventory

import (
	"strconv"
	"time"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

type BatchTrackingItem struct {
	InventoryItemID uint                    `json:"inventory_item_id"`
	ProductVariant  BatchProductVariantInfo `json:"product_variant"`
	Warehouse       BatchWarehouseInfo      `json:"warehouse"`
	BatchNumber     string                  `json:"batch_number"`
	Quantity        int                     `json:"quantity"`
	ExpiryDate      *time.Time              `json:"expiry_date"`
	DaysToExpiry    int                     `json:"days_to_expiry"`
	Status          string                  `json:"status"`
}

type BatchProductVariantInfo struct {
	ID      uint             `json:"id"`
	Name    string           `json:"name"`
	SKU     string           `json:"sku"`
	Product BatchProductInfo `json:"product"`
}

type BatchProductInfo struct {
	Name string `json:"name"`
}

type BatchWarehouseInfo struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

// GetInventoryBatches - Browse inventory by batch numbers and expiry dates
func (h *InventoryHandler) GetInventoryBatches(c *gin.Context) {
	// Query parameters
	warehouseID := c.Query("warehouse_id")
	productVariantID := c.Query("product_variant_id")
	expiringWithinDaysStr := c.Query("expiring_within_days")
	batchNumber := c.Query("batch_number")
	status := c.Query("status")

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

	var inventoryItems []models.InventoryItem

	// Base query
	db := h.db.Model(&models.InventoryItem{}).
		Preload("ProductVariant.Product").
		Preload("Warehouse")

	// Apply filters
	if warehouseID != "" {
		db = db.Where("warehouse_id = ?", warehouseID)
	}

	if productVariantID != "" {
		db = db.Where("product_variant_id = ?", productVariantID)
	}

	if batchNumber != "" {
		db = db.Where("batch_number ILIKE ?", "%"+batchNumber+"%")
	}

	if status != "" {
		db = db.Where("status = ?", status)
	}

	// Filter by expiry date if requested
	if expiringWithinDaysStr != "" {
		if expiringWithinDays, err := strconv.Atoi(expiringWithinDaysStr); err == nil {
			futureDate := time.Now().AddDate(0, 0, expiringWithinDays)
			db = db.Where("expiry_date IS NOT NULL AND expiry_date <= ?", futureDate)
		}
	}

	// Get total count
	var total int64
	if err := db.Count(&total).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "inventory/batches", err.Error())
		return
	}

	// Apply pagination and fetch data
	offset := (page - 1) * pageSize
	if err := db.Offset(offset).Limit(pageSize).Find(&inventoryItems).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "inventory/batches", err.Error())
		return
	}

	// Transform data to response format
	var batchItems []BatchTrackingItem
	for _, item := range inventoryItems {
		batchItem := BatchTrackingItem{
			InventoryItemID: item.ID,
			ProductVariant: BatchProductVariantInfo{
				ID:   item.ProductVariant.ID,
				Name: item.ProductVariant.Name,
				SKU:  item.ProductVariant.SKU,
				Product: BatchProductInfo{
					Name: item.ProductVariant.Product.Name,
				},
			},
			Warehouse: BatchWarehouseInfo{
				ID:   item.Warehouse.ID,
				Name: item.Warehouse.Name,
				Code: item.Warehouse.Code,
			},
			BatchNumber: item.BatchNumber,
			Quantity:    item.Quantity,
			ExpiryDate:  item.ExpiryDate,
			Status:      item.Status,
		}

		// Calculate days to expiry
		if item.ExpiryDate != nil {
			daysToExpiry := int(time.Until(*item.ExpiryDate).Hours() / 24)
			batchItem.DaysToExpiry = daysToExpiry
		}

		batchItems = append(batchItems, batchItem)
	}

	// Prepare paginated response
	resp := PaginatedResponse{
		Data:     batchItems,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}

	response.GenerateSuccessResponse(c, "Inventory batches retrieved successfully", resp)
}
