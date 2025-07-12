package promotion

import (
	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

func (h *PromotionHandler) GetAllPromotions(c *gin.Context) {
	var promotions []models.Promotion
	if err := h.db.Order("start_date DESC").Find(&promotions).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "promotion/get_all", "Failed to get all promotions")
		return
	}
	for i := range promotions {
		promotions[i].Image = h.appwriteService.GetFileURL(promotions[i].Image)
	}
	response.GenerateSuccessResponse(c, "Promotions fetched successfully", promotions)
}
