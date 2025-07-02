package inventory

import (
	"fmt"
	"time"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type StockAdjustmentRequest struct {
	ProductVariantID uint    `json:"product_variant_id" binding:"required"`
	WarehouseID      uint    `json:"warehouse_id" binding:"required"`
	Quantity         int     `json:"quantity" binding:"required"`
	BatchNumber      string  `json:"batch_number"`
	ExpiryDate       *string `json:"expiry_date"` // YYYY-MM-DD format
	Reason           string  `json:"reason" binding:"required"`
	Notes            string  `json:"notes"`
}

type BulkStockAdjustmentRequest struct {
	WarehouseID uint                     `json:"warehouse_id" binding:"required"`
	Items       []StockAdjustmentRequest `json:"items" binding:"required,dive"`
	Reason      string                   `json:"reason" binding:"required"`
	Notes       string                   `json:"notes"`
}

type StockTransferRequest struct {
	ProductVariantID  uint   `json:"product_variant_id" binding:"required"`
	FromWarehouseID   uint   `json:"from_warehouse_id" binding:"required"`
	ToWarehouseID     uint   `json:"to_warehouse_id" binding:"required"`
	Quantity          int    `json:"quantity" binding:"required,min=1"`
	TransferReference string `json:"transfer_reference"`
	Notes             string `json:"notes"`
}

type StockReservationRequest struct {
	ProductVariantID uint   `json:"product_variant_id" binding:"required"`
	WarehouseID      uint   `json:"warehouse_id" binding:"required"`
	Quantity         int    `json:"quantity" binding:"required,min=1"`
	OrderID          *uint  `json:"order_id"`
	ReservationType  string `json:"reservation_type" binding:"required"` // order, manual
	Notes            string `json:"notes"`
}

type StockLevelResponse struct {
	models.InventoryItem
	ProductVariant    models.ProductVariant `json:"product_variant"`
	Warehouse         models.Warehouse      `json:"warehouse"`
	AvailableQuantity int                   `json:"available_quantity"`
	StockStatus       string                `json:"stock_status"`
}

// GetStockLevels - Admin endpoint to get stock levels with filtering
func (h *InventoryHandler) GetStockLevels(c *gin.Context) {
	// Query parameters
	warehouseID := c.Query("warehouse_id")
	productVariantID := c.Query("product_variant_id")
	sku := c.Query("sku")
	status := c.Query("status")

	// Pagination
	page := 1
	pageSize := 50
	if p := c.Query("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	if ps := c.Query("page_size"); ps != "" {
		fmt.Sscanf(ps, "%d", &pageSize)
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 50
	} else if pageSize > 100 {
		pageSize = 100
	}

	// Build query
	query := h.db.Model(&models.InventoryItem{}).
		Preload("ProductVariant.Product").
		Preload("Warehouse.Address")

	// Apply filters
	if warehouseID != "" {
		query = query.Where("warehouse_id = ?", warehouseID)
	}
	if productVariantID != "" {
		query = query.Where("product_variant_id = ?", productVariantID)
	}
	if sku != "" {
		query = query.Joins("JOIN product_variants ON product_variants.id = inventory_items.product_variant_id").
			Where("product_variants.sku ILIKE ?", "%"+sku+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get inventory items
	var inventoryItems []models.InventoryItem
	if err := query.
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&inventoryItems).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "inventory/get_stock_levels", "Failed to get stock levels")
		return
	}

	// Transform to response format
	var stockLevels []StockLevelResponse
	for _, item := range inventoryItems {
		availableQuantity := item.Quantity - item.Reserved
		stockStatus := h.getStockStatus(availableQuantity)

		stockLevel := StockLevelResponse{
			InventoryItem:     item,
			ProductVariant:    item.ProductVariant,
			Warehouse:         item.Warehouse,
			AvailableQuantity: availableQuantity,
			StockStatus:       stockStatus,
		}
		stockLevels = append(stockLevels, stockLevel)
	}

	resp := map[string]interface{}{
		"data":      stockLevels,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}
	response.GenerateSuccessResponse(c, "Stock levels retrieved successfully", resp)
}

