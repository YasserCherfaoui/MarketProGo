package inventory

import (
	"strconv"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

type ProductInventoryOverview struct {
	ProductID uint                      `json:"product_id"`
	Name      string                    `json:"name"`
	Brand     string                    `json:"brand"`
	Category  string                    `json:"category"`
	Variants  []VariantInventorySummary `json:"variants"`
}

type VariantInventorySummary struct {
	VariantID       uint   `json:"variant_id"`
	Name            string `json:"name"`
	SKU             string `json:"sku"`
	TotalStock      int    `json:"total_stock"`
	AvailableStock  int    `json:"available_stock"`
	ReservedStock   int    `json:"reserved_stock"`
	WarehousesCount int    `json:"warehouses_count"`
	StockStatus     string `json:"stock_status"`
}

type PaginatedResponse struct {
	Data     interface{} `json:"data"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// GetProductInventoryOverview - Browse all products with their inventory summary
func (h *InventoryHandler) GetProductInventoryOverview(c *gin.Context) {
	// Query parameters
	categoryID := c.Query("category_id")
	brandID := c.Query("brand_id")
	isActiveStr := c.Query("is_active")
	hasStockStr := c.Query("has_stock")
	search := c.Query("search")

	// Pagination parameters
	page := 1
	pageSize := 20
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

	var products []models.Product

	// Base query
	db := h.db.Model(&models.Product{}).
		Preload("Brand").
		Preload("Categories").
		Preload("Variants.InventoryItems.Warehouse")

	// Apply filters
	if categoryID != "" {
		db = db.Joins("JOIN product_categories ON product_categories.product_id = products.id").
			Where("product_categories.category_id = ?", categoryID)
	}

	if brandID != "" {
		db = db.Where("brand_id = ?", brandID)
	}

	if isActiveStr != "" {
		if isActive, err := strconv.ParseBool(isActiveStr); err == nil {
			db = db.Where("is_active = ?", isActive)
		}
	}

	if search != "" {
		db = db.Where("name ILIKE ?", "%"+search+"%")
	}

	// Filter products with stock if requested
	if hasStockStr == "true" {
		db = db.Where("EXISTS (SELECT 1 FROM product_variants pv JOIN inventory_items ii ON ii.product_variant_id = pv.id WHERE pv.product_id = products.id AND ii.quantity > 0)")
	}

	// Get total count
	var total int64
	if err := db.Count(&total).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "inventory/products", err.Error())
		return
	}

	// Apply pagination and fetch data
	offset := (page - 1) * pageSize
	if err := db.Offset(offset).Limit(pageSize).Find(&products).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "inventory/products", err.Error())
		return
	}

	// Transform data to response format
	var productOverviews []ProductInventoryOverview
	for _, product := range products {
		overview := ProductInventoryOverview{
			ProductID: product.ID,
			Name:      product.Name,
		}

		// Set brand name
		if product.Brand != nil {
			overview.Brand = product.Brand.Name
		}

		// Set category name (first category if multiple)
		if len(product.Categories) > 0 {
			overview.Category = product.Categories[0].Name
		}

		// Process variants
		for _, variant := range product.Variants {
			variantSummary := VariantInventorySummary{
				VariantID: variant.ID,
				Name:      variant.Name,
				SKU:       variant.SKU,
			}

			// Calculate stock levels
			totalStock := 0
			reservedStock := 0
			warehousesCount := 0
			warehouseSet := make(map[uint]bool)

			for _, inventoryItem := range variant.InventoryItems {
				totalStock += inventoryItem.Quantity
				reservedStock += inventoryItem.Reserved
				if !warehouseSet[inventoryItem.WarehouseID] {
					warehouseSet[inventoryItem.WarehouseID] = true
					warehousesCount++
				}
			}

			variantSummary.TotalStock = totalStock
			variantSummary.ReservedStock = reservedStock
			variantSummary.AvailableStock = totalStock - reservedStock
			variantSummary.WarehousesCount = warehousesCount

			// Determine stock status
			if totalStock == 0 {
				variantSummary.StockStatus = "out_of_stock"
			} else if totalStock <= 10 {
				variantSummary.StockStatus = "low_stock"
			} else {
				variantSummary.StockStatus = "in_stock"
			}

			overview.Variants = append(overview.Variants, variantSummary)
		}

		productOverviews = append(productOverviews, overview)
	}

	// Prepare paginated response
	resp := PaginatedResponse{
		Data:     productOverviews,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}

	response.GenerateSuccessResponse(c, "Products with inventory retrieved successfully", resp)
}
