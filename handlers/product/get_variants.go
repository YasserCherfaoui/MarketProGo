package product

import (
	"strconv"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

// ProductVariantWithQuantity extends the variant with quantity information
type ProductVariantWithQuantity struct {
	models.ProductVariant
	TotalStock      int    `json:"total_stock"`
	AvailableStock  int    `json:"available_stock"`
	ReservedStock   int    `json:"reserved_stock"`
	WarehousesCount int    `json:"warehouses_count"`
	StockStatus     string `json:"stock_status"`
}

func (h *ProductHandler) GetProductVariants(c *gin.Context) {
	// Query parameters
	productIDStr := c.Query("product_id")
	isActiveStr := c.Query("is_active")
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

	var variants []models.ProductVariant

	// Base query with preloads
	db := h.db.Model(&models.ProductVariant{}).
		Preload("Product").
		Preload("Product.Brand").
		Preload("Images").
		Preload("OptionValues").
		Preload("InventoryItems").
		Preload("InventoryItems.Warehouse").
		Preload("PriceTiers")

	// Always filter for active products by default
	db = db.Joins("JOIN products ON products.id = product_variants.product_id").
		Where("products.is_active = ?", true)

	// Apply filters
	if productIDStr != "" {
		if productID, err := strconv.Atoi(productIDStr); err == nil {
			db = db.Where("product_id = ?", productID)
		}
	}

	if isActiveStr != "" {
		if isActive, err := strconv.ParseBool(isActiveStr); err == nil {
			db = db.Where("product_variants.is_active = ?", isActive)
		}
	}

	if search != "" {
		// Search by product name, variant SKU/barcode (products table already joined)
		db = db.Where("products.name ILIKE ? OR product_variants.sku ILIKE ? OR product_variants.barcode ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// Get total count
	var total int64
	if err := db.Count(&total).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "product_variants/get_all", err.Error())
		return
	}

	// Apply pagination and fetch data
	offset := (page - 1) * pageSize
	if err := db.Offset(offset).Limit(pageSize).Find(&variants).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "product_variants/get_all", err.Error())
		return
	}

	// Transform variants to include quantity information
	var variantsWithQuantity []ProductVariantWithQuantity
	for _, variant := range variants {
		// Process image URLs using Appwrite service
		for j := range variant.Images {
			variant.Images[j].URL = h.appwriteService.GetFileURL(variant.Images[j].URL)
		}

		// Process brand image if exists
		if variant.Product.Brand != nil {
			variant.Product.Brand.Image = h.appwriteService.GetFileURL(variant.Product.Brand.Image)
		}

		// Calculate stock levels from inventory items
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

		availableStock := totalStock - reservedStock

		// Determine stock status
		stockStatus := "out_of_stock"
		if totalStock > 0 {
			if totalStock <= 10 {
				stockStatus = "low_stock"
			} else {
				stockStatus = "in_stock"
			}
		}

		// Create variant with quantity info
		variantWithQuantity := ProductVariantWithQuantity{
			ProductVariant:  variant,
			TotalStock:      totalStock,
			AvailableStock:  availableStock,
			ReservedStock:   reservedStock,
			WarehousesCount: warehousesCount,
			StockStatus:     stockStatus,
		}

		variantsWithQuantity = append(variantsWithQuantity, variantWithQuantity)
	}

	// Prepare paginated response
	resp := PaginatedResponse{
		Data:     variantsWithQuantity,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}

	response.GenerateSuccessResponse(c, "Product variants with quantities fetched successfully", resp)
}
