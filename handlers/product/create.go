package product

import (
	"encoding/json"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

// The following structs define the shape of the JSON expected in the 'product_data' form field.

type ProductData struct {
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	IsActive       bool                   `json:"is_active"`
	IsFeatured     bool                   `json:"is_featured"`
	IsVAT          bool                   `json:"is_vat"`
	BrandID        *uint                  `json:"brand_id"`
	CategoryIDs    []uint                 `json:"category_ids"`
	Tags           []string               `json:"tags"`
	Images         []ImageData            `json:"images"`
	Options        []OptionData           `json:"options"`
	Variants       []VariantData          `json:"variants"`
	Specifications []SpecificationRequest `json:"specifications"`
}

func (h *ProductHandler) CreateProduct(c *gin.Context) {
	// Step 1: Parse Multipart Form
	form, err := c.MultipartForm()
	if err != nil {
		response.GenerateBadRequestResponse(c, "product/create", "Invalid multipart form data: "+err.Error())
		return
	}

	// Step 2: Extract and Unmarshal JSON data from 'product_data' field
	productDataJSON := form.Value["product_data"]
	if len(productDataJSON) == 0 {
		response.GenerateBadRequestResponse(c, "product/create", "Missing 'product_data' field")
		return
	}
	var data ProductData
	if err := json.Unmarshal([]byte(productDataJSON[0]), &data); err != nil {
		response.GenerateBadRequestResponse(c, "product/create", "Invalid JSON in 'product_data' field: "+err.Error())
		return
	}

	// Step 3: Upload all files and map them by filename
	files := form.File["files"]
	uploadedFileIDs := make(map[string]string)
	for _, fileHeader := range files {
		fileID, err := h.appwriteService.UploadFile(fileHeader)
		if err != nil {
			response.GenerateInternalServerErrorResponse(c, "product/create", "Failed to upload image '"+fileHeader.Filename+"': "+err.Error())
			return
		}
		uploadedFileIDs[fileHeader.Filename] = fileID
	}

	// Step 4: Create product and associations in a transaction
	tx := h.db.Begin()
	if tx.Error != nil {
		response.GenerateInternalServerErrorResponse(c, "product/create", "Failed to start transaction")
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create Product
	product := models.Product{
		Name:        data.Name,
		Description: data.Description,
		IsActive:    data.IsActive,
		IsFeatured:  data.IsFeatured,
		IsVAT:       data.IsVAT,
		BrandID:     data.BrandID,
	}
	if err := tx.Create(&product).Error; err != nil {
		tx.Rollback()
		response.GenerateInternalServerErrorResponse(c, "product/create", "Failed to create product")
		return
	}

	// Associate Images with base product
	for _, imgData := range data.Images {
		fileID, ok := uploadedFileIDs[imgData.FileName]
		if !ok {
			tx.Rollback()
			response.GenerateBadRequestResponse(c, "product/create", "Image file '"+imgData.FileName+"' not found in upload")
			return
		}
		image := models.ProductImage{ProductID: &product.ID, URL: fileID, IsPrimary: imgData.IsPrimary, AltText: imgData.AltText}
		if err := tx.Create(&image).Error; err != nil {
			tx.Rollback()
			response.GenerateInternalServerErrorResponse(c, "product/create", "Failed to create product image")
			return
		}
	}

	// Associate Categories
	if len(data.CategoryIDs) > 0 {
		var categories []models.Category
		if err := tx.Find(&categories, data.CategoryIDs).Error; err != nil {
			tx.Rollback()
			response.GenerateInternalServerErrorResponse(c, "product/create", "Failed to find categories")
			return
		}
		if err := tx.Model(&product).Association("Categories").Replace(categories); err != nil {
			tx.Rollback()
			response.GenerateInternalServerErrorResponse(c, "product/create", "Failed to associate categories")
			return
		}
	}

	// Create and Associate Tags
	if len(data.Tags) > 0 {
		var tags []*models.Tag
		for _, tagName := range data.Tags {
			tag := models.Tag{Name: tagName}
			if err := tx.Where(models.Tag{Name: tagName}).FirstOrCreate(&tag).Error; err != nil {
				tx.Rollback()
				response.GenerateInternalServerErrorResponse(c, "product/create", "Failed to find or create tag")
				return
			}
			tags = append(tags, &tag)
		}
		if err := tx.Model(&product).Association("Tags").Replace(tags); err != nil {
			tx.Rollback()
			response.GenerateInternalServerErrorResponse(c, "product/create", "Failed to associate tags")
			return
		}
	}

	// Create Options and OptionValues
	optionValueMap := make(map[string]*models.ProductOptionValue)
	for _, optData := range data.Options {
		option := models.ProductOption{ProductID: product.ID, Name: optData.Name}
		if err := tx.Create(&option).Error; err != nil {
			tx.Rollback()
			response.GenerateInternalServerErrorResponse(c, "product/create", "Failed to create product option")
			return
		}
		for _, val := range optData.Values {
			optionValue := models.ProductOptionValue{ProductOptionID: option.ID, Value: val}
			if err := tx.Create(&optionValue).Error; err != nil {
				tx.Rollback()
				response.GenerateInternalServerErrorResponse(c, "product/create", "Failed to create product option value")
				return
			}
			optionValueMap[val] = &optionValue
		}
	}

	// Create Variants
	for _, varData := range data.Variants {
		variant := models.ProductVariant{
			ProductID:       product.ID,
			Name:            varData.Name,
			SKU:             varData.SKU,
			Barcode:         varData.Barcode,
			BasePrice:       varData.BasePrice,
			B2BPrice:        varData.B2BPrice,
			CostPrice:       varData.CostPrice,
			Weight:          varData.Weight,
			WeightUnit:      varData.WeightUnit,
			Dimensions:      &varData.Dimensions,
			IsActive:        varData.IsActive,
			MinQuantity:     varData.MinQuantity,
			QuantityInStock: varData.QuantityInStock,
		}
		if err := tx.Create(&variant).Error; err != nil {
			tx.Rollback()
			response.GenerateInternalServerErrorResponse(c, "product/create", "Failed to create product variant")
			return
		}

		// Create price tiers for this variant
		for _, tier := range varData.PriceTiers {
			priceTier := models.ProductVariantPriceTier{
				ProductVariantID: variant.ID,
				MinQuantity:      tier.MinQuantity,
				Price:            tier.Price,
			}
			if err := tx.Create(&priceTier).Error; err != nil {
				tx.Rollback()
				response.GenerateInternalServerErrorResponse(c, "product/create", "Failed to create price tier for variant")
				return
			}
		}

		// Associate Images with variant
		for _, imgData := range varData.Images {
			fileID, ok := uploadedFileIDs[imgData.FileName]
			if !ok {
				tx.Rollback()
				response.GenerateBadRequestResponse(c, "product/create", "Image file '"+imgData.FileName+"' for variant '"+variant.Name+"' not found in upload")
				return
			}
			image := models.ProductImage{ProductVariantID: &variant.ID, URL: fileID, IsPrimary: imgData.IsPrimary, AltText: imgData.AltText}
			if err := tx.Create(&image).Error; err != nil {
				tx.Rollback()
				response.GenerateInternalServerErrorResponse(c, "product/create", "Failed to create variant image")
				return
			}
		}

		// Associate OptionValues with variant
		var optionValuesToAssociate []*models.ProductOptionValue
		for _, valName := range varData.OptionValues {
			if ov, ok := optionValueMap[valName]; ok {
				optionValuesToAssociate = append(optionValuesToAssociate, ov)
			} else {
				tx.Rollback()
				response.GenerateBadRequestResponse(c, "product/create", "Invalid option value '"+valName+"' provided for variant")
				return
			}
		}
		if err := tx.Model(&variant).Association("OptionValues").Replace(optionValuesToAssociate); err != nil {
			tx.Rollback()
			response.GenerateInternalServerErrorResponse(c, "product/create", "Failed to associate option values to variant")
			return
		}
	}

	// Handle Specifications
	for _, specReq := range data.Specifications {
		spec := models.ProductSpecification{
			ProductID: product.ID,
			Name:      specReq.Name,
			Value:     specReq.Value,
			Unit:      specReq.Unit,
		}
		if err := tx.Create(&spec).Error; err != nil {
			tx.Rollback()
			response.GenerateInternalServerErrorResponse(c, "product/create", "Failed to create specification")
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "product/create", "Failed to commit transaction")
		return
	}

	// Preload all associations for the response
	if err := h.db.Preload("Brand").Preload("Categories").Preload("Tags").Preload("Images").Preload("Options.Values").Preload("Variants.Images").Preload("Variants.OptionValues").Preload("Specifications").First(&product, product.ID).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "product/create", "Failed to preload product data for response")
		return
	}

	response.GenerateSuccessResponse(c, "Product created successfully", product)
}
