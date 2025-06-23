package product

import (
	"encoding/json"
	"strconv"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

// Re-using request structs from create.go for simplicity
// A dedicated UpdateProductData with pointers would be ideal for partial updates,
// but this approach uses separate arrays for explicit actions (add, update, delete).

type UpdateProductData struct {
	Name             *string                `json:"name"`
	Description      *string                `json:"description"`
	IsActive         *bool                  `json:"is_active"`
	IsFeatured       *bool                  `json:"is_featured"`
	CategoryIDs      []uint                 `json:"category_ids"`
	Tags             []string               `json:"tags"`
	ImagesToAdd      []ImageData            `json:"images_to_add"`
	Specifications   []SpecificationRequest `json:"specifications_to_add"`
	OptionsToAdd     []OptionData           `json:"options_to_add"`
	VariantsToAdd    []VariantData          `json:"variants_to_add"`
	VariantsToUpdate []VariantUpdateData    `json:"variants_to_update"`
	VariantsToDelete []uint                 `json:"variants_to_delete"`
	// Note: Image updates are handled via file upload and 'images_to_delete' form field
}

type VariantUpdateData struct {
	ID         uint               `json:"id"`
	Name       *string            `json:"name"`
	SKU        *string            `json:"sku"`
	Barcode    *string            `json:"barcode"`
	BasePrice  *float64           `json:"base_price"`
	B2BPrice   *float64           `json:"b2b_price"`
	CostPrice  *float64           `json:"cost_price"`
	Weight     *float64           `json:"weight"`
	WeightUnit *string            `json:"weight_unit"`
	Dimensions *models.Dimensions `json:"dimensions"`
	IsActive   *bool              `json:"is_active"`
}

func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	productID := c.Param("id")

	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var product models.Product
	if err := tx.Preload("Categories").Preload("Tags").Preload("Images").Preload("Variants.Images").Preload("Options.Values").First(&product, productID).Error; err != nil {
		response.GenerateNotFoundResponse(c, "product/update", "Product not found")
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		response.GenerateBadRequestResponse(c, "product/update", "Invalid multipart form data: "+err.Error())
		return
	}

	// Handle Image Deletion
	imagesToDeleteIDs := form.Value["images_to_delete"]
	if len(imagesToDeleteIDs) > 0 {
		for _, idStr := range imagesToDeleteIDs {
			id, _ := strconv.Atoi(idStr)
			if err := tx.Delete(&models.ProductImage{}, id).Error; err != nil {
				tx.Rollback()
				response.GenerateInternalServerErrorResponse(c, "product/update", "Failed to delete image")
				return
			}
		}
	}

	// Handle New Image Uploads
	files := form.File["files"]
	uploadedFileIDs := make(map[string]string)
	for _, fileHeader := range files {
		fileID, err := h.appwriteService.UploadFile(fileHeader)
		if err != nil {
			tx.Rollback()
			response.GenerateInternalServerErrorResponse(c, "product/update", "Failed to upload image '"+fileHeader.Filename+"'")
			return
		}
		uploadedFileIDs[fileHeader.Filename] = fileID
	}

	// Process JSON data for other updates
	productDataJSON, hasData := form.Value["product_data"]
	if hasData && len(productDataJSON) > 0 {
		var data UpdateProductData
		if err := json.Unmarshal([]byte(productDataJSON[0]), &data); err != nil {
			tx.Rollback()
			response.GenerateBadRequestResponse(c, "product/update", "Invalid JSON in 'product_data' field: "+err.Error())
			return
		}

		// Handle Images to Add
		for _, imgData := range data.ImagesToAdd {
			fileID, ok := uploadedFileIDs[imgData.FileName]
			if !ok {
				tx.Rollback()
				response.GenerateBadRequestResponse(c, "product/update", "Image file '"+imgData.FileName+"' was specified but not found in upload")
				return
			}
			image := models.ProductImage{
				ProductID: &product.ID,
				URL:       fileID,
				IsPrimary: imgData.IsPrimary,
				AltText:   imgData.AltText,
			}
			if err := tx.Create(&image).Error; err != nil {
				tx.Rollback()
				response.GenerateInternalServerErrorResponse(c, "product/update", "Failed to save new image")
				return
			}
		}

		// Update base product fields if provided
		if data.Name != nil {
			product.Name = *data.Name
		}
		if data.Description != nil {
			product.Description = *data.Description
		}
		if data.IsActive != nil {
			product.IsActive = *data.IsActive
		}
		if data.IsFeatured != nil {
			product.IsFeatured = *data.IsFeatured
		}

		// Replace Categories
		if data.CategoryIDs != nil {
			var categories []models.Category
			if err := tx.Find(&categories, data.CategoryIDs).Error; err != nil {
				tx.Rollback()
				response.GenerateInternalServerErrorResponse(c, "product/update", "Failed to find categories")
				return
			}
			if err := tx.Model(&product).Association("Categories").Replace(categories); err != nil {
				tx.Rollback()
				response.GenerateInternalServerErrorResponse(c, "product/update", "Failed to update categories")
				return
			}
		}

		// Replace Tags
		if data.Tags != nil {
			var tags []*models.Tag
			for _, tagName := range data.Tags {
				tag := models.Tag{Name: tagName}
				tx.Where(models.Tag{Name: tagName}).FirstOrCreate(&tag)
				tags = append(tags, &tag)
			}
			if err := tx.Model(&product).Association("Tags").Replace(tags); err != nil {
				tx.Rollback()
				response.GenerateInternalServerErrorResponse(c, "product/update", "Failed to update tags")
				return
			}
		}

		// Handle Variant Deletion
		if len(data.VariantsToDelete) > 0 {
			for _, variantID := range data.VariantsToDelete {
				// Also delete variant images and associations
				tx.Where("product_variant_id = ?", variantID).Delete(&models.ProductImage{})
				tx.Exec("DELETE FROM variant_option_values WHERE product_variant_id = ?", variantID)
				if err := tx.Delete(&models.ProductVariant{}, variantID).Error; err != nil {
					tx.Rollback()
					response.GenerateInternalServerErrorResponse(c, "product/update", "Failed to delete variant")
					return
				}
			}
		}

		// Handle Variant Updates
		for _, varUpdateData := range data.VariantsToUpdate {
			var variant models.ProductVariant
			if err := tx.First(&variant, varUpdateData.ID).Error; err != nil {
				tx.Rollback()
				response.GenerateBadRequestResponse(c, "product/update", "Variant with ID "+strconv.Itoa(int(varUpdateData.ID))+" not found.")
				return
			}
			if varUpdateData.Name != nil {
				variant.Name = *varUpdateData.Name
			}
			if varUpdateData.SKU != nil {
				variant.SKU = *varUpdateData.SKU
			}
			if varUpdateData.Barcode != nil {
				variant.Barcode = *varUpdateData.Barcode
			}
			if varUpdateData.BasePrice != nil {
				variant.BasePrice = *varUpdateData.BasePrice
			}
			if varUpdateData.B2BPrice != nil {
				variant.B2BPrice = *varUpdateData.B2BPrice
			}
			if varUpdateData.CostPrice != nil {
				variant.CostPrice = *varUpdateData.CostPrice
			}
			if varUpdateData.Weight != nil {
				variant.Weight = *varUpdateData.Weight
			}
			if varUpdateData.WeightUnit != nil {
				variant.WeightUnit = *varUpdateData.WeightUnit
			}
			if varUpdateData.Dimensions != nil {
				variant.Dimensions = varUpdateData.Dimensions
			}
			if varUpdateData.IsActive != nil {
				variant.IsActive = *varUpdateData.IsActive
			}
			if err := tx.Save(&variant).Error; err != nil {
				tx.Rollback()
				response.GenerateInternalServerErrorResponse(c, "product/update", "Failed to update variant")
				return
			}
		}
		// NOTE: A more complex implementation would be needed to add/remove/update options
		// and associate new images with existing variants. This is a simplified version.
	}

	// Save changes to the base product
	if err := tx.Save(&product).Error; err != nil {
		tx.Rollback()
		response.GenerateInternalServerErrorResponse(c, "product/update", "Failed to save product")
		return
	}

	if err := tx.Commit().Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "product/update", "Failed to commit transaction")
		return
	}

	// Re-fetch the product with all associations for the response
	h.db.Preload("Categories").Preload("Tags").Preload("Images").Preload("Variants.Images").Preload("Options.Values").Preload("Specifications").First(&product, productID)

	response.GenerateSuccessResponse(c, "Product updated successfully", product)
}
