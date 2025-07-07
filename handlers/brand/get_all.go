package brand

import (
	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

func (h *BrandHandler) GetAllBrands(c *gin.Context) {
	var brands []models.Brand
	if err := h.db.Order("name ASC").Find(&brands).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "brand/get_all", "Failed to get all brands")
		return
	}
	// Add Appwrite URLs to brand images
	for i := range brands {
		brands[i].Image = h.appwriteService.GetFileURL(brands[i].Image)
	}
	response.GenerateSuccessResponse(c, "Brands fetched successfully", brands)
}