// AdjustStock - Admin endpoint to adjust stock levels
func (h *InventoryHandler) AdjustStock(c *gin.Context) {
	var req StockAdjustmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GenerateBadRequestResponse(c, "inventory/adjust_stock", err.Error())
		return
	}

	// Validate product variant and warehouse exist
	var variant models.ProductVariant
	if err := h.db.First(&variant, req.ProductVariantID).Error; err != nil {
		response.GenerateBadRequestResponse(c, "inventory/adjust_stock", "Product variant not found")
		return
	}

	var warehouse models.Warehouse
	if err := h.db.First(&warehouse, req.WarehouseID).Error; err != nil {
		response.GenerateBadRequestResponse(c, "inventory/adjust_stock", "Warehouse not found")
		return
	}

	if !warehouse.IsActive {
		response.GenerateBadRequestResponse(c, "inventory/adjust_stock", "Warehouse is not active")
		return
	}

	// Parse expiry date if provided
	var expiryDate *time.Time
	if req.ExpiryDate != nil && *req.ExpiryDate != "" {
		parsed, err := time.Parse("2006-01-02", *req.ExpiryDate)
		if err != nil {
			response.GenerateBadRequestResponse(c, "inventory/adjust_stock", "Invalid expiry date format. Use YYYY-MM-DD")
			return
		}
		expiryDate = &parsed
	}

	// Start transaction
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get or create inventory item
	var inventoryItem models.InventoryItem
	err := tx.Where("product_variant_id = ? AND warehouse_id = ? AND batch_number = ?",
		req.ProductVariantID, req.WarehouseID, req.BatchNumber).First(&inventoryItem).Error

	if err == gorm.ErrRecordNotFound {
		// Create new inventory item
		if req.Quantity < 0 {
			tx.Rollback()
			response.GenerateBadRequestResponse(c, "inventory/adjust_stock", "Cannot create new inventory item with negative quantity")
			return
		}

		inventoryItem = models.InventoryItem{
			ProductVariantID: req.ProductVariantID,
			WarehouseID:      req.WarehouseID,
			Quantity:         req.Quantity,
			BatchNumber:      req.BatchNumber,
			ExpiryDate:       expiryDate,
			Status:           "active",
		}

		if err := tx.Create(&inventoryItem).Error; err != nil {
			tx.Rollback()
			response.GenerateInternalServerErrorResponse(c, "inventory/adjust_stock", "Failed to create inventory item")
			return
		}
	} else if err != nil {
		tx.Rollback()
		response.GenerateInternalServerErrorResponse(c, "inventory/adjust_stock", "Failed to get inventory item")
		return
	} else {
		// Update existing inventory item
		newQuantity := inventoryItem.Quantity + req.Quantity
		if newQuantity < 0 {
			tx.Rollback()
			response.GenerateBadRequestResponse(c, "inventory/adjust_stock", fmt.Sprintf("Insufficient stock. Available: %d", inventoryItem.Quantity))
			return
		}

		inventoryItem.Quantity = newQuantity
		if err := tx.Save(&inventoryItem).Error; err != nil {
			tx.Rollback()
			response.GenerateInternalServerErrorResponse(c, "inventory/adjust_stock", "Failed to update inventory item")
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "inventory/adjust_stock", "Failed to commit transaction")
		return
	}

	// Load complete inventory item for response
	h.db.Preload("ProductVariant.Product").Preload("Warehouse").First(&inventoryItem, inventoryItem.ID)

	response.GenerateSuccessResponse(c, "Stock adjusted successfully", inventoryItem)
}

