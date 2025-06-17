package product

import (
	"fmt"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

// PaginatedResponse is the struct for paginated API responses
// Use this as the data field in the response.GenerateSuccessResponse
type PaginatedResponse struct {
	Data     interface{} `json:"data"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

func (h *ProductHandler) GetAllProducts(c *gin.Context) {
	// Query params
	name := c.Query("name")
	sku := c.Query("sku")
	barcode := c.Query("barcode")
	isActive := c.Query("is_active")
	isFeatured := c.Query("is_featured")
	minPrice := c.Query("min_price")
	maxPrice := c.Query("max_price")
	categoryID := c.Query("category_id")
	priceType := c.DefaultQuery("price_type", "customer") // customer or business
	sortByPrice := c.Query("sort_by_price")               // asc or desc

	products := []models.Product{}

	// Determine which price field to use
	priceField := "base_price"
	if priceType == "business" {
		priceField = "b2b_price"
	}

	db := h.db.Model(&models.Product{}).Preload("Categories").Preload("Images")

	if name != "" {
		db = db.Where("name ILIKE ?", "%"+name+"%")
	}
	if sku != "" {
		db = db.Where("sku ILIKE ?", "%"+sku+"%")
	}
	if barcode != "" {
		db = db.Where("barcode = ?", barcode)
	}
	if isActive != "" {
		if isActive == "true" {
			db = db.Where("is_active = ?", true)
		} else if isActive == "false" {
			db = db.Where("is_active = ?", false)
		}
	}
	if isFeatured != "" {
		if isFeatured == "true" {
			db = db.Where("is_featured = ?", true)
		} else if isFeatured == "false" {
			db = db.Where("is_featured = ?", false)
		}
	}
	if minPrice != "" {
		db = db.Where(priceField+" >= ?", minPrice)
	}
	if maxPrice != "" {
		db = db.Where(priceField+" <= ?", maxPrice)
	}
	if categoryID != "" {
		db = db.Joins("JOIN product_categories ON product_categories.product_id = products.id").Where("product_categories.category_id = ?", categoryID)
	}
	if sortByPrice == "asc" {
		db = db.Order(priceField + " ASC")
	} else if sortByPrice == "desc" {
		db = db.Order(priceField + " DESC")
	}

	page := 1
	pageSize := 20
	if p := c.Query("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
		if page < 1 {
			page = 1
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		fmt.Sscanf(ps, "%d", &pageSize)
		if pageSize < 1 {
			pageSize = 20
		} else if pageSize > 100 {
			pageSize = 100
		}
	}

	offset := (page - 1) * pageSize

	var total int64
	db.Count(&total)
	db = db.Offset(offset).Limit(pageSize)

	if err := db.Find(&products).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "product/get_all", err.Error())
		return
	}

	// Add Appwrite URLs to product images
	for i := range products {
		for j := range products[i].Images {
			products[i].Images[j].URL = h.appwriteService.GetFileURL(products[i].Images[j].URL)
		}
	}

	resp := PaginatedResponse{
		Data:     products,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
	response.GenerateSuccessResponse(c, "Products fetched successfully", resp)
}
