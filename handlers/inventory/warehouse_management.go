package inventory

import (
	"fmt"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreateWarehouseRequest struct {
	Name      string `json:"name" binding:"required"`
	Code      string `json:"code" binding:"required"`
	AddressID uint   `json:"address_id" binding:"required"`
	IsActive  *bool  `json:"is_active"`
}

type UpdateWarehouseRequest struct {
	Name      *string `json:"name"`
	Code      *string `json:"code"`
	AddressID *uint   `json:"address_id"`
	IsActive  *bool   `json:"is_active"`
}

type WarehouseResponse struct {
	models.Warehouse `json:"warehouse"`
	StockSummary     StockSummary `json:"stock_summary"`
}

type StockSummary struct {
	TotalProducts    int64   `json:"total_products"`
	TotalQuantity    int64   `json:"total_quantity"`
	LowStockItems    int64   `json:"low_stock_items"`
	OutOfStockItems  int64   `json:"out_of_stock_items"`
	TotalValue       float64 `json:"total_value"`
	ReservedQuantity int64   `json:"reserved_quantity"`
}

// CreateWarehouse - Admin endpoint to create a new warehouse
func (h *InventoryHandler) CreateWarehouse(c *gin.Context) {
	var req CreateWarehouseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GenerateBadRequestResponse(c, "inventory/create_warehouse", err.Error())
		return
	}

	// Check if warehouse code already exists
	var existingWarehouse models.Warehouse
	if err := h.db.Where("code = ?", req.Code).First(&existingWarehouse).Error; err == nil {
		response.GenerateBadRequestResponse(c, "inventory/create_warehouse", "Warehouse code already exists")
		return
	}

	// Verify address exists
	var address models.Address
	if err := h.db.First(&address, req.AddressID).Error; err != nil {
		response.GenerateBadRequestResponse(c, "inventory/create_warehouse", "Address not found")
		return
	}

	// Set default is_active if not provided
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	warehouse := models.Warehouse{
		Name:      req.Name,
		Code:      req.Code,
		AddressID: req.AddressID,
		IsActive:  isActive,
	}

	if err := h.db.Create(&warehouse).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "inventory/create_warehouse", "Failed to create warehouse")
		return
	}

	// Load complete warehouse with address
	h.db.Preload("Address").First(&warehouse, warehouse.ID)

	response.GenerateCreatedResponse(c, "Warehouse created successfully", warehouse)
}

// GetAllWarehouses - Admin endpoint to get all warehouses with filtering and pagination
func (h *InventoryHandler) GetAllWarehouses(c *gin.Context) {
	// Query parameters
	name := c.Query("name")
	code := c.Query("code")
	isActive := c.Query("is_active")
	includeStock := c.DefaultQuery("include_stock", "false")

	// Pagination
	page := 1
	pageSize := 20
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
		pageSize = 20
	} else if pageSize > 100 {
		pageSize = 100
	}

	// Build query
	query := h.db.Model(&models.Warehouse{}).Preload("Address")

	// Apply filters
	if name != "" {
		query = query.Where("name ILIKE ?", "%"+name+"%")
	}
	if code != "" {
		query = query.Where("code ILIKE ?", "%"+code+"%")
	}
	if isActive != "" {
		query = query.Where("is_active = ?", isActive == "true")
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get warehouses with pagination
	var warehouses []models.Warehouse
	if err := query.
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&warehouses).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "inventory/get_warehouses", "Failed to get warehouses")
		return
	}

	// If stock summary is requested, calculate it for each warehouse
	if includeStock == "true" {
		var warehouseResponses []WarehouseResponse
		for _, warehouse := range warehouses {
			stockSummary := h.calculateStockSummary(warehouse.ID)
			warehouseResponses = append(warehouseResponses, WarehouseResponse{
				Warehouse:    warehouse,
				StockSummary: stockSummary,
			})
		}

		resp := map[string]interface{}{
			"data":      warehouseResponses,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		}
		response.GenerateSuccessResponse(c, "Warehouses retrieved successfully", resp)
		return
	}

	resp := map[string]interface{}{
		"data":      warehouses,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}
	response.GenerateSuccessResponse(c, "Warehouses retrieved successfully", resp)
}

// GetWarehouse - Admin endpoint to get a single warehouse with detailed stock information
func (h *InventoryHandler) GetWarehouse(c *gin.Context) {
	warehouseID := c.Param("id")
	if warehouseID == "" {
		response.GenerateBadRequestResponse(c, "inventory/get_warehouse", "Warehouse ID is required")
		return
	}

	var warehouse models.Warehouse
	if err := h.db.Preload("Address").First(&warehouse, warehouseID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.GenerateNotFoundResponse(c, "inventory/get_warehouse", "Warehouse not found")
		} else {
			response.GenerateInternalServerErrorResponse(c, "inventory/get_warehouse", "Failed to get warehouse")
		}
		return
	}

	// Calculate detailed stock summary
	stockSummary := h.calculateStockSummary(warehouse.ID)

	warehouseResponse := WarehouseResponse{
		Warehouse:    warehouse,
		StockSummary: stockSummary,
	}

	response.GenerateSuccessResponse(c, "Warehouse retrieved successfully", warehouseResponse)
}