// BulkAdjustStock - Admin endpoint for bulk stock adjustments (CSV import support)
func (h *InventoryHandler) BulkAdjustStock(c *gin.Context) {
	var req BulkStockAdjustmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GenerateBadRequestResponse(c, "inventory/bulk_adjust_stock", err.Error())
		return
	}

	// Validate warehouse exists
	var warehouse models.Warehouse
	if err := h.db.First(&warehouse, req.WarehouseID).Error; err != nil {
		response.GenerateBadRequestResponse(c, "inventory/bulk_adjust_stock", "Warehouse not found")
		return
	}

	if !warehouse.IsActive {
		response.GenerateBadRequestResponse(c, "inventory/bulk_adjust_stock", "Warehouse is not active")
		return
	}

	// Start transaction
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var results []map[string]interface{}
	successCount := 0
	errorCount := 0

	for i, item := range req.Items {
		// Use warehouse from request
		item.WarehouseID = req.WarehouseID
		if item.Reason == "" {
			item.Reason = req.Reason
		}
		if item.Notes == "" {
			item.Notes = req.Notes
		}

		result := map[string]interface{}{
			"index":              i,
			"product_variant_id": item.ProductVariantID,
			"quantity":           item.Quantity,
		}

		// Process individual adjustment
		if err := h.processSingleStockAdjustment(tx, item); err != nil {
			result["status"] = "error"
			result["error"] = err.Error()
			errorCount++
		} else {
			result["status"] = "success"
			successCount++
		}

		results = append(results, result)
	}

	if errorCount > 0 && successCount == 0 {
		tx.Rollback()
		response.GenerateBadRequestResponse(c, "inventory/bulk_adjust_stock", "All adjustments failed")
		return
	}

	if err := tx.Commit().Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "inventory/bulk_adjust_stock", "Failed to commit transaction")
		return
	}

	resp := map[string]interface{}{
		"results":       results,
		"success_count": successCount,
		"error_count":   errorCount,
		"total_count":   len(req.Items),
	}

	response.GenerateSuccessResponse(c, "Bulk stock adjustment completed", resp)
}

// TransferStock - Admin endpoint to transfer stock between warehouses
func (h *InventoryHandler) TransferStock(c *gin.Context) {
	var req StockTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GenerateBadRequestResponse(c, "inventory/transfer_stock", err.Error())
		return
	}

	if req.FromWarehouseID == req.ToWarehouseID {
		response.GenerateBadRequestResponse(c, "inventory/transfer_stock", "Source and destination warehouses cannot be the same")
		return
	}

	// Validate warehouses and variant exist
	var fromWarehouse, toWarehouse models.Warehouse
	var variant models.ProductVariant

	if err := h.db.First(&fromWarehouse, req.FromWarehouseID).Error; err != nil {
		response.GenerateBadRequestResponse(c, "inventory/transfer_stock", "Source warehouse not found")
		return
	}

	if err := h.db.First(&toWarehouse, req.ToWarehouseID).Error; err != nil {
		response.GenerateBadRequestResponse(c, "inventory/transfer_stock", "Destination warehouse not found")
		return
	}

	if err := h.db.First(&variant, req.ProductVariantID).Error; err != nil {
		response.GenerateBadRequestResponse(c, "inventory/transfer_stock", "Product variant not found")
		return
	}

	// Start transaction
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Check available stock in source warehouse
	var sourceItem models.InventoryItem
	if err := tx.Where("product_variant_id = ? AND warehouse_id = ?", req.ProductVariantID, req.FromWarehouseID).First(&sourceItem).Error; err != nil {
		tx.Rollback()
		response.GenerateBadRequestResponse(c, "inventory/transfer_stock", "No stock found in source warehouse")
		return
	}

	availableQuantity := sourceItem.Quantity - sourceItem.Reserved
	if availableQuantity < req.Quantity {
		tx.Rollback()
		response.GenerateBadRequestResponse(c, "inventory/transfer_stock", fmt.Sprintf("Insufficient stock. Available: %d", availableQuantity))
		return
	}

	// Reduce stock from source
	sourceItem.Quantity -= req.Quantity
	if err := tx.Save(&sourceItem).Error; err != nil {
		tx.Rollback()
		response.GenerateInternalServerErrorResponse(c, "inventory/transfer_stock", "Failed to update source inventory")
		return
	}

	// Add stock to destination
	var destItem models.InventoryItem
	err := tx.Where("product_variant_id = ? AND warehouse_id = ?", req.ProductVariantID, req.ToWarehouseID).First(&destItem).Error

	if err == gorm.ErrRecordNotFound {
		// Create new inventory item in destination
		destItem = models.InventoryItem{
			ProductVariantID: req.ProductVariantID,
			WarehouseID:      req.ToWarehouseID,
			Quantity:         req.Quantity,
			Status:           "active",
		}
		if err := tx.Create(&destItem).Error; err != nil {
			tx.Rollback()
			response.GenerateInternalServerErrorResponse(c, "inventory/transfer_stock", "Failed to create destination inventory")
			return
		}
	} else if err != nil {
		tx.Rollback()
		response.GenerateInternalServerErrorResponse(c, "inventory/transfer_stock", "Failed to get destination inventory")
		return
	} else {
		// Update existing inventory item
		destItem.Quantity += req.Quantity
		if err := tx.Save(&destItem).Error; err != nil {
			tx.Rollback()
			response.GenerateInternalServerErrorResponse(c, "inventory/transfer_stock", "Failed to update destination inventory")
			return
		}
	}

	// Create stock movement records
	userID := h.getUserIDFromContext(c)

	// Source movement (out)
	sourceMovement := models.StockMovement{
		InventoryItemID: sourceItem.ID,
		MovementType:    "transfer_out",
		Quantity:        req.Quantity,
		Reason:          "Transfer to " + toWarehouse.Name,
		Notes:           req.Notes,
		UserID:          userID,
		Reference:       req.TransferReference,
	}

	// Destination movement (in)
	destMovement := models.StockMovement{
		InventoryItemID: destItem.ID,
		MovementType:    "transfer_in",
		Quantity:        req.Quantity,
		Reason:          "Transfer from " + fromWarehouse.Name,
		Notes:           req.Notes,
		UserID:          userID,
		Reference:       req.TransferReference,
	}

	if err := tx.Create(&sourceMovement).Error; err != nil {
		tx.Rollback()
		response.GenerateInternalServerErrorResponse(c, "inventory/transfer_stock", "Failed to create source movement record")
		return
	}

	if err := tx.Create(&destMovement).Error; err != nil {
		tx.Rollback()
		response.GenerateInternalServerErrorResponse(c, "inventory/transfer_stock", "Failed to create destination movement record")
		return
	}

	if err := tx.Commit().Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "inventory/transfer_stock", "Failed to commit transaction")
		return
	}

	resp := map[string]interface{}{
		"transfer_reference": req.TransferReference,
		"from_warehouse":     fromWarehouse.Name,
		"to_warehouse":       toWarehouse.Name,
		"product_variant":    variant.Name,
		"quantity":           req.Quantity,
		"status":             "completed",
	}

	response.GenerateSuccessResponse(c, "Stock transfer completed successfully", resp)
}

