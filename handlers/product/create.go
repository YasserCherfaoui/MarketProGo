package product

import (
	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

type CreateProductRequest struct {
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	SKU         string   `json:"sku" binding:"required"`
	Barcode     string   `json:"barcode"`
	BasePrice   float64  `json:"base_price" binding:"required"`
	B2BPrice    float64  `json:"b2b_price"`
	CostPrice   float64  `json:"cost_price"`
	Weight      float64  `json:"weight"`
	WeightUnit  string   `json:"weight_unit"`
	CategoryIDs []uint   `json:"category_ids"`
	Images      []string `json:"images"`
}

func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.GenerateBadRequestResponse(c, "product/create", err.Error())
		return
	}

	product := models.Product{
		Name:        req.Name,
		Description: req.Description,
		SKU:         req.SKU,
		Barcode:     req.Barcode,
		BasePrice:   req.BasePrice,
		B2BPrice:    req.B2BPrice,
		CostPrice:   req.CostPrice,
		Weight:      req.Weight,
		WeightUnit:  req.WeightUnit,
		IsActive:    true,
	}
	// Create images
	for _, image := range req.Images {
		image := models.ProductImage{
			URL: image,
		}
		product.Images = append(product.Images, image)
	}

	// Start transaction
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create product
	if err := tx.Create(&product).Error; err != nil {
		tx.Rollback()
		response.GenerateInternalServerErrorResponse(c, "product/create", "Failed to create product")
		return
	}

	// Associate categories
	if len(req.CategoryIDs) > 0 {
		var categories []models.Category
		if err := tx.Find(&categories, req.CategoryIDs).Error; err != nil {
			tx.Rollback()
			response.GenerateInternalServerErrorResponse(c, "category/get", "Category not found")
			return
		}
		if err := tx.Model(&product).Association("Categories").Replace(categories); err != nil {
			tx.Rollback()
			response.GenerateInternalServerErrorResponse(c, "category/associate", "Failed to associate categories")
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "product/create", "Failed to commit transaction")
		return
	}

	response.GenerateSuccessResponse(c, "Product created successfully", product)
}
