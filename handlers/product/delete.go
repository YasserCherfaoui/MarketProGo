package product

import (
	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	productID := c.Param("id")

	tx := h.db.Begin()
	if tx.Error != nil {
		response.GenerateInternalServerErrorResponse(c, "product/delete", "Failed to start transaction")
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var product models.Product
	if err := tx.First(&product, productID).Error; err != nil {
		tx.Rollback()
		response.GenerateNotFoundResponse(c, "product/delete", "Product not found")
		return
	}

	// Delete associations. This ensures no orphaned records are left.

	// Find and delete all variants and their specific associations
	var variants []models.ProductVariant
	tx.Where("product_id = ?", product.ID).Find(&variants)
	for _, v := range variants {
		// Delete images associated with this variant
		tx.Where("product_variant_id = ?", v.ID).Delete(&models.ProductImage{})
		// Delete join table records for variant-option values
		tx.Exec("DELETE FROM variant_option_values WHERE product_variant_id = ?", v.ID)
	}
	// Delete all variants for the product
	tx.Where("product_id = ?", product.ID).Delete(&models.ProductVariant{})

	// Delete base product images
	tx.Where("product_id = ?", product.ID).Delete(&models.ProductImage{})

	// Find and delete all options and their values
	var options []models.ProductOption
	tx.Where("product_id = ?", product.ID).Find(&options)
	for _, o := range options {
		tx.Where("product_option_id = ?", o.ID).Delete(&models.ProductOptionValue{})
	}
	tx.Where("product_id = ?", product.ID).Delete(&models.ProductOption{})

	// Delete Specifications
	tx.Where("product_id = ?", product.ID).Delete(&models.ProductSpecification{})

	// Clear many-to-many associations for categories and tags
	if err := tx.Model(&product).Association("Categories").Clear(); err != nil {
		tx.Rollback()
		response.GenerateInternalServerErrorResponse(c, "product/delete", "Failed to clear category associations")
		return
	}
	if err := tx.Model(&product).Association("Tags").Clear(); err != nil {
		tx.Rollback()
		response.GenerateInternalServerErrorResponse(c, "product/delete", "Failed to clear tag associations")
		return
	}

	// Finally, delete the product itself
	if err := tx.Delete(&models.Product{}, productID).Error; err != nil {
		tx.Rollback()
		response.GenerateInternalServerErrorResponse(c, "product/delete", "Failed to delete product")
		return
	}

	if err := tx.Commit().Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "product/delete", "Failed to commit transaction")
		return
	}

	response.GenerateSuccessResponse(c, "Product deleted successfully", nil)
}