// UpdateWarehouse - Admin endpoint to update warehouse details
func (h *InventoryHandler) UpdateWarehouse(c *gin.Context) {
	warehouseID := c.Param("id")
	if warehouseID == "" {
		response.GenerateBadRequestResponse(c, "inventory/update_warehouse", "Warehouse ID is required")
		return
	}

	var req UpdateWarehouseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GenerateBadRequestResponse(c, "inventory/update_warehouse", err.Error())
		return
	}

	// Get existing warehouse
	var warehouse models.Warehouse
	if err := h.db.First(&warehouse, warehouseID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.GenerateNotFoundResponse(c, "inventory/update_warehouse", "Warehouse not found")
		} else {
			response.GenerateInternalServerErrorResponse(c, "inventory/update_warehouse", "Failed to get warehouse")
		}
		return
	}

	// Check if new code already exists (if code is being updated)
	if req.Code != nil && *req.Code != warehouse.Code {
		var existingWarehouse models.Warehouse
		if err := h.db.Where("code = ? AND id != ?", *req.Code, warehouse.ID).First(&existingWarehouse).Error; err == nil {
			response.GenerateBadRequestResponse(c, "inventory/update_warehouse", "Warehouse code already exists")
			return
		}
	}

	// Verify new address exists (if address is being updated)
	if req.AddressID != nil {
		var address models.Address
		if err := h.db.First(&address, *req.AddressID).Error; err != nil {
			response.GenerateBadRequestResponse(c, "inventory/update_warehouse", "Address not found")
			return
		}
	}

	// Update fields
	if req.Name != nil {
		warehouse.Name = *req.Name
	}
	if req.Code != nil {
		warehouse.Code = *req.Code
	}
	if req.AddressID != nil {
		warehouse.AddressID = *req.AddressID
	}
	if req.IsActive != nil {
		warehouse.IsActive = *req.IsActive
	}

	if err := h.db.Save(&warehouse).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "inventory/update_warehouse", "Failed to update warehouse")
		return
	}

	// Load complete warehouse with address
	h.db.Preload("Address").First(&warehouse, warehouse.ID)

	response.GenerateSuccessResponse(c, "Warehouse updated successfully", warehouse)
}

// DeleteWarehouse - Admin endpoint to delete a warehouse (soft delete)
func (h *InventoryHandler) DeleteWarehouse(c *gin.Context) {
	warehouseID := c.Param("id")
	if warehouseID == "" {
		response.GenerateBadRequestResponse(c, "inventory/delete_warehouse", "Warehouse ID is required")
		return
	}

	// Check if warehouse has inventory items
	var inventoryCount int64
	h.db.Model(&models.InventoryItem{}).Where("warehouse_id = ?", warehouseID).Count(&inventoryCount)

	if inventoryCount > 0 {
		response.GenerateBadRequestResponse(c, "inventory/delete_warehouse", "Cannot delete warehouse with existing inventory items")
		return
	}

	if err := h.db.Delete(&models.Warehouse{}, warehouseID).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "inventory/delete_warehouse", "Failed to delete warehouse")
		return
	}

	response.GenerateSuccessResponse(c, "Warehouse deleted successfully", nil)
}

// Helper function to calculate stock summary for a warehouse
func (h *InventoryHandler) calculateStockSummary(warehouseID uint) StockSummary {
	var summary StockSummary

	// Total products (distinct variants)
	h.db.Model(&models.InventoryItem{}).
		Where("warehouse_id = ?", warehouseID).
		Distinct("product_variant_id").
		Count(&summary.TotalProducts)

	// Total quantity and reserved quantity
	h.db.Model(&models.InventoryItem{}).
		Where("warehouse_id = ?", warehouseID).
		Select("COALESCE(SUM(quantity), 0)").
		Row().Scan(&summary.TotalQuantity)

	h.db.Model(&models.InventoryItem{}).
		Where("warehouse_id = ?", warehouseID).
		Select("COALESCE(SUM(reserved), 0)").
		Row().Scan(&summary.ReservedQuantity)

	// Low stock items (quantity <= 10)
	h.db.Model(&models.InventoryItem{}).
		Where("warehouse_id = ? AND quantity <= ?", warehouseID, 10).
		Count(&summary.LowStockItems)

	// Out of stock items
	h.db.Model(&models.InventoryItem{}).
		Where("warehouse_id = ? AND quantity = ?", warehouseID, 0).
		Count(&summary.OutOfStockItems)

	// Total value calculation (quantity * cost_price)
	var totalValue interface{}
	h.db.Table("inventory_items").
		Select("COALESCE(SUM(inventory_items.quantity * product_variants.cost_price), 0)").
		Joins("JOIN product_variants ON product_variants.id = inventory_items.product_variant_id").
		Where("inventory_items.warehouse_id = ?", warehouseID).
		Row().Scan(&totalValue)

	if totalValue != nil {
		if val, ok := totalValue.(float64); ok {
			summary.TotalValue = val
		}
	}

	return summary
}
