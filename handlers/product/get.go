package product

import (
	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

func (h *ProductHandler) GetProduct(c *gin.Context) {
	productID := c.Param("id")

	product := models.Product{}
	err := h.db.Where("id = ?", productID).Preload("Categories").Preload("Images").First(&product).Error
	if err != nil {
		response.GenerateNotFoundResponse(c, "product/get", "Product not found")
		return
	}

	response.GenerateSuccessResponse(c, "product/get", product)
}
