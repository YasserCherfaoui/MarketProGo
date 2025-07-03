package inventory

import (
	"strconv"
	"time"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

type StockMovementResponse struct {
	ID             uint                   `json:"id"`
	InventoryItem  MovementInventoryItem  `json:"inventory_item"`
	ProductVariant MovementProductVariant `json:"product_variant"`
	Warehouse      MovementWarehouse      `json:"warehouse"`
	MovementType   string                 `json:"movement_type"`
	Quantity       int                    `json:"quantity"`
	Reason         string                 `json:"reason"`
	Notes          string                 `json:"notes"`
	Reference      string                 `json:"reference"`
	User           *MovementUser          `json:"user,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
}

type MovementInventoryItem struct {
	ID uint `json:"id"`
}

type MovementProductVariant struct {
	ID      uint            `json:"id"`
	Name    string          `json:"name"`
	SKU     string          `json:"sku"`
	Product MovementProduct `json:"product"`
}

type MovementProduct struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type MovementWarehouse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

type MovementUser struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// GetStockMovements - Browse stock movement history with advanced filtering
func (h *InventoryHandler) GetStockMovements(c *gin.Context) {
	// Query parameters
	inventoryItemID := c.Query("inventory_item_id")
	productVariantID := c.Query("product_variant_id")
	warehouseID := c.Query("warehouse_id")
	movementType := c.Query("movement_type")
	userID := c.Query("user_id")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")

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

	var movements []models.StockMovement

	// Base query
	db := h.db.Model(&models.StockMovement{}).
		Preload("InventoryItem.ProductVariant.Product").
		Preload("InventoryItem.Warehouse").
		Preload("User").
		Order("created_at DESC")

	// Apply filters
	if inventoryItemID != "" {
		db = db.Where("inventory_item_id = ?", inventoryItemID)
	}

	if productVariantID != "" {
		db = db.Joins("JOIN inventory_items ON inventory_items.id = stock_movements.inventory_item_id").
			Where("inventory_items.product_variant_id = ?", productVariantID)
	}

	if warehouseID != "" {
		db = db.Joins("JOIN inventory_items ON inventory_items.id = stock_movements.inventory_item_id").
			Where("inventory_items.warehouse_id = ?", warehouseID)
	}

	if movementType != "" {
		db = db.Where("movement_type = ?", movementType)
	}

	if userID != "" {
		db = db.Where("user_id = ?", userID)
	}

	// Date range filters
	if dateFrom != "" {
		if parsedDate, err := time.Parse("2006-01-02", dateFrom); err == nil {
			db = db.Where("created_at >= ?", parsedDate)
		}
	}

	if dateTo != "" {
		if parsedDate, err := time.Parse("2006-01-02", dateTo); err == nil {
			// Add 1 day to include the entire day
			endDate := parsedDate.AddDate(0, 0, 1)
			db = db.Where("created_at < ?", endDate)
		}
	}

	// Get total count
	var total int64
	if err := db.Count(&total).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "inventory/movements", err.Error())
		return
	}

	// Apply pagination and fetch data
	offset := (page - 1) * pageSize
	if err := db.Offset(offset).Limit(pageSize).Find(&movements).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "inventory/movements", err.Error())
		return
	}

	// Transform data to response format
	var movementResponses []StockMovementResponse
	for _, movement := range movements {
		movementResp := StockMovementResponse{
			ID: movement.ID,
			InventoryItem: MovementInventoryItem{
				ID: movement.InventoryItem.ID,
			},
			ProductVariant: MovementProductVariant{
				ID:   movement.InventoryItem.ProductVariant.ID,
				Name: movement.InventoryItem.ProductVariant.Name,
				SKU:  movement.InventoryItem.ProductVariant.SKU,
				Product: MovementProduct{
					ID:   movement.InventoryItem.ProductVariant.Product.ID,
					Name: movement.InventoryItem.ProductVariant.Product.Name,
				},
			},
			Warehouse: MovementWarehouse{
				ID:   movement.InventoryItem.Warehouse.ID,
				Name: movement.InventoryItem.Warehouse.Name,
				Code: movement.InventoryItem.Warehouse.Code,
			},
			MovementType: movement.MovementType,
			Quantity:     movement.Quantity,
			Reason:       movement.Reason,
			Notes:        movement.Notes,
			Reference:    movement.Reference,
			CreatedAt:    movement.CreatedAt,
		}

		// Add user info if available
		if movement.User != nil {
			movementResp.User = &MovementUser{
				ID:   movement.User.ID,
				Name: movement.User.FirstName + " " + movement.User.LastName,
			}
		}

		movementResponses = append(movementResponses, movementResp)
	}

	// Prepare paginated response
	resp := PaginatedResponse{
		Data:     movementResponses,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}

	response.GenerateSuccessResponse(c, "Stock movements retrieved successfully", resp)
}

// GetStockMovement - Get a specific stock movement by ID
func (h *InventoryHandler) GetStockMovement(c *gin.Context) {
	movementIDStr := c.Param("id")
	movementID, err := strconv.ParseUint(movementIDStr, 10, 32)
	if err != nil {
		response.GenerateBadRequestResponse(c, "invalid_movement_id", "Invalid movement ID")
		return
	}

	var movement models.StockMovement
	if err := h.db.Preload("InventoryItem.ProductVariant.Product").
		Preload("InventoryItem.Warehouse").
		Preload("User").
		First(&movement, uint(movementID)).Error; err != nil {
		response.GenerateNotFoundResponse(c, "movement_not_found", "Stock movement not found")
		return
	}

	// Transform to response format
	movementResp := StockMovementResponse{
		ID: movement.ID,
		InventoryItem: MovementInventoryItem{
			ID: movement.InventoryItem.ID,
		},
		ProductVariant: MovementProductVariant{
			ID:   movement.InventoryItem.ProductVariant.ID,
			Name: movement.InventoryItem.ProductVariant.Name,
			SKU:  movement.InventoryItem.ProductVariant.SKU,
			Product: MovementProduct{
				ID:   movement.InventoryItem.ProductVariant.Product.ID,
				Name: movement.InventoryItem.ProductVariant.Product.Name,
			},
		},
		Warehouse: MovementWarehouse{
			ID:   movement.InventoryItem.Warehouse.ID,
			Name: movement.InventoryItem.Warehouse.Name,
			Code: movement.InventoryItem.Warehouse.Code,
		},
		MovementType: movement.MovementType,
		Quantity:     movement.Quantity,
		Reason:       movement.Reason,
		Notes:        movement.Notes,
		Reference:    movement.Reference,
		CreatedAt:    movement.CreatedAt,
	}

	// Add user info if available
	if movement.User != nil {
		movementResp.User = &MovementUser{
			ID:   movement.User.ID,
			Name: movement.User.FirstName + " " + movement.User.LastName,
		}
	}

	response.GenerateSuccessResponse(c, "Stock movement retrieved successfully", movementResp)
}
