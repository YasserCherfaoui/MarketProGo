package product

import (
	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

func (h *ProductHandler) GetProduct(c *gin.Context) {
	productID := c.Param("id")
	includeInactive := c.Query("include_inactive") == "true"

	var product models.Product
	query := h.db.
		Preload("Brand").
		Preload("Categories").
		Preload("Tags").
		Preload("Images").
		Preload("Options.Values").
		Preload("Variants.Images").
		Preload("Variants.OptionValues").
		Preload("Variants.InventoryItems").
		Preload("Variants.InventoryItems.Warehouse").
		Preload("Variants.PriceTiers").
		Preload("Specifications")

	// Only fetch active products by default
	if !includeInactive {
		query = query.Where("is_active = ?", true)
	}

	err := query.First(&product, "id = ?", productID).Error

	if err != nil {
		response.GenerateNotFoundResponse(c, "product/get", "Product not found")
		return
	}

	// Add Appwrite URLs to product and brand images
	if product.Brand != nil {
		product.Brand.Image = h.appwriteService.GetFileURL(product.Brand.Image)
	}
	for i := range product.Images {
		product.Images[i].URL = h.appwriteService.GetFileURL(product.Images[i].URL)
	}
	for i := range product.Variants {
		for j := range product.Variants[i].Images {
			product.Variants[i].Images[j].URL = h.appwriteService.GetFileURL(product.Variants[i].Images[j].URL)
		}
	}

	// Add review data to product
	err = h.reviewService.AddReviewDataToProduct(&product)
	if err != nil {
		// Log error but don't fail the request
		// TODO: Add proper logging
	}

	response.GenerateSuccessResponse(c, "product/get", product)
}
