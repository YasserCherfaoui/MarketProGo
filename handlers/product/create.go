package product

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

func (h *ProductHandler) CreateProduct(c *gin.Context) {
	// Parse form fields
	name := c.PostForm("name")
	description := c.PostForm("description")
	sku := c.PostForm("sku")
	barcode := c.PostForm("barcode")
	basePrice := c.PostForm("base_price")
	b2bPrice := c.PostForm("b2b_price")
	costPrice := c.PostForm("cost_price")
	weight := c.PostForm("weight")
	weightUnit := c.PostForm("weight_unit")
	categoryIDs := c.PostFormArray("category_ids")

	// Convert numeric fields
	var basePriceF, b2bPriceF, costPriceF, weightF float64
	fmt.Sscanf(basePrice, "%f", &basePriceF)
	fmt.Sscanf(b2bPrice, "%f", &b2bPriceF)
	fmt.Sscanf(costPrice, "%f", &costPriceF)
	fmt.Sscanf(weight, "%f", &weightF)

	product := models.Product{
		Name:        name,
		Description: description,
		SKU:         sku,
		Barcode:     barcode,
		BasePrice:   basePriceF,
		B2BPrice:    b2bPriceF,
		CostPrice:   costPriceF,
		Weight:      weightF,
		WeightUnit:  weightUnit,
		IsActive:    true,
	}

	// Handle image files
	form, err := c.MultipartForm()
	if err != nil {
		response.GenerateBadRequestResponse(c, "product/create", "Invalid multipart form data")
		return
	}
	files := form.File["images"]
	for idx, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			response.GenerateBadRequestResponse(c, "product/create", "Failed to open uploaded image")
			return
		}
		defer file.Close()
		objectName := fmt.Sprintf("products/%s_%d_%d%s", sku, time.Now().UnixNano(), idx, filepath.Ext(fileHeader.Filename))
		attrs, err := h.gcsService.UploadFile(c.Request.Context(), file, objectName, fileHeader.Header.Get("Content-Type"))
		if err != nil {
			response.GenerateInternalServerErrorResponse(c, "product/create", fmt.Sprintf("Failed to upload image to GCS: %v", err))
			return
		}
		image := models.ProductImage{
			URL: fmt.Sprintf("https://storage.googleapis.com/%s/%s", attrs.Bucket, attrs.Name),
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
	if len(categoryIDs) > 0 {
		var categories []models.Category
		// Convert categoryIDs to uint
		var ids []uint
		for _, idStr := range categoryIDs {
			var id uint
			fmt.Sscanf(idStr, "%d", &id)
			ids = append(ids, id)
		}
		if err := tx.Find(&categories, ids).Error; err != nil {
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
