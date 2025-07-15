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
	Name                   *string                   `json:"name"`
	Description            *string                   `json:"description"`
	IsActive               *bool                     `json:"is_active"`
	IsFeatured             *bool                     `json:"is_featured"`
	IsVAT                  *bool                     `json:"is_vat"`
	BrandID                *uint                     `json:"brand_id"`
	CategoryIDs            []uint                    `json:"category_ids"`
	Tags                   []string                  `json:"tags"`
	ImagesToAdd            []ImageData               `json:"images_to_add"`
	ImagesToUpdate         []ImageUpdateData         `json:"images_to_update"`
	ImagesToDelete         []uint                    `json:"images_to_delete"`
	SpecificationsToAdd    []SpecificationRequest    `json:"specifications_to_add"`
	SpecificationsToUpdate []SpecificationUpdateData `json:"specifications_to_update"`
	SpecificationsToDelete []uint                    `json:"specifications_to_delete"`
	OptionsToAdd           []OptionData              `json:"options_to_add"`
	OptionsToUpdate        []OptionUpdateData        `json:"options_to_update"`
	OptionsToDelete        []uint                    `json:"options_to_delete"`
	VariantsToAdd          []VariantData             `json:"variants_to_add"`
	VariantsToUpdate       []VariantUpdateData       `json:"variants_to_update"`
	VariantsToDelete       []uint                    `json:"variants_to_delete"`
	// Note: Image updates are handled via file upload and 'images_to_delete' form field
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

		// Handle Images to Update (metadata)
		for _, imgUpdate := range data.ImagesToUpdate {
			var img models.ProductImage
			if err := tx.First(&img, imgUpdate.ID).Error; err != nil {
				tx.Rollback()
				response.GenerateBadRequestResponse(c, "product/update", "Image with ID "+strconv.Itoa(int(imgUpdate.ID))+" not found.")
				return
			}
			if imgUpdate.IsPrimary != nil {
				img.IsPrimary = *imgUpdate.IsPrimary
			}
			if imgUpdate.AltText != nil {
				img.AltText = *imgUpdate.AltText
			}
			if err := tx.Save(&img).Error; err != nil {
				tx.Rollback()
				response.GenerateInternalServerErrorResponse(c, "product/update", "Failed to update image metadata")
				return
			}
		}

		// Handle Images to Delete (already handled above, but for completeness)
		for _, imgID := range data.ImagesToDelete {
			if err := tx.Delete(&models.ProductImage{}, imgID).Error; err != nil {
				tx.Rollback()
				response.GenerateInternalServerErrorResponse(c, "product/update", "Failed to delete image")
				return
			}
		}

		// --- Specifications CRUD ---
		// Add
		for _, spec := range data.SpecificationsToAdd {
			s := models.ProductSpecification{
				ProductID: product.ID,
				Name:      spec.Name,
				Value:     spec.Value,
				Unit:      spec.Unit,
			}
			if err := tx.Create(&s).Error; err != nil {
				tx.Rollback()
				response.GenerateInternalServerErrorResponse(c, "product/update", "Failed to add specification")
				return
			}
		}
		// Update
		for _, spec := range data.SpecificationsToUpdate {
			var s models.ProductSpecification
			if err := tx.First(&s, spec.ID).Error; err != nil {
				tx.Rollback()
				response.GenerateBadRequestResponse(c, "product/update", "Specification with ID "+strconv.Itoa(int(spec.ID))+" not found.")
				return
			}
			if spec.Name != nil {
				s.Name = *spec.Name
			}
			if spec.Value != nil {
				s.Value = *spec.Value
			}
			if spec.Unit != nil {
				s.Unit = *spec.Unit
			}
			if err := tx.Save(&s).Error; err != nil {
				tx.Rollback()
				response.GenerateInternalServerErrorResponse(c, "product/update", "Failed to update specification")
				return
			}
		}
		// Delete
		for _, specID := range data.SpecificationsToDelete {
			if err := tx.Delete(&models.ProductSpecification{}, specID).Error; err != nil {
				tx.Rollback()
				response.GenerateInternalServerErrorResponse(c, "product/update", "Failed to delete specification")
				return
			}
		}

		// --- Options CRUD ---
		// Add
		for _, opt := range data.OptionsToAdd {
			option := models.ProductOption{ProductID: product.ID, Name: opt.Name}
			if err := tx.Create(&option).Error; err != nil {
				tx.Rollback()
				response.GenerateInternalServerErrorResponse(c, "product/update", "Failed to add option")
				return
			}
			for _, val := range opt.Values {
				optionValue := models.ProductOptionValue{ProductOptionID: option.ID, Value: val}
				if err := tx.Create(&optionValue).Error; err != nil {
					tx.Rollback()
					response.GenerateInternalServerErrorResponse(c, "product/update", "Failed to add option value")
					return
				}
			}
		}
		// Update
		for _, opt := range data.OptionsToUpdate {
			var option models.ProductOption
			if err := tx.First(&option, opt.ID).Error; err != nil {
				tx.Rollback()
				response.GenerateBadRequestResponse(c, "product/update", "Option with ID "+strconv.Itoa(int(opt.ID))+" not found.")
				return
			}
			if opt.Name != nil {
				option.Name = *opt.Name
			}
			if err := tx.Save(&option).Error; err != nil {
				tx.Rollback()
				response.GenerateInternalServerErrorResponse(c, "product/update", "Failed to update option")
				return
			}
			if opt.Values != nil {
				// Remove all old values and add new ones
				tx.Where("product_option_id = ?", option.ID).Delete(&models.ProductOptionValue{})
				for _, val := range *opt.Values {
					optionValue := models.ProductOptionValue{ProductOptionID: option.ID, Value: val}
					if err := tx.Create(&optionValue).Error; err != nil {
						tx.Rollback()
						response.GenerateInternalServerErrorResponse(c, "product/update", "Failed to update option values")
						return
					}
				}
			}
		}
		// Delete
		for _, optID := range data.OptionsToDelete {
			tx.Where("product_option_id = ?", optID).Delete(&models.ProductOptionValue{})
			if err := tx.Delete(&models.ProductOption{}, optID).Error; err != nil {
				tx.Rollback()
				response.GenerateInternalServerErrorResponse(c, "product/update", "Failed to delete option")
				return
			}
		}

		// --- Variants CRUD ---
		// Add
		for _, varData := range data.VariantsToAdd {
			variant := models.ProductVariant{
				ProductID:   product.ID,
				Name:        varData.Name,
				SKU:         varData.SKU,
				Barcode:     varData.Barcode,
				BasePrice:   varData.BasePrice,
				B2BPrice:    varData.B2BPrice,
				CostPrice:   varData.CostPrice,
				Weight:      varData.Weight,
				WeightUnit:  varData.WeightUnit,
				Dimensions:  &varData.Dimensions,
				IsActive:    varData.IsActive,
				MinQuantity: varData.MinQuantity,
			}
			if err := tx.Create(&variant).Error; err != nil {
				tx.Rollback()
				response.GenerateInternalServerErrorResponse(c, "product/update", "Failed to add variant")
				return
			}
			// Add price tiers for this variant
			for _, tier := range varData.PriceTiers {
				priceTier := models.ProductVariantPriceTier{
					ProductVariantID: variant.ID,
					MinQuantity:      tier.MinQuantity,
					Price:            tier.Price,
				}
				if err := tx.Create(&priceTier).Error; err != nil {
					tx.Rollback()
					response.GenerateInternalServerErrorResponse(c, "product/update", "Failed to add price tier for variant")
					return
				}
			}
			// Add images for variant
			for _, imgData := range varData.Images {
				fileID, ok := uploadedFileIDs[imgData.FileName]
				if !ok {
					tx.Rollback()
					response.GenerateBadRequestResponse(c, "product/update", "Image file '"+imgData.FileName+"' for variant '"+variant.Name+"' not found in upload")
					return
				}
				image := models.ProductImage{
					ProductVariantID: &variant.ID,
					URL:              fileID,
					IsPrimary:        imgData.IsPrimary,
					AltText:          imgData.AltText,
				}
				if err := tx.Create(&image).Error; err != nil {
					tx.Rollback()
					response.GenerateInternalServerErrorResponse(c, "product/update", "Failed to add variant image")
					return
				}
			}
			// Associate option values
			if len(varData.OptionValues) > 0 {
				var optionValues []*models.ProductOptionValue
				for _, val := range varData.OptionValues {
					var ov models.ProductOptionValue
					if err := tx.Where("value = ?", val).First(&ov).Error; err == nil {
						optionValues = append(optionValues, &ov)
					}
				}
				if err := tx.Model(&variant).Association("OptionValues").Replace(optionValues); err != nil {
					tx.Rollback()
					response.GenerateInternalServerErrorResponse(c, "product/update", "Failed to associate option values to variant")
					return
				}
			}
		}
		// Update and delete for variants is already implemented above

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
		if data.IsVAT != nil {
			product.IsVAT = *data.IsVAT
		}
		if data.BrandID != nil {
			product.BrandID = data.BrandID
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
			if varUpdateData.MinQuantity != nil {
				variant.MinQuantity = *varUpdateData.MinQuantity
			}
			if err := tx.Save(&variant).Error; err != nil {
				tx.Rollback()
				response.GenerateInternalServerErrorResponse(c, "product/update", "Failed to update variant")
				return
			}
			// Update price tiers for this variant
			if varUpdateData.PriceTiers != nil {
				tx.Where("product_variant_id = ?", variant.ID).Delete(&models.ProductVariantPriceTier{})
				for _, tier := range *varUpdateData.PriceTiers {
					priceTier := models.ProductVariantPriceTier{
						ProductVariantID: variant.ID,
						MinQuantity:      tier.MinQuantity,
						Price:            tier.Price,
					}
					if err := tx.Create(&priceTier).Error; err != nil {
						tx.Rollback()
						response.GenerateInternalServerErrorResponse(c, "product/update", "Failed to update price tiers for variant")
						return
					}
				}
			}
			// --- Option Values CRUD ---
			if varUpdateData.OptionValuesToAdd != nil {
				var optionValues []*models.ProductOptionValue
				for _, val := range *varUpdateData.OptionValuesToAdd {
					var ov models.ProductOptionValue
					if err := tx.Where("value = ?", val).First(&ov).Error; err == nil {
						optionValues = append(optionValues, &ov)
					}
				}
				if err := tx.Model(&variant).Association("OptionValues").Append(optionValues); err != nil {
					tx.Rollback()
					response.GenerateInternalServerErrorResponse(c, "product/update", "Failed to add option values to variant")
					return
				}
			}
			if varUpdateData.OptionValuesToRemove != nil {
				var optionValues []*models.ProductOptionValue
				for _, val := range *varUpdateData.OptionValuesToRemove {
					var ov models.ProductOptionValue
					if err := tx.Where("value = ?", val).First(&ov).Error; err == nil {
						optionValues = append(optionValues, &ov)
					}
				}
				if err := tx.Model(&variant).Association("OptionValues").Delete(optionValues); err != nil {
					tx.Rollback()
					response.GenerateInternalServerErrorResponse(c, "product/update", "Failed to remove option values from variant")
					return
				}
			}
			// --- Images CRUD ---
			// Add new images
			for _, imgData := range varUpdateData.ImagesToAdd {
				fileID, ok := uploadedFileIDs[imgData.FileName]
				if !ok {
					tx.Rollback()
					response.GenerateBadRequestResponse(c, "product/update", "Image file '"+imgData.FileName+"' for variant '"+variant.Name+"' not found in upload")
					return
				}
				image := models.ProductImage{
					ProductVariantID: &variant.ID,
					URL:              fileID,
					IsPrimary:        imgData.IsPrimary,
					AltText:          imgData.AltText,
				}
				if err := tx.Create(&image).Error; err != nil {
					tx.Rollback()
					response.GenerateInternalServerErrorResponse(c, "product/update", "Failed to add variant image")
					return
				}
			}
			// Update image metadata
			for _, imgUpdate := range varUpdateData.ImagesToUpdate {
				var img models.ProductImage
				if err := tx.First(&img, imgUpdate.ID).Error; err != nil {
					tx.Rollback()
					response.GenerateBadRequestResponse(c, "product/update", "Image with ID "+strconv.Itoa(int(imgUpdate.ID))+" not found.")
					return
				}
				if imgUpdate.IsPrimary != nil {
					img.IsPrimary = *imgUpdate.IsPrimary
				}
				if imgUpdate.AltText != nil {
					img.AltText = *imgUpdate.AltText
				}
				if err := tx.Save(&img).Error; err != nil {
					tx.Rollback()
					response.GenerateInternalServerErrorResponse(c, "product/update", "Failed to update image metadata")
					return
				}
			}
			// Delete images
			for _, imgID := range varUpdateData.ImagesToDelete {
				if err := tx.Delete(&models.ProductImage{}, imgID).Error; err != nil {
					tx.Rollback()
					response.GenerateInternalServerErrorResponse(c, "product/update", "Failed to delete variant image")
					return
				}
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
	h.db.Preload("Brand").Preload("Categories").Preload("Tags").Preload("Images").Preload("Variants.Images").Preload("Options.Values").Preload("Specifications").First(&product, productID)

	response.GenerateSuccessResponse(c, "Product updated successfully", product)
}
