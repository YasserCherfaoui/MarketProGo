package promotion

import (
	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

func (h *PromotionHandler) GetPromotion(c *gin.Context) {
	id := c.Param("id")
	var promotion models.Promotion
	if err := h.db.Where("id = ?", id).First(&promotion).Error; err != nil {
		response.GenerateNotFoundResponse(c, "promotion/get", "Promotion not found")
		return
	}
	promotion.Image = h.appwriteService.GetFileURL(promotion.Image)
	response.GenerateSuccessResponse(c, "Promotion fetched successfully", promotion)
}
