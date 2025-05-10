package product

import (
	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	productID := c.Param("id")

	if err := h.db.Delete(&models.Product{}, productID).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "product/delete", err.Error())
		return
	}

	response.GenerateSuccessResponse(c, "Product deleted successfully", nil)
}
