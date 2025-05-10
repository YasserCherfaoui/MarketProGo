package product

import (
	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

func (h *ProductHandler) GetAllProducts(c *gin.Context) {
	products := []models.Product{}
	if err := h.db.Preload("Categories").Preload("Images").Find(&products).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "product/get_all", err.Error())
	}

	response.GenerateSuccessResponse(c, "Products fetched successfully", products)
}
