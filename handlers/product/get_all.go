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
	tag := c.Query("tag")
	brandSlug := c.Query("brand_slug")
	priceType := c.DefaultQuery("price_type", "customer") // customer or business
	sortByPrice := c.Query("sort_by_price")               // asc or desc

	var products []models.Product

	// Base query with all preloads
	db := h.db.Model(&models.Product{}).
		Preload("Brand").
		Preload("Categories").
		Preload("Tags").
		Preload("Images").
		Preload("Options.Values").
		Preload("Variants.Images").
		Preload("Variants.OptionValues").
		Preload("Specifications")

	// Use a subquery for filtering to handle variants correctly
	subQuery := h.db.Model(&models.Product{}).Select("DISTINCT products.id")

	// Apply filters that require joins
	requiresVariantJoin := sku != "" || barcode != "" || minPrice != "" || maxPrice != "" || sortByPrice != ""
	if requiresVariantJoin {
		subQuery = subQuery.Joins("JOIN product_variants ON product_variants.product_id = products.id")
	}
	if categoryID != "" {
		subQuery = subQuery.Joins("JOIN product_categories ON product_categories.product_id = products.id")
	}
	if tag != "" {
		subQuery = subQuery.Joins("JOIN product_tags ON product_tags.product_id = products.id").
			Joins("JOIN tags ON tags.id = product_tags.tag_id")
	}
	if brandSlug != "" {
		subQuery = subQuery.Joins("JOIN brands ON brands.id = products.brand_id")
	}

	// Apply filtering conditions
	if name != "" {
		subQuery = subQuery.Where("products.name ILIKE ?", "%"+name+"%")
	}
	if isActive != "" {
		subQuery = subQuery.Where("products.is_active = ?", isActive == "true")
	}
	if isFeatured != "" {
		subQuery = subQuery.Where("products.is_featured = ?", isFeatured == "true")
	}
	if categoryID != "" {
		subQuery = subQuery.Where("product_categories.category_id = ?", categoryID)
	}
	if tag != "" {
		subQuery = subQuery.Where("tags.name ILIKE ?", "%"+tag+"%")
	}
	if brandSlug != "" {
		subQuery = subQuery.Where("brands.slug = ?", brandSlug)
	}

	// Variant-specific filters
	if sku != "" {
		subQuery = subQuery.Where("product_variants.sku ILIKE ?", "%"+sku+"%")
	}
	if barcode != "" {
		subQuery = subQuery.Where("product_variants.barcode = ?", barcode)
	}
	priceField := "product_variants.base_price"
	if priceType == "business" {
		priceField = "product_variants.b2b_price"
	}
	if minPrice != "" {
		subQuery = subQuery.Where(priceField+" >= ?", minPrice)
	}
	if maxPrice != "" {
		subQuery = subQuery.Where(priceField+" <= ?", maxPrice)
	}

	if sortByPrice != "" && (sortByPrice == "asc" || sortByPrice == "desc") {
		// Ordering needs to be on the outer query
		db = db.Order("id ASC") // Default order
	}

	// Pagination logic
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

	// Get total count based on the filtered subquery
	var total int64
	h.db.Table("(?) as sub", subQuery).Count(&total)

	// Get the IDs for the current page
	var productIDs []uint
	subQuery.Offset((page-1)*pageSize).Limit(pageSize).Pluck("id", &productIDs)

	// Fetch the full product data for the paginated IDs
	if len(productIDs) > 0 {
		if err := db.Where("id IN ?", productIDs).Find(&products).Error; err != nil {
			response.GenerateInternalServerErrorResponse(c, "product/get_all", err.Error())
			return
		}
	}

	// Add Appwrite URLs to product and brand images
	for i := range products {
		if products[i].Brand != nil {
			products[i].Brand.Image = h.appwriteService.GetFileURL(products[i].Brand.Image)
		}
		for j := range products[i].Images {
			products[i].Images[j].URL = h.appwriteService.GetFileURL(products[i].Images[j].URL)
		}
		for j := range products[i].Variants {
			for k := range products[i].Variants[j].Images {
				products[i].Variants[j].Images[k].URL = h.appwriteService.GetFileURL(products[i].Variants[j].Images[k].URL)
			}
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
