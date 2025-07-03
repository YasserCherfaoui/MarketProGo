package inventory

import (
	"strconv"
	"time"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

type MultiWarehouseStockResponse struct {
	ProductVariant ProductVariantInfo `json:"product_variant"`
	WarehouseStock []WarehouseStock   `json:"warehouse_stock"`
	TotalStock     int                `json:"total_stock"`
	TotalAvailable int                `json:"total_available"`
	TotalReserved  int                `json:"total_reserved"`
}

type ProductVariantInfo struct {
	ID      uint        `json:"id"`
	Name    string      `json:"name"`
	SKU     string      `json:"sku"`
	Product ProductInfo `json:"product"`
}

type ProductInfo struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type WarehouseStock struct {
	WarehouseID   uint        `json:"warehouse_id"`
	WarehouseName string      `json:"warehouse_name"`
	WarehouseCode string      `json:"warehouse_code"`
	Quantity      int         `json:"quantity"`
	Reserved      int         `json:"reserved"`
	Available     int         `json:"available"`
	Batches       []BatchInfo `json:"batches"`
}

type BatchInfo struct {
	BatchNumber string     `json:"batch_number"`
	Quantity    int        `json:"quantity"`
	ExpiryDate  *time.Time `json:"expiry_date"`
}

// GetMultiWarehouseStock - Show stock levels across all warehouses for a specific product variant
func (h *InventoryHandler) GetMultiWarehouseStock(c *gin.Context) {
	productVariantIDStr := c.Param("product_variant_id")
	productVariantID, err := strconv.ParseUint(productVariantIDStr, 10, 32)
	if err != nil {
		response.GenerateBadRequestResponse(c, "invalid_product_variant_id", "Invalid product variant ID")
		return
	}

	// Get product variant with product info
	var productVariant models.ProductVariant
	if err := h.db.Preload("Product").First(&productVariant, uint(productVariantID)).Error; err != nil {
		response.GenerateNotFoundResponse(c, "product_variant_not_found", "Product variant not found")
		return
	}

	// Get all inventory items for this variant across warehouses
	var inventoryItems []models.InventoryItem
	if err := h.db.Where("product_variant_id = ?", productVariantID).
		Preload("Warehouse").
		Find(&inventoryItems).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "inventory/stock/by-product", err.Error())
		return
	}

	// Group by warehouse
	warehouseStockMap := make(map[uint]*WarehouseStock)
	totalStock := 0
	totalReserved := 0

	for _, item := range inventoryItems {
		warehouseID := item.WarehouseID

		if warehouseStockMap[warehouseID] == nil {
			warehouseStockMap[warehouseID] = &WarehouseStock{
				WarehouseID:   item.Warehouse.ID,
				WarehouseName: item.Warehouse.Name,
				WarehouseCode: item.Warehouse.Code,
				Quantity:      0,
				Reserved:      0,
				Available:     0,
				Batches:       []BatchInfo{},
			}
		}

		// Add to warehouse totals
		warehouseStock := warehouseStockMap[warehouseID]
		warehouseStock.Quantity += item.Quantity
		warehouseStock.Reserved += item.Reserved
		warehouseStock.Available = warehouseStock.Quantity - warehouseStock.Reserved

		// Add batch info if available
		if item.BatchNumber != "" {
			batch := BatchInfo{
				BatchNumber: item.BatchNumber,
				Quantity:    item.Quantity,
				ExpiryDate:  item.ExpiryDate,
			}
			warehouseStock.Batches = append(warehouseStock.Batches, batch)
		}

		// Add to grand totals
		totalStock += item.Quantity
		totalReserved += item.Reserved
	}

	// Convert map to slice
	var warehouseStocks []WarehouseStock
	for _, stock := range warehouseStockMap {
		warehouseStocks = append(warehouseStocks, *stock)
	}

	// Prepare response
	stockResponse := MultiWarehouseStockResponse{
		ProductVariant: ProductVariantInfo{
			ID:   productVariant.ID,
			Name: productVariant.Name,
			SKU:  productVariant.SKU,
			Product: ProductInfo{
				ID:   productVariant.Product.ID,
				Name: productVariant.Product.Name,
			},
		},
		WarehouseStock: warehouseStocks,
		TotalStock:     totalStock,
		TotalAvailable: totalStock - totalReserved,
		TotalReserved:  totalReserved,
	}

	response.GenerateSuccessResponse(c, "Product variant stock retrieved successfully", stockResponse)
}
