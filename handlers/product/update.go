package product

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	id := c.Param("id")

	// Find the product with images and categories
	var product models.Product
	if err := h.db.Preload("Images").Preload("Categories").First(&product, id).Error; err != nil {
		response.GenerateBadRequestResponse(c, "product/update", "Product not found")
		return
	}

	// Get images_to_delete from form (can be sent as comma-separated or as multiple form fields)
	imagesToDelete := c.PostFormArray("images_to_delete")
	if len(imagesToDelete) == 1 && strings.Contains(imagesToDelete[0], ",") {
		imagesToDelete = strings.Split(imagesToDelete[0], ",")
	}
	imagesToDeleteMap := make(map[string]bool)
	for _, url := range imagesToDelete {
		imagesToDeleteMap[strings.TrimSpace(url)] = true
	}

	// Remove only the images that are in images_to_delete
	var keptImages []models.ProductImage
	for _, img := range product.Images {
		if imagesToDeleteMap[img.URL] {
			// Extract object name from URL
			parts := strings.SplitN(img.URL, "/", 5)
			if len(parts) == 5 {
				objectName := parts[4]
				h.gcsService.DeleteFile(c.Request.Context(), objectName)
			}
			// Delete from DB
			h.db.Delete(&img)
		} else {
			keptImages = append(keptImages, img)
		}
	}
	product.Images = keptImages

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

	product.Name = name
	product.Description = description
	product.SKU = sku
	product.Barcode = barcode
	product.BasePrice = basePriceF
	product.B2BPrice = b2bPriceF
	product.CostPrice = costPriceF
	product.Weight = weightF
	product.WeightUnit = weightUnit

	// Handle new image files
	form, err := c.MultipartForm()
	if err != nil {
		response.GenerateBadRequestResponse(c, "product/update", "Invalid multipart form data")
		return
	}
	files := form.File["images"]
	for idx, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			response.GenerateBadRequestResponse(c, "product/update", "Failed to open uploaded image")
			return
		}
		defer file.Close()
		objectName := fmt.Sprintf("products/%s_%d_%d%s", sku, time.Now().UnixNano(), idx, filepath.Ext(fileHeader.Filename))
		attrs, err := h.gcsService.UploadFile(c.Request.Context(), file, objectName, fileHeader.Header.Get("Content-Type"))
		if err != nil {
			response.GenerateInternalServerErrorResponse(c, "product/update", fmt.Sprintf("Failed to upload image to GCS: %v", err))
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

	// Save product
	if err := tx.Save(&product).Error; err != nil {
		tx.Rollback()
		response.GenerateInternalServerErrorResponse(c, "product/update", "Failed to update product")
		return
	}

	// Associate categories
	if len(categoryIDs) > 0 {
		var categories []models.Category
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
		response.GenerateInternalServerErrorResponse(c, "product/update", "Failed to commit transaction")
		return
	}

	response.GenerateSuccessResponse(c, "Product updated successfully", product)
}
