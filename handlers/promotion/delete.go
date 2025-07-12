package promotion

import (
	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/YasserCherfaoui/MarketProGo/utils/response"
	"github.com/gin-gonic/gin"
)

func (h *PromotionHandler) DeletePromotion(c *gin.Context) {
	promotionID := c.Param("id")

	if err := h.db.Delete(&models.Promotion{}, promotionID).Error; err != nil {
		response.GenerateInternalServerErrorResponse(c, "promotion/delete", err.Error())
		return
	}

	response.GenerateSuccessResponse(c, "Promotion deleted successfully", nil)
}
