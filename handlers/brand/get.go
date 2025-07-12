package brand

import (
	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

func (h *BrandHandler) GetBrand(c *gin.Context) {
	id := c.Param("id")
	var brand models.Brand
	if err := h.db.Preload("Parent").Preload("Children").Where("id = ?", id).First(&brand).Error; err != nil {
		response.GenerateNotFoundResponse(c, "brand/get", "Brand not found")
		return
	}
	response.GenerateSuccessResponse(c, "Brand fetched successfully", brand)
}