// Helper functions
func (h *InventoryHandler) getStockStatus(quantity int) string {
	if quantity == 0 {
		return "out_of_stock"
	} else if quantity <= 10 {
		return "low_stock"
	}
	return "in_stock"
}

func (h *InventoryHandler) getMovementType(quantity int) string {
	if quantity > 0 {
		return "adjustment_in"
	}
	return "adjustment_out"
}

func (h *InventoryHandler) getUserIDFromContext(c *gin.Context) *uint {
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(uint); ok {
			return &uid
		}
	}
	return nil
}

func (h *InventoryHandler) processSingleStockAdjustment(tx *gorm.DB, req StockAdjustmentRequest) error {
	// Validate product variant exists
	var variant models.ProductVariant
	if err := tx.First(&variant, req.ProductVariantID).Error; err != nil {
		return fmt.Errorf("product variant not found")
	}

	// Parse expiry date if provided
	var expiryDate *time.Time
	if req.ExpiryDate != nil && *req.ExpiryDate != "" {
		parsed, err := time.Parse("2006-01-02", *req.ExpiryDate)
		if err != nil {
			return fmt.Errorf("invalid expiry date format")
		}
		expiryDate = &parsed
	}

	// Get or create inventory item
	var inventoryItem models.InventoryItem
	err := tx.Where("product_variant_id = ? AND warehouse_id = ? AND batch_number = ?",
		req.ProductVariantID, req.WarehouseID, req.BatchNumber).First(&inventoryItem).Error

	if err == gorm.ErrRecordNotFound {
		if req.Quantity < 0 {
			return fmt.Errorf("cannot create new inventory item with negative quantity")
		}

		inventoryItem = models.InventoryItem{
			ProductVariantID: req.ProductVariantID,
			WarehouseID:      req.WarehouseID,
			Quantity:         req.Quantity,
			BatchNumber:      req.BatchNumber,
			ExpiryDate:       expiryDate,
			Status:           "active",
		}

		return tx.Create(&inventoryItem).Error
	} else if err != nil {
		return err
	} else {
		newQuantity := inventoryItem.Quantity + req.Quantity
		if newQuantity < 0 {
			return fmt.Errorf("insufficient stock")
		}

		inventoryItem.Quantity = newQuantity
		return tx.Save(&inventoryItem).Error
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
